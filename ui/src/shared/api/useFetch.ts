import { createFetch } from '@vueuse/core'

export const useCustomFetch = createFetch({
  baseUrl: '/api',
  options: {
    async beforeFetch({ options }) {
      options.headers = {
        ...options.headers,
        'Content-Type': 'application/json',
      }
      return { options }
    },
    afterFetch(ctx) {
      // Ekstraksi data.data sesuai format API Ekken { ok: boolean, data: T, error?: string }
      const { data } = ctx
      
      if (data && typeof data === 'object' && 'ok' in data) {
        if (!data.ok) {
          // Melempar error agar ditangkap oleh state error useFetch
          throw new Error(data.error || 'API Error')
        }
        // Ganti data context dengan isi data yang sebenarnya
        ctx.data = data.data
      }
      return ctx
    },
  },
  fetchOptions: {
    mode: 'cors',
  },
})
