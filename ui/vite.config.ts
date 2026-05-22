import { defineConfig } from 'vitest/config'
import { loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig(({ mode }) => {
  // Load all env variables including those without VITE_ prefix
  const env = loadEnv(mode, process.cwd(), '')
  const host = env.EKKENAPI_HOST || 'localhost'
  const port = env.EKKENAPI_PORT || '11245'

  return {
    plugins: [
      vue(),
      AutoImport({
        resolvers: [ElementPlusResolver()],
      }),
      Components({
        dirs: ['src/shared/components', 'src/shared/form-fields'],
        resolvers: [ElementPlusResolver()],
      }),
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@workflows': path.resolve(__dirname, './src/features/workflows'),
        '@assistant': path.resolve(__dirname, './src/features/assistant'),
        '@plugins': path.resolve(__dirname, './src/features/plugins'),
        '@credentials': path.resolve(__dirname, './src/features/credentials'),
        '@profile': path.resolve(__dirname, './src/features/profile'),
        '@shared': path.resolve(__dirname, './src/shared'),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/api': {
          target: `http://${host}:${port}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: 'dist',
      chunkSizeWarningLimit: 1000,
      rollupOptions: {
        output: {
          manualChunks: {
            'vue-flow': ['@vue-flow/core', '@vue-flow/background', '@vue-flow/controls', '@vue-flow/minimap'],
            'vendor': ['vue', 'vue-router', 'pinia', '@vueuse/core'],
          },
        },
        onwarn(warning, warn) {
          // Suppress "PURE" annotation warnings from libraries like @vueuse
          if (warning.code === 'INVALID_ANNOTATION') return
          warn(warning)
        },
      },
    },
    test: {
      globals: true,
      environment: 'happy-dom',
    },
  }
})
