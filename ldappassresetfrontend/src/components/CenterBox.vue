<template>
    <div class="center-box">
      <a-card class="card-container" bordered={false}>
        <slot></slot> <!-- 插槽，子组件的内容会被渲染到这里 -->
      </a-card>
    </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { Card } from 'ant-design-vue';  // 导入 Card 组件

export default defineComponent({
name: 'CenterBox',
components: {
    Card,  // 注册 Card 组件
},
});
</script>

<style scoped>
.center-box {
display: flex;
justify-content: center;  /* 水平居中 */
align-items: center;      /* 垂直居中 */
height: 100vh;            /* 占满整个视口高度 */
padding: 20px;
box-sizing: border-box;
}

.card-container {
width: 70%;               /* 你可以根据需要调整宽度 */
max-width: 700px;         /* 最大宽度 */
min-width: 350px;
box-shadow: 0 0 30px 1px rgba(0, 0, 0, 0.1);/* 添加投影效果 */
border-radius: 8px;       /* 圆角效果 */
}
</style>