import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  base: '/static/', // 设置静态资源的根路径
  plugins: [vue()],
  // plugins: [vue()],
  // server: {
  //   proxy: {
  //     '/api': 'http://127.0.0.1:8000',  // 假设 GoFrame 运行在 8000 端口
  //     '/captcha': 'http://localhost:8000',  // 假设 GoFrame 运行在 8000 端
  //   },
  // },
})
