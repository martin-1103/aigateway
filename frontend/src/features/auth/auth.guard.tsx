import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from './auth.store'
import type { Role } from './auth.types'

interface AuthGuardProps {
  children: React.ReactNode
}

export function AuthGuard({ children }: AuthGuardProps) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }

  return <>{children}</>
}

interface RoleGuardProps {
  children: React.ReactNode
  allowed: Role[]
}

export function RoleGuard({ children, allowed }: RoleGuardProps) {
  const user = useAuthStore((s) => s.user)

  if (!user || !allowed.includes(user.role)) {
    return <Navigate to="/dashboard" replace />
  }

  return <>{children}</>
}
