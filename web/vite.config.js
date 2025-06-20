import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  define: {
    global: 'globalThis',
  },
  server: {
    port: 5173,
    proxy: {
      // Proxy REST API and WebSocket connections to backend
      "/api": {
        target: "http://localhost:8080",
        ws: true,
      },
      // Proxy webhook endpoints to backend
      "/webhooks": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      external: [],
    },
  },
  resolve: {
    alias: {
      url: 'url',
    },
  },
  optimizeDeps: {
    include: ['@mel-agent/api-client'],
  },
});
