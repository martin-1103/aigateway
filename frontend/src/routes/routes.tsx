import { useEffect, useState } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/components/layout'
import { AuthGuard, RoleGuard, LoginPage, useAuthStore } from '@/features/auth'
import { getMe } from '@/features/auth/api/get-me.api'
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
import { LiteLayout, LiteAccountsPage, LiteAPIKeysPage } from '@/features/lite'

function AuthenticatedLayout() {
  const { user, logout, isAuthenticated, setAuth } = useAuthStore()
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isAuthenticated && !user) {
      setLoading(true)
      getMe()
        .then((meData) => {
          // Get token from store to pass to setAuth
          const state = useAuthStore.getState()
          if (state.token) {
            // Cast MeResponse to User (extra fields will be undefined)
            const user = meData as User
            setAuth(state.token, user)
          }
        })
        .catch(() => {
          // If fetching user fails, redirect to login
          useAuthStore.getState().logout()
        })
        .finally(() => {
          setLoading(false)
        })
    }
  }, [isAuthenticated, user, setAuth])

  if (!isAuthenticated) {
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
  {
    path: '/lite',
    element: <LiteLayout />,
    children: [
      {
        index: true,
        element: <LiteAccountsPage />,
      },
      {
        path: 'api-keys',
        element: <LiteAPIKeysPage />,
      },
    ],
  },
])
