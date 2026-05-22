export interface ApiResponse<T> {
  ok: boolean
  data: T
  error?: string
}

const BASE = '/api'

export async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  })

  const contentType = res.headers.get('content-type') || ''
  if (!contentType.includes('application/json')) {
    throw new Error(`Server error: ${res.status} ${res.statusText}`)
  }

  const data: ApiResponse<T> = await res.json()
  if (!data.ok) {
    throw new Error(data.error || 'Request failed')
  }
  return data.data
}
