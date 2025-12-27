# Frontend Dashboard Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a role-based dashboard for AIGateway with Admin, User, and Provider views.

**Architecture:** React SPA with feature-based structure. Each feature is self-contained with api/, hooks/, components/. Zustand for auth state, TanStack Query for server state.

**Tech Stack:** React 18, Vite, TailwindCSS, shadcn/ui, Tremor, TanStack Query/Table, Zustand, Axios, Zod

---

## Execution Strategy

```
Batch 1: Project Setup (sequential)
    ↓
Batch 2: Core Lib (parallel: 4 tasks)
    ↓
┌───────────────────────┬────────────────────────┐
│ Batch 3: Auth Feature │ Batch 4: Layout        │
│ (sequential)          │ (parallel: 3 tasks)    │
└───────────────────────┴────────────────────────┘
    ↓
Batch 5: Shared Components (parallel: 4 tasks)
    ↓
Batch 6: Routes Setup (sequential)
    ↓
Batch 7-13: Features (parallel: 7 tasks)
```

---

## Batch 1: Project Setup

> **Execution:** Sequential - foundation for everything

### Task 1.1: Create Vite Project

**Files:**
- Create: `frontend/` directory with Vite React-TS template

**Step 1: Create project**

```bash
cd D:/temp/aigateway
npm create vite@latest frontend -- --template react-ts
```

**Step 2: Verify structure**

```bash
ls frontend/src
```

Expected: `App.tsx`, `main.tsx`, `App.css`, etc.

**Step 3: Commit**

```bash
cd frontend && git add -A && git commit -m "chore: scaffold vite react-ts project"
```

---

### Task 1.2: Install Dependencies

**Files:**
- Modify: `frontend/package.json`

**Step 1: Install core dependencies**

```bash
cd D:/temp/aigateway/frontend
npm install react-router-dom@6 @tanstack/react-query@5 @tanstack/react-table@8 react-hook-form@7 @hookform/resolvers@3 zod@3 zustand@4 axios@1 lucide-react clsx tailwind-merge class-variance-authority
```

**Step 2: Install Tremor**

```bash
npm install @tremor/react
```

**Step 3: Install dev dependencies**

```bash
npm install -D tailwindcss postcss autoprefixer @types/node
```

**Step 4: Commit**

```bash
git add -A && git commit -m "chore: install frontend dependencies"
```

---

### Task 1.3: Setup TailwindCSS

**Files:**
- Create: `frontend/tailwind.config.js`
- Create: `frontend/postcss.config.js`
- Modify: `frontend/src/index.css`

**Step 1: Init Tailwind**

```bash
cd D:/temp/aigateway/frontend
npx tailwindcss init -p
```

**Step 2: Configure tailwind.config.js**

```javascript
/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/@tremor/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

**Step 3: Replace src/index.css**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    --secondary: 210 40% 96.1%;
    --secondary-foreground: 222.2 47.4% 11.2%;
    --muted: 210 40% 96.1%;
    --muted-foreground: 215.4 16.3% 46.9%;
    --accent: 210 40% 96.1%;
    --accent-foreground: 222.2 47.4% 11.2%;
    --destructive: 0 84.2% 60.2%;
    --destructive-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%;
    --input: 214.3 31.8% 91.4%;
    --ring: 222.2 84% 4.9%;
    --radius: 0.5rem;
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;
    --primary: 210 40% 98%;
    --primary-foreground: 222.2 47.4% 11.2%;
    --secondary: 217.2 32.6% 17.5%;
    --secondary-foreground: 210 40% 98%;
    --muted: 217.2 32.6% 17.5%;
    --muted-foreground: 215 20.2% 65.1%;
    --accent: 217.2 32.6% 17.5%;
    --accent-foreground: 210 40% 98%;
    --destructive: 0 62.8% 30.6%;
    --destructive-foreground: 210 40% 98%;
    --border: 217.2 32.6% 17.5%;
    --input: 217.2 32.6% 17.5%;
    --ring: 212.7 26.8% 83.9%;
  }
}

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground;
  }
}
```

**Step 4: Commit**

```bash
git add -A && git commit -m "chore: setup tailwindcss with dark mode"
```

---

### Task 1.4: Setup shadcn/ui

**Files:**
- Create: `frontend/components.json`
- Create: `frontend/src/lib/utils.ts`
- Modify: `frontend/tsconfig.json`

**Step 1: Update tsconfig.json for path aliases**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

**Step 2: Update vite.config.ts**

```typescript
import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
})
```

**Step 3: Create components.json**

```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "default",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "src/index.css",
    "baseColor": "slate",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils"
  }
}
```

**Step 4: Create src/lib/utils.ts**

```typescript
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

**Step 5: Commit**

```bash
git add -A && git commit -m "chore: setup shadcn/ui configuration"
```

---

### Task 1.5: Install shadcn/ui Base Components

**Files:**
- Create: `frontend/src/components/ui/*.tsx`

**Step 1: Install essential components**

```bash
cd D:/temp/aigateway/frontend
npx shadcn@latest add button card dialog dropdown-menu input label table toast sonner
```

**Step 2: Verify components**

```bash
ls src/components/ui
```

Expected: `button.tsx`, `card.tsx`, `dialog.tsx`, etc.

**Step 3: Commit**

```bash
git add -A && git commit -m "chore: add shadcn/ui base components"
```

---

### Task 1.6: Create Environment Config

**Files:**
- Create: `frontend/.env.example`
- Create: `frontend/.env.local`

**Step 1: Create .env.example**

```bash
VITE_API_URL=http://localhost:8088
VITE_APP_NAME=AIGateway
```

**Step 2: Create .env.local (same content)**

```bash
VITE_API_URL=http://localhost:8088
VITE_APP_NAME=AIGateway
```

**Step 3: Commit**

```bash
git add .env.example && git commit -m "chore: add environment config template"
```

---

## Batch 2: Core Library

> **Execution:** Parallel - 4 independent tasks

### Task 2.1: API Client

**Files:**
- Create: `frontend/src/lib/api-client.ts`

**Step 1: Create api-client.ts**

```typescript
import axios from 'axios'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const authStorage = localStorage.getItem('auth-storage')
  if (authStorage) {
    const { state } = JSON.parse(authStorage)
    if (state?.token) {
      config.headers.Authorization = `Bearer ${state.token}`
    }
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth-storage')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(lib): add axios api client with auth interceptors"
```

---

### Task 2.2: Query Client

**Files:**
- Create: `frontend/src/lib/query-client.ts`

**Step 1: Create query-client.ts**

```typescript
import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
})
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(lib): add tanstack query client config"
```

---

### Task 2.3: Error Handler

**Files:**
- Create: `frontend/src/lib/handle-error.ts`

**Step 1: Create handle-error.ts**

```typescript
import { toast } from 'sonner'
import { AxiosError } from 'axios'

interface ApiError {
  error?: string
  message?: string
}

export function handleError(error: unknown): void {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ApiError | undefined
    const message = data?.error ?? data?.message ?? 'Request failed'
    toast.error(message)
    return
  }

  if (error instanceof Error) {
    toast.error(error.message)
    return
  }

  toast.error('Something went wrong')
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(lib): add centralized error handler"
```

---

### Task 2.4: Utility Functions

**Files:**
- Modify: `frontend/src/lib/utils.ts`

**Step 1: Add utility functions to utils.ts**

```typescript
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(date: string | Date, format = 'PP'): string {
  const d = new Date(date)
  return d.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(date: string | Date): string {
  const d = new Date(date)
  return d.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(lib): add date formatting utilities"
```

---

## Batch 3: Auth Feature

> **Execution:** Sequential - auth is foundational

### Task 3.1: Auth Types

**Files:**
- Create: `frontend/src/features/auth/auth.types.ts`

**Step 1: Create auth.types.ts**

```typescript
export type Role = 'admin' | 'user' | 'provider'

export interface User {
  id: string
  username: string
  role: Role
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(auth): add auth types"
```

---

### Task 3.2: Auth Store

**Files:**
- Create: `frontend/src/features/auth/auth.store.ts`

**Step 1: Create auth.store.ts**

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from './auth.types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  setAuth: (token: string, user: User) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      setAuth: (token, user) =>
        set({ token, user, isAuthenticated: true }),
      logout: () =>
        set({ token: null, user: null, isAuthenticated: false }),
    }),
    { name: 'auth-storage' }
  )
)
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(auth): add zustand auth store"
```

---

### Task 3.3: Auth API Functions

**Files:**
- Create: `frontend/src/features/auth/api/login.api.ts`
- Create: `frontend/src/features/auth/api/get-me.api.ts`
- Create: `frontend/src/features/auth/api/change-password.api.ts`
- Create: `frontend/src/features/auth/api/index.ts`

**Step 1: Create login.api.ts**

```typescript
import { apiClient } from '@/lib/api-client'
import type { LoginRequest, LoginResponse } from '../auth.types'

export async function login(data: LoginRequest): Promise<LoginResponse> {
  const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', data)
  return response.data
}
```

**Step 2: Create get-me.api.ts**

```typescript
import { apiClient } from '@/lib/api-client'
import type { User } from '../auth.types'

export async function getMe(): Promise<User> {
  const response = await apiClient.get<User>('/api/v1/auth/me')
  return response.data
}
```

**Step 3: Create change-password.api.ts**

```typescript
import { apiClient } from '@/lib/api-client'
import type { ChangePasswordRequest } from '../auth.types'

export async function changePassword(data: ChangePasswordRequest): Promise<void> {
  await apiClient.put('/api/v1/auth/password', data)
}
```

**Step 4: Create index.ts**

```typescript
export { login } from './login.api'
export { getMe } from './get-me.api'
export { changePassword } from './change-password.api'
```

**Step 5: Commit**

```bash
git add -A && git commit -m "feat(auth): add auth api functions"
```

---

### Task 3.4: Auth Hooks

**Files:**
- Create: `frontend/src/features/auth/hooks/use-login.ts`
- Create: `frontend/src/features/auth/hooks/use-change-password.ts`
- Create: `frontend/src/features/auth/hooks/index.ts`

**Step 1: Create use-login.ts**

```typescript
import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { login } from '../api'
import { useAuthStore } from '../auth.store'
import { handleError } from '@/lib/handle-error'

export function useLogin() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  return useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setAuth(data.token, data.user)
      toast.success('Login successful')
      navigate('/dashboard')
    },
    onError: handleError,
  })
}
```

**Step 2: Create use-change-password.ts**

```typescript
import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { changePassword } from '../api'
import { handleError } from '@/lib/handle-error'

export function useChangePassword() {
  return useMutation({
    mutationFn: changePassword,
    onSuccess: () => {
      toast.success('Password changed successfully')
    },
    onError: handleError,
  })
}
```

**Step 3: Create index.ts**

```typescript
export { useLogin } from './use-login'
export { useChangePassword } from './use-change-password'
```

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(auth): add auth mutation hooks"
```

---

### Task 3.5: Auth Guard

**Files:**
- Create: `frontend/src/features/auth/auth.guard.tsx`

**Step 1: Create auth.guard.tsx**

```typescript
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
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(auth): add auth and role guards"
```

---

### Task 3.6: Login Form Component

**Files:**
- Create: `frontend/src/features/auth/components/login-form.tsx`

**Step 1: Create login-form.tsx**

```typescript
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useLogin } from '../hooks'

const loginSchema = z.object({
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
})

type LoginFormData = z.infer<typeof loginSchema>

export function LoginForm() {
  const { mutate, isPending } = useLogin()
  const { register, handleSubmit, formState: { errors } } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = (data: LoginFormData) => mutate(data)

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="username">Username</Label>
        <Input id="username" {...register('username')} />
        {errors.username && (
          <p className="text-sm text-destructive">{errors.username.message}</p>
        )}
      </div>
      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <Input id="password" type="password" {...register('password')} />
        {errors.password && (
          <p className="text-sm text-destructive">{errors.password.message}</p>
        )}
      </div>
      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Signing in...' : 'Sign in'}
      </Button>
    </form>
  )
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(auth): add login form component"
```

---

### Task 3.7: Login Page

**Files:**
- Create: `frontend/src/features/auth/login-page.tsx`

**Step 1: Create login-page.tsx**

```typescript
import { Navigate } from 'react-router-dom'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { LoginForm } from './components/login-form'
import { useAuthStore } from './auth.store'

export function LoginPage() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">AIGateway</CardTitle>
          <CardDescription>Sign in to your account</CardDescription>
        </CardHeader>
        <CardContent>
          <LoginForm />
        </CardContent>
      </Card>
    </div>
  )
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(auth): add login page"
```

---

### Task 3.8: Auth Feature Index

**Files:**
- Create: `frontend/src/features/auth/index.ts`
- Create: `frontend/src/features/auth/components/index.ts`

**Step 1: Create components/index.ts**

```typescript
export { LoginForm } from './login-form'
```

**Step 2: Create feature index.ts**

```typescript
export { LoginPage } from './login-page'
export { AuthGuard, RoleGuard } from './auth.guard'
export { useAuthStore } from './auth.store'
export { useLogin, useChangePassword } from './hooks'
export type { User, Role, LoginRequest } from './auth.types'
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(auth): add auth feature exports"
```

---

## Batch 4: Layout Components

> **Execution:** Parallel - 3 independent component groups

### Task 4.1: Sidebar Components

**Files:**
- Create: `frontend/src/components/sidebar/sidebar.store.ts`
- Create: `frontend/src/components/sidebar/sidebar.constants.ts`
- Create: `frontend/src/components/sidebar/sidebar-item.tsx`
- Create: `frontend/src/components/sidebar/sidebar-nav.tsx`
- Create: `frontend/src/components/sidebar/sidebar.tsx`
- Create: `frontend/src/components/sidebar/index.ts`

**Step 1: Create sidebar.store.ts**

```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface SidebarState {
  isCollapsed: boolean
  toggle: () => void
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      isCollapsed: false,
      toggle: () => set((s) => ({ isCollapsed: !s.isCollapsed })),
    }),
    { name: 'sidebar-storage' }
  )
)
```

**Step 2: Create sidebar.constants.ts**

```typescript
import { BarChart3, Database, GitBranch, Key, Server, Shield, Users } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { Role } from '@/features/auth'

export interface NavItem {
  label: string
  href: string
  icon: LucideIcon
  roles: Role[]
}

export const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: BarChart3, roles: ['admin', 'user', 'provider'] },
  { label: 'Users', href: '/users', icon: Users, roles: ['admin'] },
  { label: 'Accounts', href: '/accounts', icon: Database, roles: ['admin', 'user', 'provider'] },
  { label: 'Proxies', href: '/proxies', icon: Server, roles: ['admin'] },
  { label: 'API Keys', href: '/api-keys', icon: Key, roles: ['admin', 'user'] },
  { label: 'Stats', href: '/stats', icon: BarChart3, roles: ['admin', 'user'] },
  { label: 'Model Mappings', href: '/model-mappings', icon: GitBranch, roles: ['admin', 'user'] },
  { label: 'OAuth', href: '/oauth', icon: Shield, roles: ['admin', 'user', 'provider'] },
]
```

**Step 3: Create sidebar-item.tsx**

```typescript
import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useSidebarStore } from './sidebar.store'
import type { NavItem } from './sidebar.constants'

interface SidebarItemProps {
  item: NavItem
}

export function SidebarItem({ item }: SidebarItemProps) {
  const isCollapsed = useSidebarStore((s) => s.isCollapsed)
  const Icon = item.icon

  return (
    <NavLink
      to={item.href}
      className={({ isActive }) =>
        cn(
          'flex items-center gap-3 rounded-lg px-3 py-2 transition-colors',
          'hover:bg-accent hover:text-accent-foreground',
          isActive ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'
        )
      }
    >
      <Icon className="h-4 w-4 shrink-0" />
      {!isCollapsed && <span>{item.label}</span>}
    </NavLink>
  )
}
```

**Step 4: Create sidebar-nav.tsx**

```typescript
import { useAuthStore } from '@/features/auth'
import { NAV_ITEMS } from './sidebar.constants'
import { SidebarItem } from './sidebar-item'

export function SidebarNav() {
  const user = useAuthStore((s) => s.user)

  const visibleItems = NAV_ITEMS.filter(
    (item) => user && item.roles.includes(user.role)
  )

  return (
    <nav className="flex flex-col gap-1 p-2">
      {visibleItems.map((item) => (
        <SidebarItem key={item.href} item={item} />
      ))}
    </nav>
  )
}
```

**Step 5: Create sidebar.tsx**

```typescript
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useSidebarStore } from './sidebar.store'
import { SidebarNav } from './sidebar-nav'

export function Sidebar() {
  const { isCollapsed, toggle } = useSidebarStore()

  return (
    <aside
      className={cn(
        'flex flex-col border-r bg-card transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-64'
      )}
    >
      <div className="flex h-14 items-center justify-between border-b px-4">
        {!isCollapsed && <span className="font-semibold">AIGateway</span>}
        <Button variant="ghost" size="icon" onClick={toggle}>
          {isCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </Button>
      </div>
      <SidebarNav />
    </aside>
  )
}
```

**Step 6: Create index.ts**

```typescript
export { Sidebar } from './sidebar'
export { useSidebarStore } from './sidebar.store'
```

**Step 7: Commit**

```bash
git add -A && git commit -m "feat(layout): add collapsible sidebar components"
```

---

### Task 4.2: Header Components

**Files:**
- Create: `frontend/src/components/header/header-theme-toggle.tsx`
- Create: `frontend/src/components/header/header-user-menu.tsx`
- Create: `frontend/src/components/header/header.tsx`
- Create: `frontend/src/components/header/index.ts`

**Step 1: Create header-theme-toggle.tsx**

```typescript
import { Moon, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'

export function HeaderThemeToggle() {
  const [isDark, setIsDark] = useState(true)

  useEffect(() => {
    document.documentElement.classList.toggle('dark', isDark)
  }, [isDark])

  return (
    <Button variant="ghost" size="icon" onClick={() => setIsDark(!isDark)}>
      {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </Button>
  )
}
```

**Step 2: Create header-user-menu.tsx**

```typescript
import { LogOut, User } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useAuthStore } from '@/features/auth'

export function HeaderUserMenu() {
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon">
          <User className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuLabel>
          {user?.username}
          <span className="block text-xs font-normal text-muted-foreground">
            {user?.role}
          </span>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={handleLogout}>
          <LogOut className="mr-2 h-4 w-4" />
          Logout
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
```

**Step 3: Create header.tsx**

```typescript
import { HeaderThemeToggle } from './header-theme-toggle'
import { HeaderUserMenu } from './header-user-menu'

export function Header() {
  return (
    <header className="flex h-14 items-center justify-end gap-2 border-b bg-card px-4">
      <HeaderThemeToggle />
      <HeaderUserMenu />
    </header>
  )
}
```

**Step 4: Create index.ts**

```typescript
export { Header } from './header'
```

**Step 5: Commit**

```bash
git add -A && git commit -m "feat(layout): add header with theme toggle and user menu"
```

---

### Task 4.3: App Layout

**Files:**
- Create: `frontend/src/components/layout/app-layout.tsx`
- Create: `frontend/src/components/layout/index.ts`

**Step 1: Create app-layout.tsx**

```typescript
import { Outlet } from 'react-router-dom'
import { Sidebar } from '@/components/sidebar'
import { Header } from '@/components/header'

export function AppLayout() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header />
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
```

**Step 2: Create index.ts**

```typescript
export { AppLayout } from './app-layout'
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(layout): add app layout with sidebar and header"
```

---

## Batch 5: Shared Components

> **Execution:** Parallel - 4 independent component groups

### Task 5.1: Page Components

**Files:**
- Create: `frontend/src/components/page/page-header.tsx`
- Create: `frontend/src/components/page/page-content.tsx`
- Create: `frontend/src/components/page/index.ts`

**Step 1: Create page-header.tsx**

```typescript
interface PageHeaderProps {
  title: string
  description?: string
  actions?: React.ReactNode
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <div className="flex items-center justify-between">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{title}</h1>
        {description && (
          <p className="text-muted-foreground">{description}</p>
        )}
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}
```

**Step 2: Create page-content.tsx**

```typescript
import { cn } from '@/lib/utils'

interface PageContentProps {
  children: React.ReactNode
  className?: string
}

export function PageContent({ children, className }: PageContentProps) {
  return (
    <div className={cn('mt-6 space-y-6', className)}>
      {children}
    </div>
  )
}
```

**Step 3: Create index.ts**

```typescript
export { PageHeader } from './page-header'
export { PageContent } from './page-content'
```

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(components): add page header and content components"
```

---

### Task 5.2: Feedback Components

**Files:**
- Create: `frontend/src/components/feedback/loading-spinner.tsx`
- Create: `frontend/src/components/feedback/empty-state.tsx`
- Create: `frontend/src/components/feedback/index.ts`

**Step 1: Create loading-spinner.tsx**

```typescript
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface LoadingSpinnerProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizeClasses = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-8 w-8',
}

export function LoadingSpinner({ className, size = 'md' }: LoadingSpinnerProps) {
  return (
    <Loader2 className={cn('animate-spin', sizeClasses[size], className)} />
  )
}
```

**Step 2: Create empty-state.tsx**

```typescript
import type { LucideIcon } from 'lucide-react'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  action?: React.ReactNode
}

export function EmptyState({ icon: Icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <Icon className="h-12 w-12 text-muted-foreground" />
      <h3 className="mt-4 text-lg font-semibold">{title}</h3>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  )
}
```

**Step 3: Create index.ts**

```typescript
export { LoadingSpinner } from './loading-spinner'
export { EmptyState } from './empty-state'
```

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(components): add loading spinner and empty state"
```

---

### Task 5.3: Chart Components

**Files:**
- Create: `frontend/src/components/charts/stat-card.tsx`
- Create: `frontend/src/components/charts/area-chart-card.tsx`
- Create: `frontend/src/components/charts/index.ts`

**Step 1: Create stat-card.tsx**

```typescript
import type { LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  description?: string
}

export function StatCard({ title, value, icon: Icon, description }: StatCardProps) {
  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <p className="text-sm font-medium text-muted-foreground">{title}</p>
          <Icon className="h-4 w-4 text-muted-foreground" />
        </div>
        <div className="mt-2">
          <p className="text-2xl font-bold">{value}</p>
          {description && (
            <p className="text-xs text-muted-foreground">{description}</p>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
```

**Step 2: Create area-chart-card.tsx**

```typescript
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AreaChart } from '@tremor/react'

interface AreaChartCardProps {
  title: string
  data: Record<string, unknown>[]
  index: string
  categories: string[]
  colors?: string[]
}

export function AreaChartCard({ title, data, index, categories, colors }: AreaChartCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <AreaChart
          data={data}
          index={index}
          categories={categories}
          colors={colors ?? ['blue', 'emerald']}
          showLegend
          showGridLines={false}
          className="h-72"
        />
      </CardContent>
    </Card>
  )
}
```

**Step 3: Create index.ts**

```typescript
export { StatCard } from './stat-card'
export { AreaChartCard } from './area-chart-card'
```

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(components): add stat card and area chart components"
```

---

### Task 5.4: Form Components

**Files:**
- Create: `frontend/src/components/form/form-field.tsx`
- Create: `frontend/src/components/form/form-dialog.tsx`
- Create: `frontend/src/components/form/index.ts`

**Step 1: Create form-field.tsx**

```typescript
import { Label } from '@/components/ui/label'

interface FormFieldProps {
  label: string
  error?: string
  children: React.ReactNode
}

export function FormField({ label, error, children }: FormFieldProps) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
```

**Step 2: Create form-dialog.tsx**

```typescript
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface FormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: React.ReactNode
  isSubmitting?: boolean
  onSubmit: (e: React.FormEvent) => void
  submitLabel?: string
}

export function FormDialog({
  open,
  onOpenChange,
  title,
  description,
  children,
  isSubmitting,
  onSubmit,
  submitLabel = 'Save',
}: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={onSubmit}>
          <DialogHeader>
            <DialogTitle>{title}</DialogTitle>
            {description && <DialogDescription>{description}</DialogDescription>}
          </DialogHeader>
          <div className="space-y-4 py-4">{children}</div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : submitLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
```

**Step 3: Create index.ts**

```typescript
export { FormField } from './form-field'
export { FormDialog } from './form-dialog'
```

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(components): add form field and dialog components"
```

---

## Batch 6: Routes Setup

> **Execution:** Sequential - connects everything

### Task 6.1: Route Definitions

**Files:**
- Create: `frontend/src/routes/routes.tsx`

**Step 1: Create routes.tsx**

```typescript
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AppLayout } from '@/components/layout'
import { AuthGuard, RoleGuard, LoginPage } from '@/features/auth'

// Lazy load pages
import { DashboardPage } from '@/features/dashboard'

export const router = createBrowserRouter([
  {
    path: '/login',
    element: <LoginPage />,
  },
  {
    path: '/',
    element: (
      <AuthGuard>
        <AppLayout />
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
    ],
  },
])
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(routes): add basic route configuration"
```

---

### Task 6.2: Dashboard Feature

**Files:**
- Create: `frontend/src/features/dashboard/dashboard-page.tsx`
- Create: `frontend/src/features/dashboard/index.ts`

**Step 1: Create dashboard-page.tsx**

```typescript
import { Database, Key, Server, Users } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { StatCard } from '@/components/charts'
import { useAuthStore } from '@/features/auth'

export function DashboardPage() {
  const user = useAuthStore((s) => s.user)

  return (
    <>
      <PageHeader
        title="Dashboard"
        description={`Welcome back, ${user?.username}`}
      />
      <PageContent>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard title="Users" value="--" icon={Users} />
          <StatCard title="Accounts" value="--" icon={Database} />
          <StatCard title="Proxies" value="--" icon={Server} />
          <StatCard title="API Keys" value="--" icon={Key} />
        </div>
      </PageContent>
    </>
  )
}
```

**Step 2: Create index.ts**

```typescript
export { DashboardPage } from './dashboard-page'
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(dashboard): add dashboard page with stat cards"
```

---

### Task 6.3: App Entry Point

**Files:**
- Modify: `frontend/src/App.tsx`
- Modify: `frontend/src/main.tsx`

**Step 1: Update App.tsx**

```typescript
import { RouterProvider } from 'react-router-dom'
import { QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import { queryClient } from '@/lib/query-client'
import { router } from '@/routes/routes'

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <Toaster position="top-right" theme="dark" richColors closeButton />
    </QueryClientProvider>
  )
}
```

**Step 2: Update main.tsx**

```typescript
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(app): wire up app with router, query client, and toaster"
```

---

## Batch 7-13: Feature Modules

> **Execution:** Parallel - each feature is independent

Each feature follows the same pattern. I'll detail Users as the template, others follow identically.

### Task 7: Users Feature (Admin only)

**Files to create:**
- `frontend/src/features/users/users.types.ts`
- `frontend/src/features/users/users.schemas.ts`
- `frontend/src/features/users/api/get-users.api.ts`
- `frontend/src/features/users/api/create-user.api.ts`
- `frontend/src/features/users/api/update-user.api.ts`
- `frontend/src/features/users/api/delete-user.api.ts`
- `frontend/src/features/users/api/index.ts`
- `frontend/src/features/users/hooks/use-users-query.ts`
- `frontend/src/features/users/hooks/use-create-user.ts`
- `frontend/src/features/users/hooks/use-update-user.ts`
- `frontend/src/features/users/hooks/use-delete-user.ts`
- `frontend/src/features/users/hooks/index.ts`
- `frontend/src/features/users/components/users-table.tsx`
- `frontend/src/features/users/components/users-table-columns.tsx`
- `frontend/src/features/users/components/user-create-dialog.tsx`
- `frontend/src/features/users/components/user-edit-dialog.tsx`
- `frontend/src/features/users/components/user-delete-dialog.tsx`
- `frontend/src/features/users/components/index.ts`
- `frontend/src/features/users/users-page.tsx`
- `frontend/src/features/users/index.ts`

**Pattern:** Same as auth feature with CRUD operations. See design doc for exact code patterns.

---

### Task 8: Accounts Feature (Admin, User, Provider)

Same pattern as Users. Endpoint: `/api/v1/accounts`

---

### Task 9: Proxies Feature (Admin only)

Same pattern as Users. Endpoint: `/api/v1/proxies`

---

### Task 10: API Keys Feature (Admin, User)

Same pattern as Users. Endpoint: `/api/v1/api-keys`

---

### Task 11: Stats Feature (Admin, User)

Different pattern - read-only with charts. Endpoint: `/api/v1/stats`

---

### Task 12: Model Mappings Feature (Admin, User)

Same pattern as Users. Endpoint: `/api/v1/model-mappings`

---

### Task 13: OAuth Feature (Admin, User, Provider)

Different pattern - OAuth flow management. Endpoint: `/api/v1/oauth`

---

## Final Tasks

### Task 14: Update Routes with All Features

Add all feature routes to `routes.tsx` with proper role guards.

### Task 15: Verify Build

```bash
cd D:/temp/aigateway/frontend
npm run build
```

Expected: Build succeeds with no errors.

### Task 16: Test Run

```bash
npm run dev
```

Expected: App runs on localhost:5173, login works, dashboard loads.

---

## Verification Checklist

- [ ] All files created per structure
- [ ] `npm run build` succeeds
- [ ] Login flow works
- [ ] Role-based sidebar filtering works
- [ ] Dashboard loads for each role
- [ ] CRUD operations work for each feature
- [ ] Dark mode toggle works
- [ ] Sidebar collapse works
