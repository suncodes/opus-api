package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"opus-api/internal/converter"
	"opus-api/internal/logger"
	"opus-api/internal/model"
	"opus-api/internal/stream"
	"opus-api/internal/tokenizer"
	"opus-api/internal/types"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleMessages handles POST /v1/messages
func HandleMessages(c *gin.Context) {
	// Generate request ID
	requestID := uuid.New().String()[:8]

	// Rotate logs before creating new folder
	if types.DebugMode {
		logger.RotateLogs()
	}

	// Parse Claude request
	var claudeReq types.ClaudeRequest
	if err := c.ShouldBindJSON(&claudeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 验证模型是否支持
	if claudeReq.Model == "" {
		claudeReq.Model = types.DefaultModel
	}
	if !types.IsModelSupported(claudeReq.Model) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Model '%s' is not supported. Supported models: %v", claudeReq.Model, types.SupportedModels),
		})
		return
	}

	// Log Point 1: Claude request
	var logFolder string
	if types.DebugMode {
		logFolder, _ = logger.CreateLogFolder(requestID)
		logger.WriteJSONLog(logFolder, "1_claude_request.json", claudeReq)
	}

	// Convert to Morph format
	morphReq := converter.ClaudeToMorph(claudeReq)

	// Log Point 2: Morph request
	if types.DebugMode && logFolder != "" {
		logger.WriteJSONLog(logFolder, "2_morph_request.json", morphReq)
	}

	// Send request to MorphLLM API
	morphReqJSON, err := json.Marshal(morphReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	req, err := http.NewRequest("POST", types.MorphAPIURL, bytes.NewReader(morphReqJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set headers
	headers := make(map[string]string)
	for key, value := range types.MorphHeaders {
		headers[key] = value
	}

	// 如果启用了 Cookie 轮询器，使用轮询的 Cookie
	if types.CookieRotatorInstance != nil {
		cookieInterface, err := types.CookieRotatorInstance.NextCookie()
		if err == nil && cookieInterface != nil {
			// 类型断言为 *model.MorphCookie
			if cookie, ok := cookieInterface.(*model.MorphCookie); ok {
				headers["cookie"] = cookie.APIKey
				log.Printf("[INFO] Using rotated cookie (ID: %d, Priority: %d)", cookie.ID, cookie.Priority)
			} else {
				log.Printf("[WARN] Cookie type assertion failed, using default")
			}
		} else {
			log.Printf("[WARN] Failed to get rotated cookie: %v, using default", err)
		}
	}

	// 应用所有请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Log Point 3: Upstream request with headers
	if types.DebugMode && logFolder != "" {
		var reqLog strings.Builder
		reqLog.WriteString(fmt.Sprintf("%s %s\n", req.Method, req.URL))
		for k, v := range req.Header {
			reqLog.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
		reqLog.WriteString("\n")
		reqLog.Write(morphReqJSON)
		logger.WriteTextLog(logFolder, "3_upstream_request.txt", reqLog.String())
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to upstream API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		if types.DebugMode && logFolder != "" {
			logger.WriteTextLog(logFolder, "error.txt", fmt.Sprintf("Error: %d %s\n%s", resp.StatusCode, resp.Status, string(bodyBytes)))
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to connect to upstream API",
			"status": resp.StatusCode,
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Create response writer that captures output for logging
	var clientResponseWriter io.Writer = io.Discard
	if types.DebugMode && logFolder != "" {
		logger.WriteTextLog(logFolder, "5_client_response.txt", "")
		clientResponseWriter = &logWriter{logFolder: logFolder, fileName: "5_client_response.txt"}
	}
	onChunk := func(chunk string) {
		if types.DebugMode {
			clientResponseWriter.Write([]byte(chunk))
		}
	}

	// Calculate input tokens from request
	inputTokens := calculateInputTokens(claudeReq)

	// Create a pipe for streaming
	pr, pw := io.Pipe()

	// Start goroutine to transform stream
	go func() {
		defer pw.Close()

		// Log Point 4: Upstream response
		var morphResponseWriter io.Writer = io.Discard
		if types.DebugMode && logFolder != "" {
			logger.WriteTextLog(logFolder, "4_upstream_response.txt", "")
			morphResponseWriter = &logWriter{logFolder: logFolder, fileName: "4_upstream_response.txt"}
		}

		// Tee the response body
		teeReader := io.TeeReader(resp.Body, morphResponseWriter)

		// Transform stream
		if err := stream.TransformMorphToClaudeStream(teeReader, claudeReq.Model, inputTokens, pw, onChunk); err != nil {
			log.Printf("[ERROR] Stream transformation error: %v", err)
		}
	}()

	// Stream response to client
	c.Stream(func(w io.Writer) bool {
		buf := make([]byte, 4096)
		n, err := pr.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		return err == nil
	})
}

// logWriter writes to log file
type logWriter struct {
	logFolder string
	fileName  string
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	if types.DebugMode && w.logFolder != "" {
		logger.AppendLog(w.logFolder, w.fileName, string(p))
	}
	return len(p), nil
}

// calculateInputTokens calculates the total input tokens from a Claude request
func calculateInputTokens(req types.ClaudeRequest) int {
	var totalText strings.Builder

	// Add system prompt
	if req.System != nil {
		if sysStr, ok := req.System.(string); ok {
			totalText.WriteString(sysStr)
		} else if sysList, ok := req.System.([]interface{}); ok {
			for _, item := range sysList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						totalText.WriteString(text)
					}
				}
			}
		}
	}

	// Add messages content
	for _, msg := range req.Messages {
		if content, ok := msg.Content.(string); ok {
			totalText.WriteString(content)
		} else if contentBlocks, ok := msg.Content.([]types.ClaudeContentBlock); ok {
			for _, block := range contentBlocks {
				if textBlock, ok := block.(types.ClaudeContentBlockText); ok {
					totalText.WriteString(textBlock.Text)
				} else if toolResult, ok := block.(types.ClaudeContentBlockToolResult); ok {
					if resultStr, ok := toolResult.Content.(string); ok {
						totalText.WriteString(resultStr)
					}
				}
			}
		}
	}

	// Add tools definitions
	for _, tool := range req.Tools {
		totalText.WriteString(tool.Name)
		totalText.WriteString(tool.Description)
		if tool.InputSchema != nil {
			schemaBytes, _ := json.Marshal(tool.InputSchema)
			totalText.Write(schemaBytes)
		}
	}

	return tokenizer.CountTokens(totalText.String())
}