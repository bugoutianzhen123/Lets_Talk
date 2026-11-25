import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// 本地开发：把后端接口与 WebSocket 代理到 8080。
// 前端统一使用相对路径请求，开发环境与线上（Nginx 反代）行为保持一致，
// 无需在代码里写死后端地址。
const proxy = {
  '/user': { target: 'http://localhost:8080', changeOrigin: true },
  '/friend': { target: 'http://localhost:8080', changeOrigin: true },
  '/group': { target: 'http://localhost:8080', changeOrigin: true },
  '/ws': { target: 'ws://localhost:8080', ws: true },
};

export default defineConfig({
  plugins: [react()],
  server: {
    host: true, // 监听 0.0.0.0，允许通过服务器公网 IP 访问
    proxy,
  },
  // 生产构建预览：演示时用 `npm run build && npm run preview`，
  // 直接提供打包后的静态文件，首屏更快、更稳定（避免开发服务器现场编译）
  preview: {
    host: true,
    port: 5173,
    strictPort: true,
    proxy,
  },
});
