import { createApp } from 'vue';
import App from './App.vue';
import 'ant-design-vue/dist/reset.css';  // 引入 Ant Design 样式
import Antd from 'ant-design-vue';     // 引入 Ant Design 组件库

createApp(App)
  .use(Antd)                          // 使用 Ant Design 组件库
  .mount('#app');
