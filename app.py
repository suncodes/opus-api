#!/usr/bin/env python3
"""
Hugging Face Spaces 启动脚本
用于启动 Go 编写的 Opus API 服务
"""
import os
import subprocess
import sys
import signal

def signal_handler(sig, frame):
    """处理终止信号"""
    print("\n正在关闭服务...")
    sys.exit(0)

def main():
    """启动 Go 服务"""
    # 注册信号处理器
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    print("=" * 50)
    print("启动 Opus API 服务")
    print("=" * 50)
    
    # 确保服务器可执行文件存在
    if not os.path.exists("./server"):
        print("错误: 找不到服务器可执行文件 './server'")
        sys.exit(1)
    
    # 设置端口（Hugging Face Spaces 使用 7860）
    port = os.getenv("PORT", "7860")
    print(f"端口: {port}")
    print(f"日志目录: /app/logs")
    print("=" * 50)
    
    try:
        # 启动 Go 服务
        process = subprocess.Popen(
            ["./server"],
            stdout=sys.stdout,
            stderr=sys.stderr
        )
        
        # 等待进程结束
        process.wait()
        
    except Exception as e:
        print(f"启动服务时出错: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()