import { defineConfig } from 'vitest/config'
import { loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig(({ mode }) => {
  // Load all env variables including those without VITE_ prefix
  const env = loadEnv(mode, process.cwd(), '')
  const host = env.EKKENAPI_HOST || 'localhost'
  const port = env.EKKENAPI_PORT || '11245'

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@workflows': path.resolve(__dirname, './src/features/workflows'),
        '@assistant': path.resolve(__dirname, './src/features/assistant'),
        '@plugins': path.resolve(__dirname, './src/features/plugins'),
        '@credentials': path.resolve(__dirname, './src/features/credentials'),
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
    },
    test: {
      globals: true,
      environment: 'happy-dom',
    },
  }
})
