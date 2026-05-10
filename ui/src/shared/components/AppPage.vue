<template>
  <div class="ek-page" :class="{ 'scrollable': scrollable }">
    <div class="page-card floating-card">
      <!-- Header -->
      <header v-if="title || $slots.title || $slots.header || $slots['header-extra']" class="page-header"
        :style="headerPadding ? { padding: headerPadding } : undefined">
        <div class="header-main">
          <slot name="header">
            <slot name="title">
              <h1 v-if="title" class="page-title">{{ title }}</h1>
            </slot>
            <p v-if="subtitle" class="page-subtitle">{{ subtitle }}</p>
          </slot>
        </div>
        <div v-if="$slots['header-extra']" class="header-extra">
          <slot name="header-extra" />
        </div>
      </header>

      <!-- Content -->
      <div class="page-body" :class="[
        scrollable ? 'page-body-scroll' : 'page-body-static',
        { 'no-padding': noPadding }
      ]">
        <slot />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  subtitle?: string
  scrollable?: boolean
  noPadding?: boolean
  headerPadding?: string
}>()
</script>

<style scoped>
.ek-page {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.page-card {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.page-header {
  padding: 32px 40px 32px 40px;
  /* Balanced */
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.page-title {
  font-size: 24px;
  font-weight: 800;
  color: var(--el-text-color-primary);
  letter-spacing: -0.5px;
}

.page-subtitle {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.header-extra {
  display: flex;
  gap: 12px;
}

.page-body {
  flex: 1;
  min-height: 0;
  display: flex;
}

.page-body-scroll {
  flex-direction: column;
  overflow-y: auto;
  padding: 0 40px 40px 40px;
}

.page-body-static {
  flex-direction: column;
  padding: 0 40px 40px 40px;
}

.page-body.no-padding {
  padding: 0 !important;
}
</style>
