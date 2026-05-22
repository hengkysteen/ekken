import { request } from '@shared/api/request'

export interface Profile {
  name: string
  pin_enabled: boolean
  updated_at?: string
  pin_updated_at?: string
  security_question?: string
}

export interface UpdateProfilePayload {
  name: string
  pin_enabled: boolean
  pin?: string
  security_question?: string
  security_answer?: string
}

export interface ResetPinPayload {
  answer: string
  new_pin: string
}

export const profileApi = {
  get: () => request<Profile>('/profile'),
  update: (payload: UpdateProfilePayload) =>
    request<Profile>('/profile', { method: 'PUT', body: JSON.stringify(payload) }),
  verifyPin: (pin: string) =>
    request<{ valid: boolean }>('/profile/pin/verify', { method: 'POST', body: JSON.stringify({ pin }) }),
  resetPin: (payload: ResetPinPayload) =>
    request<boolean>('/profile/pin/reset', { method: 'POST', body: JSON.stringify(payload) }),
}
