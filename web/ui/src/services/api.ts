import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';

// API基础URL
const API_BASE_URL = '/api';

// 创建axios实例
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      // 未授权，清除token并重定向到登录页
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// API服务
const apiService = {
  // 认证相关
  auth: {
    login: (clientId: string, secret: string) => {
      return api.post('/auth/login', { client_id: clientId, secret });
    },
    validateToken: () => {
      return api.get('/auth/validate');
    },
  },

  // 插件管理
  plugins: {
    list: () => {
      return api.get('/plugins');
    },
    getInfo: (id: string) => {
      return api.get(`/plugins/${id}`);
    },
    start: (id: string) => {
      return api.post(`/plugins/${id}/start`);
    },
    stop: (id: string) => {
      return api.post(`/plugins/${id}/stop`);
    },
    executeCommand: (plugin: string, command: string, args: string[]) => {
      return api.post('/command', { plugin, command, args });
    },
  },

  // 文件管理
  files: {
    list: (path: string = '.') => {
      return api.get('/files', { params: { path } });
    },
    upload: (file: File, remotePath: string, options: { compress?: boolean, overwrite?: boolean } = {}) => {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('remote_path', remotePath);
      if (options.compress) {
        formData.append('compress', 'true');
      }
      if (options.overwrite) {
        formData.append('overwrite', 'true');
      }
      return api.post('/files/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    },
    download: (path: string) => {
      window.open(`${API_BASE_URL}/files/download?path=${encodeURIComponent(path)}`, '_blank');
    },
    delete: (path: string) => {
      return api.delete('/files', { params: { path } });
    },
    mkdir: (path: string) => {
      return api.post('/files/mkdir', { path });
    },
  },

  // 终端管理
  terminals: {
    list: () => {
      return api.get('/terminals');
    },
    create: (id: string, command: string = '', args: string[] = []) => {
      return api.post('/terminals', { id, command, args });
    },
    kill: (id: string) => {
      return api.delete(`/terminals/${id}`);
    },
    write: (id: string, data: string) => {
      return api.post('/terminals/write', { id, data });
    },
    read: (id: string) => {
      return api.get(`/terminals/${id}/read`);
    },
  },

  // 代理服务
  proxy: {
    getStatus: () => {
      return api.get('/proxy/status');
    },
    start: (type: string) => {
      return api.post(`/proxy/${type}/start`);
    },
    stop: (type: string) => {
      return api.post(`/proxy/${type}/stop`);
    },
  },

  // Shell命令
  shell: {
    execute: (command: string) => {
      return api.post('/shell/exec', { command });
    },
  },
};

export default apiService;
