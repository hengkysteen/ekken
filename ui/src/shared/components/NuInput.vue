<template>
  <div class="nu-input-item">
    <label v-if="label" class="nu-input-label">
      {{ label }} <span v-if="required" class="nu-input-required">*</span>
    </label>
    <div class="nu-input-content">
      <slot></slot>
      <div v-if="notes" class="nu-input-notes" v-html="notes"></div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  label: { type: String, default: '' },
  notes: { type: String, default: '' },
  required: { type: Boolean, default: false },
})
</script>

<style scoped>
 
</style>
