// 动态获取配置
let API_BASE = 'http://localhost:8083/api';
let BASE_URL = 'http://localhost:8083';

// 时区处理工具函数
function convertToBeijingTime(utcTimeString) {
    if (!utcTimeString) return null;
    const date = new Date(utcTimeString);
    // 转换为北京时间（UTC+8）
    const beijingTime = new Date(date.getTime() + 8 * 60 * 60 * 1000);
    return beijingTime.toISOString().slice(0, 16);
}

function convertFromBeijingTime(localTimeString) {
    if (!localTimeString) return null;
    // 将本地时间字符串视为北京时间，转换为UTC
    const beijingDate = new Date(localTimeString);
    const utcDate = new Date(beijingDate.getTime() - 8 * 60 * 60 * 1000);
    return utcDate.toISOString();
}

function formatBeijingTime(utcTimeString) {
    if (!utcTimeString) return '无限制';
    const date = new Date(utcTimeString);
    // 检查日期是否有效
    if (isNaN(date.getTime())) return '无效时间';
    // 转换为北京时间显示
    return date.toLocaleString('zh-CN', {
        timeZone: 'Asia/Shanghai',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

// 专门用于时间范围显示的格式化函数
function formatTimeRange(startTime, endTime) {
    const start = formatBeijingTime(startTime);
    const end = formatBeijingTime(endTime);
    
    // 如果开始时间和结束时间都是无限制
    if (start === '无限制' && end === '无限制') {
        return '永久有效';
    }
    
    return `${start} - ${end}`;
}

// 清空时间字段
function clearTimeField(fieldId) {
    document.getElementById(fieldId).value = '';
}

// 初始化配置
async function initConfig() {
    try {
        const response = await fetch('/api/config');
        if (response.ok) {
            const config = await response.json();
            BASE_URL = config.base_url || 'http://localhost:8083';
            API_BASE = `${BASE_URL}/api`;
            console.log('Config loaded:', { BASE_URL, API_BASE });
        }
    } catch (error) {
        console.warn('Failed to load config, using defaults:', error);
    }
}
let currentUser = null;
let authToken = null;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', async function() {
    // 初始化配置
    await initConfig();
    
    // 检查是否已登录
    const token = localStorage.getItem('authToken');
    console.log('Page loaded, checking token:', token ? token.substring(0, 20) + '...' : 'No token found');
    
    if (token) {
        authToken = token;
        console.log('Token restored, showing main app');
        showMainApp();
        loadDashboardData();
    } else {
        console.log('No token found, showing login page');
    }
    
    // 绑定事件
    bindEvents();
});

// 绑定所有事件
function bindEvents() {
    // 登录表单
    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    
    // 注册表单
    document.getElementById('registerForm').addEventListener('submit', handleRegister);
    
    // 显示注册页面
    document.getElementById('showRegister').addEventListener('click', function(e) {
        e.preventDefault();
        document.getElementById('loginPage').style.display = 'none';
        document.getElementById('registerPage').classList.remove('hidden');
        document.getElementById('registerPage').style.display = 'flex';
    });
    
    // 显示登录页面
    document.getElementById('showLogin').addEventListener('click', function(e) {
        e.preventDefault();
        document.getElementById('registerPage').style.display = 'none';
        document.getElementById('registerPage').classList.add('hidden');
        document.getElementById('loginPage').style.display = 'flex';
    });
    
    // 侧边栏导航
    document.querySelectorAll('.sidebar .nav-link').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            if (this.getAttribute('onclick')) return; // 跳过退出登录链接
            
            // 移除所有活动状态
            document.querySelectorAll('.sidebar .nav-link').forEach(l => l.classList.remove('active'));
            document.querySelectorAll('.content-section').forEach(s => s.classList.remove('active'));
            
            // 添加当前活动状态
            this.classList.add('active');
            const section = this.getAttribute('data-section');
            document.getElementById(section).classList.add('active');
            
            // 加载对应数据
            loadSectionData(section);
        });
    });
    
    // 创建活码表单
    document.getElementById('createActiveQRForm').addEventListener('submit', function(e) {
        e.preventDefault();
        createActiveQR();
    });
    
    // 创建静态码表单
    document.getElementById('createStaticQRForm').addEventListener('submit', function(e) {
        e.preventDefault();
        createStaticQR();
    });
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    
    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            authToken = data.data.token;
            currentUser = data.data.user;
            localStorage.setItem('authToken', authToken);
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            
            showMainApp();
            loadDashboardData();
            showAlert('登录成功！', 'success');
        } else {
            showAlert(data.error || data.message || '登录失败', 'danger');
        }
    } catch (error) {
        console.error('Login error:', error);
        showAlert('网络错误，请稍后重试', 'danger');
    }
}

// 处理注册
async function handleRegister(e) {
    e.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const password = document.getElementById('regPassword').value;
    const confirmPassword = document.getElementById('regConfirmPassword').value;
    
    if (password !== confirmPassword) {
        showAlert('两次输入的密码不一致', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/auth/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showAlert('注册成功！请登录', 'success');
            document.getElementById('registerPage').style.display = 'none';
            document.getElementById('registerPage').classList.add('hidden');
            document.getElementById('loginPage').style.display = 'flex';
        } else {
            showAlert(data.error || data.message || '注册失败', 'danger');
        }
    } catch (error) {
        console.error('Register error:', error);
        showAlert('网络错误，请稍后重试', 'danger');
    }
}

// 显示主应用
function showMainApp() {
    document.getElementById('loginPage').style.display = 'none';
    document.getElementById('registerPage').style.display = 'none';
    document.getElementById('registerPage').classList.add('hidden');
    document.getElementById('mainApp').classList.remove('main-app-hidden');
    document.getElementById('mainApp').style.display = 'block';
    
    // 显示当前用户
    const user = localStorage.getItem('currentUser');
    if (user) {
        const userData = JSON.parse(user);
        document.getElementById('currentUser').textContent = userData.username || '管理员';
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('authToken');
    localStorage.removeItem('currentUser');
    authToken = null;
    currentUser = null;
    
    document.getElementById('mainApp').style.display = 'none';
    document.getElementById('mainApp').classList.add('main-app-hidden');
    document.getElementById('loginPage').style.display = 'flex';
    showAlert('已成功退出登录', 'info');
}

// API请求辅助函数
async function apiRequest(url, options = {}) {
    console.log('=== API Request Debug ===');
    console.log('URL:', url);
    console.log('Current authToken variable:', authToken);
    console.log('LocalStorage authToken:', localStorage.getItem('authToken'));
    
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
        }
    };
    
    if (authToken) {
        defaultOptions.headers['Authorization'] = `Bearer ${authToken}`;
        console.log('API Request with token (first 20 chars):', authToken.substring(0, 20) + '...');
        console.log('Full Authorization header:', defaultOptions.headers['Authorization']);
    } else {
        console.log('API Request without token - authToken is:', authToken);
    }
    
    const mergedOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers
        }
    };
    
    console.log('Final request options:', JSON.stringify(mergedOptions, null, 2));
    
    try {
        const response = await fetch(`${API_BASE}${url}`, mergedOptions);
        console.log('Response status:', response.status);
        console.log('Response headers:', [...response.headers.entries()]);
        
        if (!response.ok) {
            console.log('API Response not OK:', response.status, response.statusText);
            const errorText = await response.text();
            console.log('Error response body:', errorText);
            throw new Error(`HTTP ${response.status}: ${errorText}`);
        }
        
        const data = await response.json();
        console.log('Success response data:', data);
        console.log('=== End API Request Debug ===');
        
        return data;
    } catch (error) {
        console.error('API Request failed:', error);
        console.log('=== End API Request Debug (Error) ===');
        throw error;
    }
}

// 加载仪表盘数据
async function loadDashboardData() {
    try {
        // 并行加载所有数据
        const [activeQRsResponse, staticQRsResponse, statsResponse] = await Promise.all([
            apiRequest('/active-qrcodes'),
            apiRequest('/static-qrcodes'),  
            apiRequest('/statistics')
        ]);
        
        // 提取数据，处理不同的响应格式
        const activeQRs = activeQRsResponse.data?.data || activeQRsResponse.data || activeQRsResponse || [];
        const staticQRs = staticQRsResponse.data?.data || staticQRsResponse.data || staticQRsResponse || [];
        const stats = statsResponse.data || statsResponse || {};
        
        // 更新统计数字
        document.getElementById('totalActiveQR').textContent = Array.isArray(activeQRs) ? activeQRs.length : 0;
        document.getElementById('totalStaticQR').textContent = Array.isArray(staticQRs) ? staticQRs.length : 0;
        document.getElementById('totalScans').textContent = stats.total_scans || 0;
        document.getElementById('totalUsers').textContent = stats.total_users || 1;
        
        // 加载最近活动
        loadRecentActivity();
        
        // 加载热门活码
        loadTopQRCodes(activeQRs);
        
    } catch (error) {
        console.error('Failed to load dashboard data:', error);
        showAlert('加载数据失败', 'warning');
    }
}

// 加载最近活动
async function loadRecentActivity() {
    try {
        const response = await apiRequest('/statistics/scan-records?limit=10');
        const activity = response.data || response || [];
        const container = document.getElementById('recentActivity');
        
        if (activity && activity.length > 0) {
            container.innerHTML = activity.map(record => `
                <div class="d-flex align-items-center mb-3">
                    <div class="bg-primary rounded-circle d-flex align-items-center justify-content-center me-3" 
                         style="width: 40px; height: 40px;">
                        <i class="bi bi-qr-code-scan text-white"></i>
                    </div>
                    <div class="flex-grow-1">
                        <div class="fw-bold">${record.target_url || '未知URL'}</div>
                        <small class="text-muted">${formatTime(record.created_at)}</small>
                    </div>
                    <span class="badge bg-success">扫描</span>
                </div>
            `).join('');
        } else {
            container.innerHTML = '<div class="text-center text-muted py-4">暂无活动记录</div>';
        }
    } catch (error) {
        console.error('Failed to load recent activity:', error);
        document.getElementById('recentActivity').innerHTML = 
            '<div class="text-center text-muted py-4">加载失败</div>';
    }
}

// 加载热门活码
function loadTopQRCodes(activeQRs) {
    const container = document.getElementById('topQRCodes');
    
    if (activeQRs && activeQRs.length > 0) {
        // 按扫描次数排序（模拟数据）
        const topQRs = activeQRs.slice(0, 5);
        
        container.innerHTML = topQRs.map((qr, index) => `
            <div class="d-flex align-items-center mb-3">
                <div class="bg-info rounded-circle d-flex align-items-center justify-content-center me-3 text-white fw-bold" 
                     style="width: 30px; height: 30px; font-size: 0.8rem;">
                    #${index + 1}
                </div>
                <div class="flex-grow-1">
                    <div class="fw-bold">${qr.name}</div>
                    <small class="text-muted">${qr.short_code}</small>
                </div>
                <span class="badge bg-primary">${Math.floor(Math.random() * 100) + 1}</span>
            </div>
        `).join('');
    } else {
        container.innerHTML = '<div class="text-center text-muted py-4">暂无数据</div>';
    }
}

// 根据选择的section加载数据
function loadSectionData(section) {
    switch (section) {
        case 'dashboard':
            loadDashboardData();
            break;
        case 'active-qrcodes':
            loadActiveQRCodes();
            break;
        case 'static-qrcodes':
            loadStaticQRCodes();
            loadActiveQROptions();
            break;
        case 'statistics':
            loadStatistics();
            // 延迟初始化图表，确保DOM已渲染
            setTimeout(() => {
                initStatisticsCharts();
            }, 100);
            break;
        default:
            break;
    }
}

// 加载活码数据
async function loadActiveQRCodes() {
    try {
        const response = await apiRequest('/active-qrcodes');
        const activeQRs = response.data?.data || response.data || response || [];
        console.log('Loaded active QR codes:', activeQRs); // 调试信息
        const tbody = document.getElementById('activeQRTable');
        
        if (activeQRs && activeQRs.length > 0) {
            tbody.innerHTML = activeQRs.map(qr => `
                <tr>
                    <td>${qr.id}</td>
                    <td>${qr.name}</td>
                    <td>
                        <code>${qr.short_code || 'N/A'}</code>
                        <button class="btn btn-sm btn-outline-secondary ms-2 copy-shortcode-btn" data-shortcode="${qr.short_code || ''}">
                            <i class="bi bi-clipboard"></i>
                        </button>
                    </td>
                    <td><span class="badge bg-info">${qr.switch_rule || 'unknown'}</span></td>
                    <td>
                        <span class="badge ${qr.status ? 'bg-success' : 'bg-secondary'}">
                            ${qr.status ? '启用' : '禁用'}
                        </span>
                    </td>
                    <td>
                        <img src="/api/public/active-qrcodes/${qr.id}/image" class="qr-code-preview" alt="QR Code" 
                             onerror="this.src='data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwIiBoZWlnaHQ9IjEwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwIiBoZWlnaHQ9IjEwMCIgZmlsbD0iI2Y4ZjlmYSIvPjx0ZXh0IHg9IjUwIiB5PSI1MCIgZm9udC1mYW1pbHk9IkFyaWFsIiBmb250LXNpemU9IjEwIiBmaWxsPSIjNmM3NTdkIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBkeT0iMC4zZW0iPkVSUk9SPC90ZXh0Pjwvc3ZnPg=='">
                    </td>
                    <td>${formatTime(qr.created_at)}</td>
                    <td>
                        <div class="btn-group" role="group">
                            <button class="btn btn-sm btn-outline-primary" onclick="viewActiveQR(${qr.id})" title="查看">
                                <i class="bi bi-eye"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-warning" onclick="editActiveQR(${qr.id})" title="编辑">
                                <i class="bi bi-pencil"></i>
                            </button>
                            <button class="btn btn-sm ${qr.status ? 'btn-outline-secondary' : 'btn-outline-success'}" 
                                    onclick="toggleActiveQRStatus(${qr.id})" title="${qr.status ? '禁用' : '启用'}">
                                <i class="bi ${qr.status ? 'bi-pause-circle' : 'bi-play-circle'}"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-danger" onclick="deleteActiveQR(${qr.id})" title="删除">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = '<tr><td colspan="8" class="text-center text-muted py-4">暂无活码数据</td></tr>';
        }
    } catch (error) {
        console.error('Failed to load active QR codes:', error);
        document.getElementById('activeQRTable').innerHTML = 
            '<tr><td colspan="8" class="text-center text-danger py-4">加载失败</td></tr>';
    }
}

// 加载静态码数据
async function loadStaticQRCodes() {
    try {
        const response = await apiRequest('/static-qrcodes');
        const staticQRs = response.data?.data || response.data || response || [];
        const tbody = document.getElementById('staticQRTable');
        
        if (staticQRs && staticQRs.length > 0) {
            tbody.innerHTML = staticQRs.map(qr => `
                <tr>
                    <td>${qr.id}</td>
                    <td>${qr.active_qr_code?.name || '未关联'}</td>
                    <td>${qr.name}</td>
                    <td>
                        <a href="${qr.target_url}" target="_blank" class="text-decoration-none" title="${qr.target_url}">
                            ${qr.target_url.length > 40 ? qr.target_url.substring(0, 40) + '...' : qr.target_url}
                        </a>
                    </td>
                    <td><span class="badge bg-primary">${qr.weight || 1}</span></td>
                    <td>
                        <span class="badge ${qr.status ? 'bg-success' : 'bg-secondary'}">
                            ${qr.status ? '启用' : '禁用'}
                        </span>
                    </td>
                    <td>
                        ${formatTimeRange(qr.start_time, qr.end_time)}
                    </td>
                    <td>${formatTime(qr.created_at)}</td>
                    <td>
                        <div class="btn-group" role="group">
                            <button class="btn btn-sm btn-outline-primary" onclick="viewStaticQR(${qr.id})" title="查看详情">
                                <i class="bi bi-eye"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-warning" onclick="editStaticQR(${qr.id})" title="编辑">
                                <i class="bi bi-pencil"></i>
                            </button>
                            <button class="btn btn-sm ${qr.status ? 'btn-outline-secondary' : 'btn-outline-success'}" 
                                    onclick="toggleStaticQRStatus(${qr.id})" title="${qr.status ? '禁用' : '启用'}">
                                <i class="bi ${qr.status ? 'bi-pause-circle' : 'bi-play-circle'}"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-danger" onclick="deleteStaticQR(${qr.id})" title="删除">
                                <i class="bi bi-trash"></i>
                            </button>
                        </div>
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = '<tr><td colspan="9" class="text-center text-muted py-4">暂无静态码数据</td></tr>';
        }
    } catch (error) {
        console.error('Failed to load static QR codes:', error);
        document.getElementById('staticQRTable').innerHTML = 
            '<tr><td colspan="9" class="text-center text-danger py-4">加载失败</td></tr>';
    }
}

// 创建活码
async function createActiveQR() {
    const name = document.getElementById('activeQRName').value;
    const switchRule = document.getElementById('switchRule').value;
    const description = document.getElementById('activeQRDesc').value;
    
    if (!name.trim()) {
        showAlert('请输入活码名称', 'warning');
        return;
    }
    
    try {
        const newQR = await apiRequest('/active-qrcodes', {
            method: 'POST',
            body: JSON.stringify({
                name: name.trim(),
                switch_rule: switchRule,
                description: description.trim()
            })
        });
        
        showAlert('活码创建成功！', 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('createActiveQRModal'));
        modal.hide();
        
        // 清空表单
        document.getElementById('createActiveQRForm').reset();
        
        // 重新加载活码列表
        loadActiveQRCodes();
        
    } catch (error) {
        console.error('Failed to create active QR:', error);
        showAlert('创建失败: ' + error.message, 'danger');
    }
}

// 加载活码选项到下拉列表
async function loadActiveQROptions() {
    try {
        const response = await apiRequest('/active-qrcodes');
        const activeQRs = response.data?.data || response.data || response || [];
        
        const selectElements = [
            document.getElementById('staticQRActiveQRId'),
            document.getElementById('filterActiveQR')
        ];
        
        selectElements.forEach(select => {
            if (select) {
                // 保留第一个默认选项
                const firstOption = select.querySelector('option');
                select.innerHTML = '';
                if (firstOption) {
                    select.appendChild(firstOption);
                }
                
                // 添加活码选项
                activeQRs.forEach(activeQR => {
                    const option = document.createElement('option');
                    option.value = activeQR.id;
                    option.textContent = activeQR.name;
                    select.appendChild(option);
                });
            }
        });
    } catch (error) {
        console.error('Failed to load active QR options:', error);
    }
}

// 创建静态码
async function createStaticQR() {
    const activeQRId = document.getElementById('staticQRActiveQRId').value;
    const name = document.getElementById('staticQRName').value;
    const targetURL = document.getElementById('staticQRTargetURL').value;
    const weight = parseInt(document.getElementById('staticQRWeight').value) || 1;
    const status = parseInt(document.getElementById('staticQRStatus').value);
    const startTime = document.getElementById('staticQRStartTime').value;
    const endTime = document.getElementById('staticQREndTime').value;
    const regions = document.getElementById('staticQRRegions').value;
    const devices = document.getElementById('staticQRDevices').value;
    
    if (!activeQRId) {
        showAlert('请选择所属活码', 'warning');
        return;
    }
    
    if (!name.trim()) {
        showAlert('请输入静态码名称', 'warning');
        return;
    }
    
    if (!targetURL.trim()) {
        showAlert('请输入目标URL', 'warning');
        return;
    }
    
    // 处理地区和设备限制
    const allowedRegions = regions.trim() ? regions.split(',').map(r => r.trim()).filter(r => r) : [];
    const allowedDevices = devices.trim() ? devices.split(',').map(d => d.trim()).filter(d => d) : [];
    
    const requestData = {
        active_qr_code_id: parseInt(activeQRId),
        name: name.trim(),
        target_url: targetURL.trim(),
        weight: weight,
        status: status,
        start_time: startTime ? convertFromBeijingTime(startTime) : null,
        end_time: endTime ? convertFromBeijingTime(endTime) : null,
        allowed_regions: allowedRegions.length > 0 ? JSON.stringify(allowedRegions) : '',
        allowed_devices: allowedDevices.length > 0 ? JSON.stringify(allowedDevices) : ''
    };
    
    try {
        await apiRequest('/static-qrcodes', {
            method: 'POST',
            body: JSON.stringify(requestData)
        });
        
        showAlert('静态码创建成功！', 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('staticQRModal'));
        modal.hide();
        
        // 清空表单
        document.getElementById('createStaticQRForm').reset();
        
        // 重新加载静态码列表
        loadStaticQRCodes();
        
    } catch (error) {
        console.error('Failed to create static QR:', error);
        showAlert('创建失败: ' + error.message, 'danger');
    }
}

// 过滤静态码
function filterStaticQRCodes() {
    // 这个函数将在后端实现分页和过滤时使用
    // 目前先重新加载数据
    loadStaticQRCodes();
}

// 删除活码
async function deleteActiveQR(id) {
    if (!confirm('确定要删除这个活码吗？此操作不可恢复。')) {
        return;
    }
    
    try {
        await apiRequest(`/active-qrcodes/${id}`, { method: 'DELETE' });
        showAlert('活码删除成功！', 'success');
        loadActiveQRCodes();
    } catch (error) {
        console.error('Failed to delete active QR:', error);
        showAlert('删除失败: ' + error.message, 'danger');
    }
}

// 复制到剪贴板
function copyToClipboard(text) {
    copyToClipboardWithCallback(text, function(success) {
        // 默认的成功/失败处理已经在函数内部完成
    });
}

// 带回调的复制函数
function copyToClipboardWithCallback(text, callback) {
    console.log('Copying text:', text); // 调试信息
    console.log('BASE_URL:', BASE_URL); // 调试信息
    
    const fullUrl = `${BASE_URL}/r/${text}`;
    console.log('Full URL:', fullUrl); // 调试信息
    
    // 尝试使用现代的 Clipboard API
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(fullUrl).then(() => {
            showAlert('链接已复制到剪贴板！', 'success');
            if (callback) callback(true);
        }).catch(err => {
            console.error('Clipboard API failed: ', err);
            // 如果现代API失败，使用传统方法
            fallbackCopyTextToClipboard(fullUrl, callback);
        });
    } else {
        // 如果不支持现代API，直接使用传统方法
        console.log('Clipboard API not available, using fallback method');
        fallbackCopyTextToClipboard(fullUrl, callback);
    }
}

// 传统的复制方法（兼容性更好）
function fallbackCopyTextToClipboard(text, callback) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    
    // 避免滚动到底部
    textArea.style.top = '0';
    textArea.style.left = '0';
    textArea.style.position = 'fixed';
    textArea.style.width = '2em';
    textArea.style.height = '2em';
    textArea.style.padding = '0';
    textArea.style.border = 'none';
    textArea.style.outline = 'none';
    textArea.style.boxShadow = 'none';
    textArea.style.background = 'transparent';
    
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        const successful = document.execCommand('copy');
        if (successful) {
            showAlert('链接已复制到剪贴板！', 'success');
            console.log('Fallback copy successful');
            if (callback) callback(true);
        } else {
            showAlert('复制失败，请手动复制链接', 'warning');
            console.log('Fallback copy failed');
            // 显示手动复制模态框
            showCopyLinkModal(text);
            if (callback) callback(false);
        }
    } catch (err) {
        console.error('Fallback copy error:', err);
        showAlert('复制失败，请手动复制链接', 'warning');
        // 显示手动复制模态框
        showCopyLinkModal(text);
        if (callback) callback(false);
    }
    
    document.body.removeChild(textArea);
}

// 显示复制链接模态框
function showCopyLinkModal(url) {
    const modal = new bootstrap.Modal(document.getElementById('copyLinkModal'));
    const input = document.getElementById('copyLinkInput');
    input.value = url;
    modal.show();
    
    // 选中输入框中的文本
    setTimeout(() => {
        input.select();
        input.focus();
    }, 300);
}

// 手动复制链接
function manualCopyLink() {
    const input = document.getElementById('copyLinkInput');
    input.select();
    
    try {
        const successful = document.execCommand('copy');
        if (successful) {
            showAlert('链接已复制到剪贴板！', 'success');
        } else {
            showAlert('请使用 Ctrl+C 或 Cmd+C 复制链接', 'info');
        }
    } catch (err) {
        showAlert('请使用 Ctrl+C 或 Cmd+C 复制链接', 'info');
    }
}

// 显示提示信息
function showAlert(message, type = 'info') {
    // 创建alert元素
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
    alertDiv.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
    alertDiv.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(alertDiv);
    
    // 3秒后自动消失
    setTimeout(() => {
        if (alertDiv.parentNode) {
            alertDiv.remove();
        }
    }, 3000);
}

// 格式化时间
function formatTime(timeString) {
    if (!timeString) return '未知';
    
    const date = new Date(timeString);
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) { // 1分钟内
        return '刚刚';
    } else if (diff < 3600000) { // 1小时内
        return Math.floor(diff / 60000) + '分钟前';
    } else if (diff < 86400000) { // 1天内
        return Math.floor(diff / 3600000) + '小时前';
    } else {
        return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    }
}

// 切换活码启用/禁用状态
async function toggleActiveQRStatus(id) {
    if (!confirm('确定要切换此活码的状态吗？')) {
        return;
    }
    
    try {
        const response = await apiRequest(`/active-qrcodes/${id}/toggle-status`, {
            method: 'PATCH'
        });
        
        if (response.success) {
            showAlert(response.message, 'success');
            // 重新加载活码列表
            loadActiveQRCodes();
        } else {
            showAlert(response.message || '状态切换失败', 'danger');
        }
    } catch (error) {
        console.error('Failed to toggle active QR status:', error);
        showAlert('状态切换失败: ' + error.message, 'danger');
    }
}

// 切换静态码启用/禁用状态
async function toggleStaticQRStatus(id) {
    if (!confirm('确定要切换此静态码的状态吗？')) {
        return;
    }
    
    try {
        const response = await apiRequest(`/static-qrcodes/${id}/toggle-status`, {
            method: 'PATCH'
        });
        
        if (response.success) {
            showAlert(response.message, 'success');
            // 重新加载静态码列表
            loadStaticQRCodes();
        } else {
            showAlert(response.message || '状态切换失败', 'danger');
        }
    } catch (error) {
        console.error('Failed to toggle static QR status:', error);
        showAlert('状态切换失败: ' + error.message, 'danger');
    }
}

// 刷新数据
function refreshData() {
    const activeSection = document.querySelector('.content-section.active');
    if (activeSection) {
        const sectionId = activeSection.id;
        loadSectionData(sectionId);
    }
}

// 查看活码详情
async function viewActiveQR(id) {
    try {
        const response = await apiRequest(`/active-qrcodes/${id}`);
        const activeQR = response.data; // 从响应中提取 data 字段
        
        console.log('查看活码详情数据:', activeQR); // 调试日志
        
        // 填充详情模态框
        document.getElementById('viewActiveQRName').textContent = activeQR.name;
        document.getElementById('viewActiveQRShortCode').textContent = activeQR.short_code;
        document.getElementById('viewActiveQRRule').textContent = activeQR.switch_rule;
        document.getElementById('viewActiveQRDesc').textContent = activeQR.description || '无描述';
        document.getElementById('viewActiveQRStatus').textContent = activeQR.status ? '启用' : '禁用';
        document.getElementById('viewActiveQRCreated').textContent = formatTime(activeQR.created_at);
        
        // 生成完整链接
        const fullLink = `${BASE_URL}/r/${activeQR.short_code}`;
        document.getElementById('viewActiveQRLink').textContent = fullLink;
        document.getElementById('viewActiveQRLink').href = fullLink;
        
        // 显示关联的静态码
        const staticQRList = document.getElementById('viewStaticQRList');
        if (activeQR.static_qr_codes && activeQR.static_qr_codes.length > 0) {
            staticQRList.innerHTML = activeQR.static_qr_codes.map(sqr => `
                <div class="list-group-item d-flex justify-content-between align-items-center">
                    <div>
                        <strong>${sqr.name}</strong>
                        <br>
                        <small class="text-muted">${sqr.target_url}</small>
                        <span class="badge bg-primary ms-2">权重: ${sqr.weight}</span>
                    </div>
                    <span class="badge ${sqr.status ? 'bg-success' : 'bg-secondary'}">
                        ${sqr.status ? '启用' : '禁用'}
                    </span>
                </div>
            `).join('');
        } else {
            staticQRList.innerHTML = '<div class="text-muted text-center py-3">暂无关联的静态码</div>';
        }
        
        // 显示模态框
        const modal = new bootstrap.Modal(document.getElementById('viewActiveQRModal'));
        modal.show();
        
    } catch (error) {
        console.error('Failed to load active QR details:', error);
        showAlert('加载活码详情失败: ' + error.message, 'danger');
    }
}

// 编辑活码
async function editActiveQR(id) {
    try {
        const response = await apiRequest(`/active-qrcodes/${id}`);
        const activeQR = response.data; // 从响应中提取 data 字段
        
        console.log('编辑活码数据:', activeQR); // 调试日志
        
        // 填充编辑表单
        document.getElementById('editActiveQRId').value = id;
        document.getElementById('editActiveQRName').value = activeQR.name;
        document.getElementById('editSwitchRule').value = activeQR.switch_rule;
        document.getElementById('editActiveQRDesc').value = activeQR.description || '';
        
        // 显示模态框
        const modal = new bootstrap.Modal(document.getElementById('editActiveQRModal'));
        modal.show();
        
    } catch (error) {
        console.error('Failed to load active QR for editing:', error);
        showAlert('加载活码信息失败: ' + error.message, 'danger');
    }
}

// 保存活码编辑
async function saveActiveQREdit() {
    const id = document.getElementById('editActiveQRId').value;
    const name = document.getElementById('editActiveQRName').value;
    const switchRule = document.getElementById('editSwitchRule').value;
    const description = document.getElementById('editActiveQRDesc').value;
    
    if (!name.trim()) {
        showAlert('请输入活码名称', 'warning');
        return;
    }
    
    try {
        await apiRequest(`/active-qrcodes/${id}`, {
            method: 'PUT',
            body: JSON.stringify({
                name: name.trim(),
                switch_rule: switchRule,
                description: description.trim()
            })
        });
        
        showAlert('活码更新成功！', 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('editActiveQRModal'));
        modal.hide();
        
        // 重新加载活码列表
        loadActiveQRCodes();
        
    } catch (error) {
        console.error('Failed to update active QR:', error);
        showAlert('更新失败: ' + error.message, 'danger');
    }
}

// 查看静态码详情
async function viewStaticQR(id) {
    try {
        const response = await apiRequest(`/static-qrcodes/${id}`);
        const staticQR = response.data; // 从响应中提取 data 字段
        
        console.log('查看静态码详情数据:', staticQR); // 调试日志
        
        // 填充静态码详情
        document.getElementById('viewStaticQRName').textContent = staticQR.name;
        document.getElementById('viewStaticQRURL').textContent = staticQR.target_url;
        document.getElementById('viewStaticQRURL').href = staticQR.target_url;
        document.getElementById('viewStaticQRWeight').textContent = staticQR.weight || 1;
        document.getElementById('viewStaticQRStatus').textContent = staticQR.status ? '启用' : '禁用';
        document.getElementById('viewStaticQRCreated').textContent = formatTime(staticQR.created_at);
        
        // 显示时间范围 - 使用专门的时间范围格式化函数
        document.getElementById('viewStaticQRTimeRange').textContent = formatTimeRange(staticQR.start_time, staticQR.end_time);
        
        // 显示地区和设备限制
        let regions = '无限制';
        let devices = '无限制';
        
        try {
            if (staticQR.allowed_regions && staticQR.allowed_regions !== 'null') {
                const regionList = JSON.parse(staticQR.allowed_regions);
                regions = regionList.length > 0 ? regionList.join(', ') : '无限制';
            }
            if (staticQR.allowed_devices && staticQR.allowed_devices !== 'null') {
                const deviceList = JSON.parse(staticQR.allowed_devices);
                devices = deviceList.length > 0 ? deviceList.join(', ') : '无限制';
            }
        } catch (e) {
            console.warn('Failed to parse restrictions:', e);
        }
        
        document.getElementById('viewStaticQRRegions').textContent = regions;
        document.getElementById('viewStaticQRDevices').textContent = devices;
        
        // 显示模态框
        const modal = new bootstrap.Modal(document.getElementById('viewStaticQRModal'));
        modal.show();
        
    } catch (error) {
        console.error('Failed to load static QR details:', error);
        showAlert('加载静态码详情失败: ' + error.message, 'danger');
    }
}

// 编辑静态码
async function editStaticQR(id) {
    try {
        const response = await apiRequest(`/static-qrcodes/${id}`);
        const staticQR = response.data; // 从响应中提取 data 字段
        
        console.log('编辑静态码数据:', staticQR); // 调试日志
        
        // 填充编辑表单
        document.getElementById('editStaticQRId').value = id;
        document.getElementById('editStaticQRName').value = staticQR.name;
        document.getElementById('editStaticQRURL').value = staticQR.target_url;
        document.getElementById('editStaticQRWeight').value = staticQR.weight || 1;
        
        // 时间范围 - 转换为北京时间显示
        if (staticQR.start_time) {
            document.getElementById('editStaticQRStartTime').value = convertToBeijingTime(staticQR.start_time);
        } else {
            document.getElementById('editStaticQRStartTime').value = '';
        }
        if (staticQR.end_time) {
            document.getElementById('editStaticQREndTime').value = convertToBeijingTime(staticQR.end_time);
        } else {
            document.getElementById('editStaticQREndTime').value = '';
        }
        
        // 地区和设备限制
        try {
            if (staticQR.allowed_regions && staticQR.allowed_regions !== 'null') {
                const regions = JSON.parse(staticQR.allowed_regions);
                document.getElementById('editStaticQRRegions').value = regions.join(', ');
            }
            if (staticQR.allowed_devices && staticQR.allowed_devices !== 'null') {
                const devices = JSON.parse(staticQR.allowed_devices);
                document.getElementById('editStaticQRDevices').value = devices.join(', ');
            }
        } catch (e) {
            console.warn('Failed to parse restrictions for editing:', e);
        }
        
        // 显示模态框
        const modal = new bootstrap.Modal(document.getElementById('editStaticQRModal'));
        modal.show();
        
    } catch (error) {
        console.error('Failed to load static QR for editing:', error);
        showAlert('加载静态码信息失败: ' + error.message, 'danger');
    }
}

// 删除静态码
async function deleteStaticQR(id) {
    if (!confirm('确定要删除这个静态码吗？此操作不可恢复。')) {
        return;
    }
    
    try {
        await apiRequest(`/static-qrcodes/${id}`, { method: 'DELETE' });
        showAlert('静态码删除成功！', 'success');
        loadStaticQRCodes();
    } catch (error) {
        console.error('Failed to delete static QR:', error);
        showAlert('删除失败: ' + error.message, 'danger');
    }
}

// 保存静态码编辑
async function saveStaticQREdit() {
    const id = document.getElementById('editStaticQRId').value;
    const name = document.getElementById('editStaticQRName').value;
    const targetURL = document.getElementById('editStaticQRURL').value;
    const weight = document.getElementById('editStaticQRWeight').value;
    const startTime = document.getElementById('editStaticQRStartTime').value;
    const endTime = document.getElementById('editStaticQREndTime').value;
    const regions = document.getElementById('editStaticQRRegions').value;
    const devices = document.getElementById('editStaticQRDevices').value;
    
    if (!name.trim() || !targetURL.trim()) {
        showAlert('请填写必要字段', 'warning');
        return;
    }
    
    // 处理地区和设备限制
    const allowedRegions = regions.trim() ? regions.split(',').map(r => r.trim()).filter(r => r) : [];
    const allowedDevices = devices.trim() ? devices.split(',').map(d => d.trim()).filter(d => d) : [];
    
    const requestData = {
        name: name.trim(),
        target_url: targetURL.trim(),
        weight: parseInt(weight) || 1,
        allowed_regions: allowedRegions.length > 0 ? JSON.stringify(allowedRegions) : '',
        allowed_devices: allowedDevices.length > 0 ? JSON.stringify(allowedDevices) : ''
    };
    
    // 添加时间范围 - 从北京时间转换为UTC
    if (startTime) {
        requestData.start_time = convertFromBeijingTime(startTime);
    }
    if (endTime) {
        requestData.end_time = convertFromBeijingTime(endTime);
    }
    
    try {
        await apiRequest(`/static-qrcodes/${id}`, {
            method: 'PUT',
            body: JSON.stringify(requestData)
        });
        
        showAlert('静态码更新成功！', 'success');
        
        // 关闭模态框
        const modal = bootstrap.Modal.getInstance(document.getElementById('editStaticQRModal'));
        modal.hide();
        
        // 重新加载静态码列表
        loadStaticQRCodes();
        
    } catch (error) {
        console.error('Failed to update static QR:', error);
        showAlert('更新失败: ' + error.message, 'danger');
    }
}

// 加载统计数据
function loadStatistics() {
    // 加载统计图表数据
    console.log('Loading statistics charts...');
    // 图表初始化已在 loadSectionData 中处理
}

// 查看活码详情（占位符函数）
function viewActiveQR(id) {
    showAlert(`查看活码 ID: ${id}`, 'info');
}

// 编辑活码（占位符函数）

// 添加复制短码按钮的事件监听器
document.addEventListener('DOMContentLoaded', function() {
    // 使用事件委托处理复制按钮点击
    document.addEventListener('click', function(e) {
        if (e.target.closest('.copy-shortcode-btn')) {
            const button = e.target.closest('.copy-shortcode-btn');
            const shortcode = button.getAttribute('data-shortcode');
            
            // 添加视觉反馈
            const originalText = button.innerHTML;
            button.innerHTML = '<i class="bi bi-check2"></i> 复制中...';
            button.disabled = true;
            
            // 执行复制
            copyToClipboardWithCallback(shortcode, function(success) {
                // 恢复按钮状态
                setTimeout(() => {
                    if (success) {
                        button.innerHTML = '<i class="bi bi-check2"></i> 已复制';
                        button.className = 'btn btn-sm btn-success ms-2 copy-shortcode-btn';
                    } else {
                        button.innerHTML = '<i class="bi bi-x-circle"></i> 复制失败';
                        button.className = 'btn btn-sm btn-danger ms-2 copy-shortcode-btn';
                    }
                    
                    // 2秒后恢复原状态
                    setTimeout(() => {
                        button.innerHTML = originalText;
                        button.className = 'btn btn-sm btn-outline-secondary ms-2 copy-shortcode-btn';
                        button.disabled = false;
                    }, 2000);
                }, 100);
            });
        }
    });

    // 系统设置表单处理
    const settingsForm = document.getElementById('settingsForm');
    if (settingsForm) {
        settingsForm.addEventListener('submit', function(e) {
            e.preventDefault();
            saveSystemSettings();
        });
        
        // 页面加载时从localStorage加载设置
        loadSystemSettings();
    }
});

// 保存系统设置
function saveSystemSettings() {
    const systemName = document.getElementById('systemName').value.trim();
    const apiBaseUrl = document.getElementById('apiBaseUrl').value.trim();
    
    if (!systemName) {
        showAlert('系统名称不能为空', 'error');
        return;
    }
    
    if (!apiBaseUrl) {
        showAlert('API基础URL不能为空', 'error');
        return;
    }
    
    // 验证URL格式
    try {
        new URL(apiBaseUrl);
    } catch (e) {
        showAlert('API基础URL格式不正确，请输入有效的URL', 'error');
        return;
    }
    
    // 保存到localStorage
    const settings = {
        systemName: systemName,
        apiBaseUrl: apiBaseUrl,
        savedAt: new Date().toISOString()
    };
    
    localStorage.setItem('systemSettings', JSON.stringify(settings));
    
    // 更新全局变量
    BASE_URL = apiBaseUrl;
    API_BASE = `${apiBaseUrl}/api`;
    
    console.log('Settings saved:', { BASE_URL, API_BASE });
    showAlert('系统设置保存成功！', 'success');
}

// 加载系统设置
function loadSystemSettings() {
    try {
        const savedSettings = localStorage.getItem('systemSettings');
        if (savedSettings) {
            const settings = JSON.parse(savedSettings);
            
            // 更新表单字段
            const systemNameInput = document.getElementById('systemName');
            const apiBaseUrlInput = document.getElementById('apiBaseUrl');
            
            if (systemNameInput && settings.systemName) {
                systemNameInput.value = settings.systemName;
            }
            
            if (apiBaseUrlInput && settings.apiBaseUrl) {
                apiBaseUrlInput.value = settings.apiBaseUrl;
                
                // 更新全局变量
                BASE_URL = settings.apiBaseUrl;
                API_BASE = `${settings.apiBaseUrl}/api`;
                
                console.log('Settings loaded from localStorage:', { BASE_URL, API_BASE });
            }
        }
    } catch (e) {
        console.error('Error loading system settings:', e);
    }
}

// ===== 统计图表功能 =====

// 图表实例存储
let scanTrendChart = null;
let trafficDistChart = null;

// 初始化统计图表
function initStatisticsCharts() {
    initScanTrendChart();
    initTrafficDistChart();
    loadStatisticsData();
}

// 初始化扫描趋势图表
function initScanTrendChart() {
    const ctx = document.getElementById('scanTrendChart');
    if (!ctx) return;

    // 销毁已存在的图表
    if (scanTrendChart) {
        scanTrendChart.destroy();
    }

    scanTrendChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: [],
            datasets: [{
                label: '扫描次数',
                data: [],
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: 'rgba(75, 192, 192, 0.2)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            plugins: {
                title: {
                    display: true,
                    text: '最近7天扫描趋势'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        stepSize: 1
                    }
                }
            }
        }
    });
}

// 初始化流量分布图表
function initTrafficDistChart() {
    const ctx = document.getElementById('trafficDistChart');
    if (!ctx) return;

    // 销毁已存在的图表
    if (trafficDistChart) {
        trafficDistChart.destroy();
    }

    trafficDistChart = new Chart(ctx, {
        type: 'doughnut',
        data: {
            labels: [],
            datasets: [{
                data: [],
                backgroundColor: [
                    'rgba(255, 99, 132, 0.8)',
                    'rgba(54, 162, 235, 0.8)',
                    'rgba(255, 205, 86, 0.8)',
                    'rgba(75, 192, 192, 0.8)',
                    'rgba(153, 102, 255, 0.8)',
                    'rgba(255, 159, 64, 0.8)'
                ],
                borderColor: [
                    'rgba(255, 99, 132, 1)',
                    'rgba(54, 162, 235, 1)',
                    'rgba(255, 205, 86, 1)',
                    'rgba(75, 192, 192, 1)',
                    'rgba(153, 102, 255, 1)',
                    'rgba(255, 159, 64, 1)'
                ],
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            plugins: {
                title: {
                    display: true,
                    text: '设备类型分布'
                },
                legend: {
                    position: 'bottom'
                }
            }
        }
    });
}

// 加载统计数据
async function loadStatisticsData() {
    try {
        // 加载扫描趋势数据
        await loadScanTrendData();
        
        // 加载流量分布数据
        await loadTrafficDistData();
        
    } catch (error) {
        console.error('Error loading statistics data:', error);
        showToast('加载统计数据失败', 'error');
    }
}

// 加载扫描趋势数据
async function loadScanTrendData() {
    try {
        console.log('开始加载扫描趋势数据...');
        const response = await fetch(`${API_BASE}/statistics/trends?days=7`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        });

        console.log('趋势数据响应状态:', response.status);
        
        if (!response.ok) {
            throw new Error('Failed to fetch trend data');
        }

        const result = await response.json();
        console.log('趋势数据响应:', result);
        
        if (result.success && result.data) {
            console.log('更新趋势图表，数据:', result.data);
            updateScanTrendChart(result.data);
        } else {
            console.log('趋势数据响应不成功或无数据');
            updateScanTrendChart([]);
        }
    } catch (error) {
        console.error('Error loading scan trend data:', error);
        // 显示空数据而不是模拟数据
        updateScanTrendChart([]);
    }
}

// 加载流量分布数据
async function loadTrafficDistData() {
    try {
        console.log('开始加载流量分布数据...');
        const response = await fetch(`${API_BASE}/statistics/device-stats`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        });

        console.log('设备统计响应状态:', response.status);

        if (!response.ok) {
            throw new Error('Failed to fetch device stats');
        }

        const result = await response.json();
        console.log('设备统计响应:', result);
        
        if (result.success && result.data) {
            // 格式化设备统计数据
            const formattedStats = {};
            Object.keys(result.data).forEach(device => {
                const deviceName = formatDeviceName(device);
                formattedStats[deviceName] = result.data[device];
            });
            console.log('格式化后的设备统计:', formattedStats);
            updateTrafficDistChart(formattedStats);
        } else {
            console.log('设备统计响应不成功或无数据');
            throw new Error('No device stats data');
        }
    } catch (error) {
        console.error('Error loading traffic distribution data:', error);
        // 显示空数据而不是模拟数据
        updateTrafficDistChart({});
    }
}

// 更新扫描趋势图表
function updateScanTrendChart(trendData) {
    console.log('updateScanTrendChart 被调用，数据:', trendData);
    
    if (!scanTrendChart) {
        console.log('scanTrendChart 实例不存在');
        return;
    }
    
    if (!trendData) {
        console.log('trendData 为空');
        return;
    }

    const labels = [];
    const data = [];

    // 处理趋势数据
    if (Array.isArray(trendData)) {
        trendData.forEach(item => {
            const dateLabel = formatDateForChart(item.date || item.day);
            const scanCount = item.scans || item.count || 0;
            labels.push(dateLabel);
            data.push(scanCount);
            console.log(`日期: ${dateLabel}, 扫描数: ${scanCount}`);
        });
    }

    // 如果没有数据，生成最近7天的空数据
    if (labels.length === 0) {
        console.log('没有数据，生成空数据');
        for (let i = 6; i >= 0; i--) {
            const date = new Date();
            date.setDate(date.getDate() - i);
            labels.push(formatDateForChart(date.toISOString().split('T')[0]));
            data.push(0); // 显示0而不是随机数据
        }
    }

    console.log('最终标签:', labels);
    console.log('最终数据:', data);

    scanTrendChart.data.labels = labels;
    scanTrendChart.data.datasets[0].data = data;
    scanTrendChart.update();
    
    console.log('趋势图表已更新');
}

// 更新流量分布图表
function updateTrafficDistChart(deviceStats) {
    console.log('updateTrafficDistChart 被调用，数据:', deviceStats);
    
    if (!trafficDistChart) {
        console.log('trafficDistChart 实例不存在');
        return;
    }

    const labels = Object.keys(deviceStats || {});
    const data = Object.values(deviceStats || {});

    console.log('设备标签:', labels);
    console.log('设备数据:', data);

    // 如果没有数据，显示"暂无数据"
    if (labels.length === 0) {
        console.log('没有设备数据，显示暂无数据');
        labels.push('暂无数据');
        data.push(1);
    }

    trafficDistChart.data.labels = labels;
    trafficDistChart.data.datasets[0].data = data;
    trafficDistChart.update();
    
    console.log('流量分布图表已更新');
}

// 处理设备分布数据
function processDeviceDistribution(scanRecords) {
    const deviceStats = {};
    
    if (Array.isArray(scanRecords)) {
        scanRecords.forEach(record => {
            const device = record.device || 'Unknown';
            const deviceName = formatDeviceName(device);
            deviceStats[deviceName] = (deviceStats[deviceName] || 0) + 1;
        });
    }

    return deviceStats;
}

// 格式化设备名称
function formatDeviceName(device) {
    const deviceMap = {
        'mobile': 'Mobile',
        'desktop': 'Desktop', 
        'tablet': 'Tablet',
        'unknown': 'Unknown'
    };
    
    return deviceMap[device.toLowerCase()] || 'Unknown';
}

// 格式化日期用于图表显示
function formatDateForChart(dateString) {
    const date = new Date(dateString);
    return `${date.getMonth() + 1}/${date.getDate()}`;
}

// 生成模拟趋势数据
function generateMockTrendData() {
    const data = [];
    for (let i = 6; i >= 0; i--) {
        const date = new Date();
        date.setDate(date.getDate() - i);
        data.push({
            date: date.toISOString().split('T')[0],
            count: Math.floor(Math.random() * 50) + 10
        });
    }
    return data;
}

// 刷新统计图表
function refreshStatisticsCharts() {
    loadStatisticsData();
}

// ===== 页面切换事件监听 =====

// 监听统计分析选项卡切换
function onStatisticsTabShow() {
    // 延迟初始化图表，确保DOM已渲染
    setTimeout(() => {
        initStatisticsCharts();
    }, 100);
}

// ===== 二维码上传功能 =====

// 当前操作模式：'create' 或 'edit'
let currentQRUploadMode = 'create';

// 显示二维码上传模态框
function showQRUploadModal(mode = 'create') {
    currentQRUploadMode = mode;
    
    // 重置模态框状态
    resetQRUploadModal();
    
    // 显示模态框
    const modal = new bootstrap.Modal(document.getElementById('qrUploadModal'));
    modal.show();
}

// 重置二维码上传模态框
function resetQRUploadModal() {
    // 清空文件选择
    document.getElementById('qrCodeFile').value = '';
    
    // 隐藏预览区域
    document.getElementById('qrPreviewArea').style.display = 'none';
    document.getElementById('qrPreviewImage').src = '';
    
    // 隐藏结果和错误区域
    document.getElementById('qrParseResult').style.display = 'none';
    document.getElementById('qrParseError').style.display = 'none';
    
    // 重置按钮状态
    document.getElementById('parseQRCodeBtn').disabled = true;
    document.getElementById('useQRCodeBtn').style.display = 'none';
    
    // 清空解析结果
    window.parsedQRCodeURL = null;
}

// 预览上传的图片
function previewQRImage(input) {
    if (input.files && input.files[0]) {
        const file = input.files[0];
        
        // 检查文件类型
        if (!file.type.startsWith('image/')) {
            showAlert('请选择图片文件', 'error');
            return;
        }
        
        // 检查文件大小（限制5MB）
        if (file.size > 5 * 1024 * 1024) {
            showAlert('图片文件大小不能超过5MB', 'error');
            return;
        }
        
        const reader = new FileReader();
        reader.onload = function(e) {
            // 显示预览图片
            document.getElementById('qrPreviewImage').src = e.target.result;
            document.getElementById('qrPreviewArea').style.display = 'block';
            
            // 启用解析按钮
            document.getElementById('parseQRCodeBtn').disabled = false;
            
            // 隐藏之前的结果
            document.getElementById('qrParseResult').style.display = 'none';
            document.getElementById('qrParseError').style.display = 'none';
            document.getElementById('useQRCodeBtn').style.display = 'none';
        };
        reader.readAsDataURL(file);
    }
}

// 解析二维码
async function parseQRCode() {
    const fileInput = document.getElementById('qrCodeFile');
    const file = fileInput.files[0];
    
    if (!file) {
        showAlert('请先选择二维码图片', 'error');
        return;
    }
    
    // 禁用解析按钮，显示加载状态
    const parseBtn = document.getElementById('parseQRCodeBtn');
    const originalText = parseBtn.innerHTML;
    parseBtn.disabled = true;
    parseBtn.innerHTML = '<i class="spinner-border spinner-border-sm me-2" role="status"></i>解析中...';
    
    // 隐藏之前的结果
    document.getElementById('qrParseResult').style.display = 'none';
    document.getElementById('qrParseError').style.display = 'none';
    document.getElementById('useQRCodeBtn').style.display = 'none';
    
    try {
        // 创建FormData
        const formData = new FormData();
        formData.append('qrcode', file);
        
        // 发送请求
        const token = localStorage.getItem('authToken');
        const response = await fetch(`${API_BASE}/tools/parse-qrcode`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            },
            body: formData
        });
        
        const result = await response.json();
        
        if (result.success) {
            // 解析成功
            window.parsedQRCodeURL = result.data.url;
            document.getElementById('parsedURL').textContent = result.data.url;
            document.getElementById('qrParseResult').style.display = 'block';
            document.getElementById('useQRCodeBtn').style.display = 'inline-block';
        } else {
            // 解析失败
            document.getElementById('parseErrorMessage').textContent = result.message || '解析失败';
            document.getElementById('qrParseError').style.display = 'block';
        }
    } catch (error) {
        console.error('Error parsing QR code:', error);
        document.getElementById('parseErrorMessage').textContent = '网络错误，请重试';
        document.getElementById('qrParseError').style.display = 'block';
    } finally {
        // 恢复按钮状态
        parseBtn.disabled = false;
        parseBtn.innerHTML = originalText;
    }
}

// 使用解析出的URL
function useQRCodeURL() {
    if (!window.parsedQRCodeURL) {
        showAlert('没有可用的URL', 'error');
        return;
    }
    
    // 根据当前模式填充对应的输入框
    if (currentQRUploadMode === 'create') {
        document.getElementById('staticQRTargetURL').value = window.parsedQRCodeURL;
    } else if (currentQRUploadMode === 'edit') {
        document.getElementById('editStaticQRURL').value = window.parsedQRCodeURL;
    }
    
    // 关闭模态框
    const modal = bootstrap.Modal.getInstance(document.getElementById('qrUploadModal'));
    modal.hide();
    
    showAlert('URL填充成功', 'success');
}

