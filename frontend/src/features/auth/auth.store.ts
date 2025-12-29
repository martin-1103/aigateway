import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from './auth.types'

interface AuthState {
  token: string | null
  accessKey: string | null
  user: User | null
  isAuthenticated: boolean
  authMethod: 'jwt' | 'access_key' | null
  setAuth: (token: string, user: User) => void
  setAccessKeyAuth: (accessKey: string, user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      accessKey: null,
      user: null,
      isAuthenticated: false,
      authMethod: null,
      setAuth: (token, user) =>
        set({ token, accessKey: null, user, isAuthenticated: true, authMethod: 'jwt' }),
      setAccessKeyAuth: (accessKey, user) =>
        set({ token: null, accessKey, user, isAuthenticated: true, authMethod: 'access_key' }),
      logout: () =>
        set({ token: null, accessKey: null, user: null, isAuthenticated: false, authMethod: null }),
    }),
    { name: 'auth-storage' }
  )
)
