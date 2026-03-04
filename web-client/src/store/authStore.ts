import { create } from 'zustand'
import type { User } from '../types'
import { apiClient } from '../api/client'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, domain: string) => Promise<void>
  logout: () => void
  checkAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  login: async (username: string, password: string) => {
    const response = await apiClient.login({ username, password })
    set({
      user: { user_id: response.user_id, username, domain: '' },
      isAuthenticated: true,
    })
  },

  register: async (username: string, password: string, domain: string) => {
    const response = await apiClient.register({ username, password, domain })
    set({
      user: { user_id: response.user_id, username, domain },
      isAuthenticated: true,
    })
  },

  logout: () => {
    apiClient.clearToken()
    set({ user: null, isAuthenticated: false })
  },

  checkAuth: () => {
    const token = apiClient.getToken()
    if (token) {
      set({ isAuthenticated: true })
    }
  },
}))
