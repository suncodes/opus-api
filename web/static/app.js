// API 基础路径
const API_BASE = window.location.origin;

// 存储认证 token
let authToken = localStorage.getItem('auth_token');

// 页面加载时检查认证状态
document.addEventListener('DOMContentLoaded', function() {
    const currentPage = window.location.pathname;
    console.log('[DEBUG] Current page:', currentPage);
    console.log('[DEBUG] Auth token:', authToken ? 'exists' : 'none');
    
    if (currentPage.includes('dashboard')) {
        console.log('[DEBUG] Loading dashboard...');
        if (!authToken) {
            window.location.href = '/static/index.html';
            return;
        }
        loadDashboard();
    } else {
        // 登录页面（包括 /static/index.html, /, 或其他）
        console.log('[DEBUG] Loading login page...');
        if (authToken) {
            console.log('[DEBUG] Already logged in, redirecting to dashboard...');
            window.location.href = '/dashboard';
            return;
        }
        setupLoginForm();
    }
});

// ========== 认证功能 ==========

// 设置登录表单
function setupLoginForm() {
    console.log('[DEBUG] setupLoginForm called');
    const loginForm = document.getElementById('loginForm');
    console.log('[DEBUG] loginForm element:', loginForm);
    if (loginForm) {
        console.log('[DEBUG] Adding submit event listener to loginForm');
        loginForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            console.log('[DEBUG] Form submitted!');
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('loginError');
            
            console.log('[DEBUG] Attempting login with username:', username);
            
            try {
                console.log('[DEBUG] Sending request to:', `${API_BASE}/api/auth/login`);
                const response = await fetch(`${API_BASE}/api/auth/login`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ username, password })
                });
                
                const data = await response.json();
                
                if (response.ok) {
                    authToken = data.token;
                    localStorage.setItem('auth_token', authToken);
                    window.location.href = '/dashboard';
                } else {
                    errorDiv.textContent = data.error || '登录失败';
                    errorDiv.style.display = 'block';
                }
            } catch (error) {
                errorDiv.textContent = '网络错误，请稍后重试';
                errorDiv.style.display = 'block';
            }
        });
    }
}

// 登出
function logout() {
    if (confirm('确定要退出吗？')) {
        localStorage.removeItem('auth_token');
        authToken = null;
        window.location.href = '/';
    }
}

// API 请求封装
async function apiRequest(url, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    if (authToken) {
        headers['Authorization'] = `Bearer ${authToken}`;
    }
    
    const response = await fetch(`${API_BASE}${url}`, {
        ...options,
        headers
    });
    
    // 如果 401，跳转到登录页
    if (response.status === 401) {
        localStorage.removeItem('auth_token');
        authToken = null;
        window.location.href = '/';
        return;
    }
    
    return response;
}

// ========== Dashboard 功能 ==========

// 加载 Dashboard
async function loadDashboard() {
    await loadUserInfo();
    await loadStats();
    await loadCookies();
}

// 加载用户信息
async function loadUserInfo() {
    try {
        const response = await apiRequest('/api/auth/me');
        if (response.ok) {
            const user = await response.json();
            document.getElementById('currentUser').textContent = user.username;
        }
    } catch (error) {
        console.error('加载用户信息失败:', error);
    }
}

// 加载统计信息
async function loadStats() {
    try {
        const response = await apiRequest('/api/cookies/stats');
        if (response.ok) {
            const stats = await response.json();
            document.getElementById('totalCount').textContent = stats.total_count;
            document.getElementById('validCount').textContent = stats.valid_count;
            document.getElementById('invalidCount').textContent = stats.invalid_count;
            document.getElementById('totalUsage').textContent = stats.total_usage.toLocaleString();
        }
    } catch (error) {
        console.error('加载统计信息失败:', error);
    }
}

// 加载 Cookie 列表
async function loadCookies() {
    try {
        const response = await apiRequest('/api/cookies');
        if (response.ok) {
            const data = await response.json();
            // API 返回的是数组而不是对象
            renderCookieTable(Array.isArray(data) ? data : (data.cookies || []));
        }
    } catch (error) {
        console.error('加载 Cookie 列表失败:', error);
    }
}

// 渲染 Cookie 表格
function renderCookieTable(cookies) {
    const tbody = document.getElementById('cookieTableBody');
    const emptyState = document.getElementById('emptyState');
    
    if (cookies.length === 0) {
        tbody.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }
    
    emptyState.style.display = 'none';
    
    tbody.innerHTML = cookies.map((cookie, index) => `
        <tr>
            <td>${index + 1}</td>
            <td>${escapeHtml(cookie.name)}</td>
            <td>
                <span class="status-badge ${cookie.is_valid ? 'status-valid' : 'status-invalid'}">
                    ${cookie.is_valid ? '✅ 有效' : '❌ 无效'}
                </span>
            </td>
            <td>${(cookie.usage_count || 0).toLocaleString()}</td>
            <td>${cookie.priority || 0}</td>
            <td>${formatTime(cookie.last_validated)}</td>
            <td>
                <div class="action-buttons">
                    <button class="btn btn-secondary btn-sm" onclick="validateCookie(${cookie.id})">🔄</button>
                    <button class="btn btn-secondary btn-sm" onclick="editCookie(${cookie.id})">⚙️</button>
                    <button class="btn btn-danger btn-sm" onclick="deleteCookie(${cookie.id})">🗑️</button>
                </div>
            </td>
        </tr>
    `).join('');
}

// 刷新 Cookie 列表
function refreshCookies() {
    loadCookies();
    loadStats();
}

// ========== Cookie 操作 ==========

// 显示添加弹窗
function showAddModal() {
    document.getElementById('addModal').style.display = 'flex';
    document.getElementById('addCookieForm').reset();
}

// 关闭添加弹窗
function closeAddModal() {
    document.getElementById('addModal').style.display = 'none';
}

// 添加 Cookie
document.getElementById('addCookieForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        api_key: formData.get('api_key'),
        session_key: formData.get('session_key') || '',
        priority: parseInt(formData.get('priority') || '0')
    };
    
    try {
        const response = await apiRequest('/api/cookies', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        if (response.ok) {
            showToast('Cookie 添加成功', 'success');
            closeAddModal();
            refreshCookies();
        } else {
            const error = await response.json();
            showToast(error.error || '添加失败', 'error');
        }
    } catch (error) {
        showToast('网络错误', 'error');
    }
});

// 显示编辑弹窗
async function editCookie(id) {
    try {
        const response = await apiRequest(`/api/cookies/${id}`);
        if (response.ok) {
            const cookie = await response.json();
            document.getElementById('editCookieId').value = cookie.id;
            document.getElementById('editCookieName').value = cookie.name;
            document.getElementById('editApiKey').value = cookie.api_key || '';
            document.getElementById('editSessionKey').value = cookie.session_key || '';
            document.getElementById('editCookiePriority').value = cookie.priority || 0;
            document.getElementById('editModal').style.display = 'flex';
        }
    } catch (error) {
        showToast('加载失败', 'error');
    }
}

// 关闭编辑弹窗
function closeEditModal() {
    document.getElementById('editModal').style.display = 'none';
}

// 更新 Cookie
document.getElementById('editCookieForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const id = document.getElementById('editCookieId').value;
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        api_key: formData.get('api_key'),
        session_key: formData.get('session_key') || '',
        priority: parseInt(formData.get('priority') || '0')
    };
    
    try {
        const response = await apiRequest(`/api/cookies/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
        
        if (response.ok) {
            showToast('Cookie 更新成功', 'success');
            closeEditModal();
            refreshCookies();
        } else {
            const error = await response.json();
            showToast(error.error || '更新失败', 'error');
        }
    } catch (error) {
        showToast('网络错误', 'error');
    }
});

// 删除 Cookie
async function deleteCookie(id) {
    if (!confirm('确定要删除这个 Cookie 吗？')) {
        return;
    }
    
    try {
        const response = await apiRequest(`/api/cookies/${id}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showToast('Cookie 删除成功', 'success');
            refreshCookies();
        } else {
            const error = await response.json();
            showToast(error.error || '删除失败', 'error');
        }
    } catch (error) {
        showToast('网络错误', 'error');
    }
}

// 验证单个 Cookie
async function validateCookie(id) {
    try {
        showToast('正在验证...', 'success');
        const response = await apiRequest(`/api/cookies/${id}/validate`, {
            method: 'POST'
        });
        
        if (response.ok) {
            const result = await response.json();
            showToast(result.is_valid ? '✅ Cookie 有效' : '❌ Cookie 无效', result.is_valid ? 'success' : 'error');
            refreshCookies();
        } else {
            const error = await response.json();
            showToast(error.error || '验证失败', 'error');
        }
    } catch (error) {
        showToast('网络错误', 'error');
    }
}

// 批量验证所有 Cookie
async function validateAll() {
    if (!confirm('确定要验证所有 Cookie 吗？这可能需要一些时间。')) {
        return;
    }
    
    try {
        showToast('正在批量验证...', 'success');
        const response = await apiRequest('/api/cookies/validate/all', {
            method: 'POST'
        });
        
        if (response.ok) {
            const result = await response.json();
            showToast(`验证完成：${result.valid_count} 个有效，${result.invalid_count} 个无效`, 'success');
            refreshCookies();
        } else {
            const error = await response.json();
            showToast(error.error || '验证失败', 'error');
        }
    } catch (error) {
        showToast('网络错误', 'error');
    }
}

// ========== 工具函数 ==========

// 显示 Toast 通知
function showToast(message, type = 'success') {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.className = `toast ${type}`;
    toast.style.display = 'block';
    
    setTimeout(() => {
        toast.style.display = 'none';
    }, 3000);
}

// 格式化时间
function formatTime(timeStr) {
    if (!timeStr) return '-';
    const date = new Date(timeStr);
    const now = new Date();
    const diff = Math.floor((now - date) / 1000);
    
    if (diff < 60) return '刚刚';
    if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
    if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
    return `${Math.floor(diff / 86400)} 天前`;
}

// HTML 转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}