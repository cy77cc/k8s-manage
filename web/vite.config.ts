import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import { visualizer } from 'rollup-plugin-visualizer'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // 5.1.11 打包体积分析
    visualizer({
      open: false,
      filename: 'dist/stats.html',
      gzipSize: true,
      brotliSize: true,
    }),
  ],
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setupTests.ts',
    globals: false,
    css: true,
    testTimeout: 15000,
    hookTimeout: 15000,
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
      },
    },
  },
  // 5.1.8 Vite 打包优化
  build: {
    // 5.1.10 启用代码压缩
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境移除 console
        drop_debugger: true,
      },
    },
    // 代码分割优化
    rollupOptions: {
      output: {
        // 手动分块
        manualChunks: {
          // React 核心库
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          // Ant Design
          'antd-vendor': ['antd', '@ant-design/icons'],
          // 动画库
          'animation-vendor': ['framer-motion'],
          // 工具库
          'utils-vendor': ['axios', 'dayjs'],
        },
        // 文件命名
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]',
      },
    },
    // 5.1.9 启用 Tree Shaking (默认启用)
    // chunk 大小警告限制
    chunkSizeWarningLimit: 1000,
    // 启用 CSS 代码分割
    cssCodeSplit: true,
    // 生成 source map
    sourcemap: false,
  },
  // 优化依赖预构建
  optimizeDeps: {
    include: [
      'react',
      'react-dom',
      'react-router-dom',
      'antd',
      '@ant-design/icons',
      'framer-motion',
      'axios',
    ],
  },
})
