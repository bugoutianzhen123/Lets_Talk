import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// 本地开发：把后端接口与 WebSocket 代理到 8080。
// 前端统一使用相对路径请求，开发环境与线上（Nginx 反代）行为保持一致，
// 无需在代码里写死后端地址。
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/user': { target: 'http://localhost:8080', changeOrigin: true },
      '/friend': { target: 'http://localhost:8080', changeOrigin: true },
      '/group': { target: 'http://localhost:8080', changeOrigin: true },
      '/ws': { target: 'ws://localhost:8080', ws: true },
    },
  },
});
