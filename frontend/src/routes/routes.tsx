import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/components/layout'
import { AuthGuard, RoleGuard, LoginPage, useAuthStore } from '@/features/auth'
import { DashboardPage } from '@/features/dashboard'
import { UsersPage } from '@/features/users'
import { AccountsPage } from '@/features/accounts'
import { ProxiesPage } from '@/features/proxies'
import { ApiKeysPage } from '@/features/api-keys'
import { StatsPage } from '@/features/stats'
import { ModelMappingsPage } from '@/features/model-mappings'
import { OAuthPage } from '@/features/oauth'

function AuthenticatedLayout() {
  const { user, logout } = useAuthStore()

  if (!user) return null

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
    ],
  },
])
