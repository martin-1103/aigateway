import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/components/layout'
import { AuthGuard, RoleGuard, LoginPage, useAuthStore } from '@/features/auth'
import { DashboardPage } from '@/features/dashboard'
import { OAuthPage } from '@/features/oauth'
import { UsersPage } from '@/features/users'

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
        path: 'oauth',
        element: <OAuthPage />,
      },
      {
        path: 'users',
        element: (
          <RoleGuard allowed={['admin']}>
            <UsersPage />
          </RoleGuard>
        ),
      },
    ],
  },
])
