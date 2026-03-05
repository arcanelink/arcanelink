import { create } from 'zustand'
import type { User } from '../types'
import { apiClient } from '../api/client'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  isInitialized: boolean
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, domain: string) => Promise<void>
  logout: () => void
  checkAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  isInitialized: false,

  login: async (username: string, password: string) => {
    const response = await apiClient.login({ username, password })
    const user = { user_id: response.user_id, username, domain: '' }

    // Save user info to localStorage
    console.log('Saving user to localStorage:', user)
    localStorage.setItem('user', JSON.stringify(user))
    console.log('User saved, verifying:', localStorage.getItem('user'))

    set({
      user,
      isAuthenticated: true,
    })
  },

  register: async (username: string, password: string, domain: string) => {
    const response = await apiClient.register({ username, password, domain })
    const user = { user_id: response.user_id, username, domain }

    // Save user info to localStorage
    localStorage.setItem('user', JSON.stringify(user))

    set({
      user,
      isAuthenticated: true,
    })
  },

  logout: () => {
    apiClient.clearToken()
    localStorage.removeItem('user')
    set({ user: null, isAuthenticated: false })
  },

  checkAuth: () => {
    console.log('checkAuth called')
    const token = apiClient.getToken()
    const userStr = localStorage.getItem('user')
    console.log('Token:', token ? 'exists' : 'missing')
    console.log('User data:', userStr ? 'exists' : 'missing')

    if (token && userStr) {
      try {
        const user = JSON.parse(userStr)
        console.log('Restoring user:', user)
        set({ user, isAuthenticated: true, isInitialized: true })
      } catch (error) {
        console.error('Failed to parse user data:', error)
        apiClient.clearToken()
        localStorage.removeItem('user')
        set({ user: null, isAuthenticated: false, isInitialized: true })
      }
    } else {
      console.log('Auth check failed: missing token or user data')
      set({ user: null, isAuthenticated: false, isInitialized: true })
    }
  },
}))
