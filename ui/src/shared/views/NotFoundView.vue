<template>
  <div class="not-found-container">
    <el-result icon="warning" title="404 - Page Not Found"
      sub-title="Sorry, the page you are looking for does not exist or has been moved.">
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">Back to Home</el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 100vw;
  background-color: var(--el-bg-color-page);
}

:deep(.el-result__title p) {
  font-size: 24px;
  font-weight: 700;
}
</style>
