import { useEffect, useState } from 'react'
import { createBrowserRouter, Navigate, useSearchParams } from 'react-router-dom'
import { AppLayout } from '@/components/layout'
import { AuthGuard, RoleGuard, LoginPage, useAuthStore } from '@/features/auth'
import { getMe, getMeWithAccessKey } from '@/features/auth/api/get-me.api'
import type { User } from '@/features/auth'
import { DashboardPage } from '@/features/dashboard'
import { UsersPage } from '@/features/users'
import { AccountsPage } from '@/features/accounts'
import { ProxiesPage } from '@/features/proxies'
import { ApiKeysPage } from '@/features/api-keys'
import { StatsPage } from '@/features/stats'
import { ModelMappingsPage } from '@/features/model-mappings'
import { OAuthPage } from '@/features/oauth'
import { SettingsPage } from '@/features/settings'

function AuthenticatedLayout() {
  const { user, logout, isAuthenticated, setAuth, setAccessKeyAuth, authMethod } = useAuthStore()
  const [loading, setLoading] = useState(false)
  const [searchParams, setSearchParams] = useSearchParams()

  useEffect(() => {
    const accessKey = searchParams.get('key')

    // If access key in URL, authenticate with it
    if (accessKey && accessKey.startsWith('uk_')) {
      setLoading(true)
      getMeWithAccessKey(accessKey)
        .then((meData) => {
          const user = meData as User
          setAccessKeyAuth(accessKey, user)
          // Remove key from URL after successful auth
          searchParams.delete('key')
          setSearchParams(searchParams, { replace: true })
        })
        .catch(() => {
          // Invalid access key
          searchParams.delete('key')
          setSearchParams(searchParams, { replace: true })
        })
        .finally(() => {
          setLoading(false)
        })
      return
    }

    // If authenticated but no user data, fetch it
    if (isAuthenticated && !user) {
      setLoading(true)
      getMe()
        .then((meData) => {
          const state = useAuthStore.getState()
          const user = meData as User
          if (state.accessKey) {
            setAccessKeyAuth(state.accessKey, user)
          } else if (state.token) {
            setAuth(state.token, user)
          }
        })
        .catch(() => {
          useAuthStore.getState().logout()
        })
        .finally(() => {
          setLoading(false)
        })
    }
  }, [isAuthenticated, user, setAuth, setAccessKeyAuth, searchParams, setSearchParams])

  if (!isAuthenticated && !searchParams.get('key')) {
    return <Navigate to="/login" replace />
  }

  if (!user || loading) {
    return <div className="min-h-screen bg-background" />
  }

  return (
    <AppLayout
      username={user.username}
      role={user.role}
      onLogout={logout}
      authMethod={authMethod}
    />
  )
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AuthenticatedLayout />
      </AuthGuard>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />,
      },
      {
        path: 'dashboard',
        element: <DashboardPage />,
      },
      {
        path: 'users',
        element: (
          <RoleGuard allowed={['admin']}>
            <UsersPage />
          </RoleGuard>
        ),
      },
      {
        path: 'accounts',
        element: (
          <RoleGuard allowed={['admin', 'user', 'provider']}>
            <AccountsPage />
          </RoleGuard>
        ),
      },
      {
        path: 'proxies',
        element: (
          <RoleGuard allowed={['admin']}>
            <ProxiesPage />
          </RoleGuard>
        ),
      },
      {
        path: 'api-keys',
        element: (
          <RoleGuard allowed={['admin', 'user']}>
            <ApiKeysPage />
          </RoleGuard>
        ),
      },
      {
        path: 'stats',
        element: (
          <RoleGuard allowed={['admin', 'user']}>
            <StatsPage />
          </RoleGuard>
        ),
      },
      {
        path: 'model-mappings',
        element: (
          <RoleGuard allowed={['admin', 'user']}>
            <ModelMappingsPage />
          </RoleGuard>
        ),
      },
      {
        path: 'oauth',
        element: (
          <RoleGuard allowed={['admin', 'user', 'provider']}>
            <OAuthPage />
          </RoleGuard>
        ),
      },
      {
        path: 'settings',
        element: <SettingsPage />,
      },
    ],
  },
])
