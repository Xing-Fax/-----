<script setup>
import Forget from './components/Forget.vue';
import CenterBox from './components/CenterBox.vue';
</script>

<template>
  <CenterBox>
    <div class="title-container">
      <img src="../src/assets/title.png" alt="logo">
      <p class="reset-text">重置密码</p>
    </div>
    <Forget></Forget>
  </CenterBox>
</template>

<style scoped>

img {
  height: 57px;  /* 保持图片的高度 */
  margin: 30px;
}

.title-container {
  margin-top: 60px;
  display: flex;            /* 使用 Flexbox */
  flex-direction: column;   /* 使子元素纵向排列 */
  justify-content: center;  /* 垂直居中 */
  align-items: center;      /* 水平居中 */
  text-align: center;       /* 确保文字居中 */
}

.reset-text {
  font-size: 25px;  /* 设置字体大小 */
  margin-top: 10px;  /* 图片和文本之间的间距 */
  margin-bottom: 50px;
}
</style>

<style>
/* 这将是全局样式 */
/* 强制覆盖 Ant Design Vue 组件的默认样式 */

/* 设置选择框的高度 */
.ant-select-selector {
  height: 40px !important;  /* 强制覆盖选择框的高度 */

}

.ant-select-selection-item{
  line-height: 37px !important
}

</style>
