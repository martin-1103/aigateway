# Frontend Dashboard Design

## Overview

Single-page dashboard application for AIGateway with role-based views for Admin, User, and Provider roles.

## Tech Stack

| Category | Choice |
|----------|--------|
| Framework | React 18 + Vite |
| Styling | TailwindCSS |
| UI Components | shadcn/ui |
| Charts | Tremor |
| Data Fetching | TanStack Query v5 |
| Tables | TanStack Table v8 |
| Forms | React Hook Form + Zod |
| State | Zustand |
| HTTP Client | Axios |
| Icons | Lucide React |

## Design Decisions

- **Dark-first + Minimalism** - Developer-focused, less eye strain, charts pop better
- **Collapsible sidebar** - Modern, space-efficient
- **Feature-based structure** - Scalable, colocated files
- **Small files (max 150 lines)** - AI-friendly, single responsibility
- **No separate settings page** - Password change via modal, theme toggle in header

---

## Role-Based Access

| Route | Admin | User | Provider |
|-------|-------|------|----------|
| `/dashboard` | ✅ | ✅ | ✅ |
| `/users` | ✅ | ❌ | ❌ |
| `/accounts` | ✅ | ✅ | ✅ |
| `/proxies` | ✅ | ❌ | ❌ |
| `/api-keys` | ✅ | ✅ | ❌ |
| `/stats` | ✅ | ✅ | ❌ |
| `/model-mappings` | ✅ | ✅ | ❌ |
| `/oauth` | ✅ | ✅ | ✅ |

---

## Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ui/                    # shadcn/ui components
│   │   ├── data-table/
│   │   │   ├── data-table.tsx
│   │   │   ├── data-table-pagination.tsx
│   │   │   ├── data-table-toolbar.tsx
│   │   │   ├── data-table-column-header.tsx
│   │   │   ├── data-table-row-actions.tsx
│   │   │   ├── data-table.types.ts
│   │   │   └── index.ts
│   │   ├── layout/
│   │   │   ├── app-layout.tsx
│   │   │   ├── app-layout.types.ts
│   │   │   └── index.ts
│   │   ├── sidebar/
│   │   │   ├── sidebar.tsx
│   │   │   ├── sidebar-nav.tsx
│   │   │   ├── sidebar-item.tsx
│   │   │   ├── sidebar.store.ts
│   │   │   ├── sidebar.constants.ts
│   │   │   └── index.ts
│   │   ├── header/
│   │   │   ├── header.tsx
│   │   │   ├── header-user-menu.tsx
│   │   │   ├── header-theme-toggle.tsx
│   │   │   └── index.ts
│   │   ├── page/
│   │   │   ├── page-header.tsx
│   │   │   ├── page-content.tsx
│   │   │   └── index.ts
│   │   ├── charts/
│   │   │   ├── stat-card.tsx
│   │   │   ├── stat-card-grid.tsx
│   │   │   ├── area-chart-card.tsx
│   │   │   ├── bar-chart-card.tsx
│   │   │   ├── donut-chart-card.tsx
│   │   │   └── index.ts
│   │   ├── form/
│   │   │   ├── form-field.tsx
│   │   │   ├── form-select.tsx
│   │   │   ├── form-switch.tsx
│   │   │   ├── form-dialog.tsx
│   │   │   └── index.ts
│   │   └── feedback/
│   │       ├── toast-provider.tsx
│   │       ├── error-boundary.tsx
│   │       ├── loading-spinner.tsx
│   │       ├── empty-state.tsx
│   │       └── index.ts
│   │
│   ├── features/
│   │   ├── auth/
│   │   │   ├── api/
│   │   │   │   ├── login.api.ts
│   │   │   │   ├── get-me.api.ts
│   │   │   │   ├── change-password.api.ts
│   │   │   │   └── index.ts
│   │   │   ├── hooks/
│   │   │   │   ├── use-login.ts
│   │   │   │   ├── use-me-query.ts
│   │   │   │   ├── use-change-password.ts
│   │   │   │   └── index.ts
│   │   │   ├── components/
│   │   │   │   ├── login-form.tsx
│   │   │   │   ├── password-change-dialog.tsx
│   │   │   │   ├── user-menu.tsx
│   │   │   │   └── index.ts
│   │   │   ├── auth.store.ts
│   │   │   ├── auth.guard.tsx
│   │   │   ├── auth.types.ts
│   │   │   ├── login-page.tsx
│   │   │   ├── CLAUDE.md
│   │   │   └── index.ts
│   │   │
│   │   ├── dashboard/
│   │   │   ├── components/
│   │   │   │   ├── admin-dashboard.tsx
│   │   │   │   ├── user-dashboard.tsx
│   │   │   │   ├── provider-dashboard.tsx
│   │   │   │   └── index.ts
│   │   │   ├── dashboard-page.tsx
│   │   │   ├── CLAUDE.md
│   │   │   └── index.ts
│   │   │
│   │   ├── users/
│   │   │   ├── api/
│   │   │   │   ├── get-users.api.ts
│   │   │   │   ├── create-user.api.ts
│   │   │   │   ├── update-user.api.ts
│   │   │   │   ├── delete-user.api.ts
│   │   │   │   └── index.ts
│   │   │   ├── hooks/
│   │   │   │   ├── use-users-query.ts
│   │   │   │   ├── use-create-user.ts
│   │   │   │   ├── use-update-user.ts
│   │   │   │   ├── use-delete-user.ts
│   │   │   │   └── index.ts
│   │   │   ├── components/
│   │   │   │   ├── users-table.tsx
│   │   │   │   ├── users-table-columns.tsx
│   │   │   │   ├── users-table-actions.tsx
│   │   │   │   ├── users-table-toolbar.tsx
│   │   │   │   ├── user-form-fields.tsx
│   │   │   │   ├── user-create-dialog.tsx
│   │   │   │   ├── user-edit-dialog.tsx
│   │   │   │   ├── user-delete-dialog.tsx
│   │   │   │   └── index.ts
│   │   │   ├── users.types.ts
│   │   │   ├── users.schemas.ts
│   │   │   ├── users.constants.ts
│   │   │   ├── users-page.tsx
│   │   │   ├── CLAUDE.md
│   │   │   └── index.ts
│   │   │
│   │   ├── accounts/        # Same pattern as users
│   │   ├── proxies/         # Same pattern as users
│   │   ├── api-keys/        # Same pattern as users
│   │   ├── stats/           # Stats + logs specific
│   │   ├── model-mappings/  # Same pattern as users
│   │   └── oauth/           # OAuth flow specific
│   │
│   ├── lib/
│   │   ├── api-client.ts
│   │   ├── query-client.ts
│   │   ├── handle-error.ts
│   │   └── utils.ts
│   │
│   ├── routes/
│   │   ├── routes.tsx
│   │   ├── route-guards.tsx
│   │   ├── admin.routes.tsx
│   │   ├── user.routes.tsx
│   │   └── provider.routes.tsx
│   │
│   ├── types/
│   │   └── api.d.ts          # Generated from OpenAPI
│   │
│   ├── App.tsx
│   ├── main.tsx
│   ├── index.css
│   └── CLAUDE.md
│
├── scripts/
│   └── generate-types.ts
│
├── .env.example
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
├── components.json
├── package.json
└── CLAUDE.md
```

---

## File Size Guidelines

| File Type | Max Lines | Responsibility |
|-----------|-----------|----------------|
| `*.api.ts` | ~20 | Single API call |
| `use-*.ts` | ~30 | Single hook |
| `*-table.tsx` | ~50 | Table wrapper |
| `*-columns.tsx` | ~60 | Column definitions |
| `*-form-fields.tsx` | ~50 | Form inputs only |
| `*-dialog.tsx` | ~40 | Modal shell + form |
| `*-page.tsx` | ~40 | Layout composition |
| `*.types.ts` | ~30 | Types/interfaces |
| `*.schemas.ts` | ~20 | Zod validation |
| `*.constants.ts` | ~20 | Static config |
| `*.store.ts` | ~30 | Zustand store |

---

## Auth Flow

```
Login Page
    ↓
POST /api/v1/auth/login → JWT token
    ↓
Store token (Zustand + localStorage)
    ↓
GET /api/v1/auth/me → User data (role)
    ↓
Redirect to /dashboard
    ↓
Sidebar renders based on role
```

### Auth Store

```typescript
interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  setAuth: (token: string, user: User) => void
  logout: () => void
}
```

### API Client Interceptors

- Request: Auto-attach Bearer token
- Response: Redirect to /login on 401

---

## Dashboard Views

### Admin Dashboard
- Summary cards: Total users, accounts, proxies, requests today
- Chart: Request trend (7 days)
- Table: Recent errors/failed requests
- Quick actions: Add user, add account

### User Dashboard
- Summary cards: API keys, accounts, model mappings, requests today
- Chart: Usage trend (7 days)
- Table: Recent requests
- Quick actions: Create API key, add mapping

### Provider Dashboard
- Summary cards: Accounts managed, OAuth tokens active
- Table: Account status overview
- Quick actions: Add account, refresh tokens

---

## Sidebar Navigation

```typescript
const NAV_ITEMS = [
  { label: 'Dashboard', href: '/dashboard', icon: BarChart,
    roles: ['admin', 'user', 'provider'] },
  { label: 'Users', href: '/users', icon: Users,
    roles: ['admin'] },
  { label: 'Accounts', href: '/accounts', icon: Database,
    roles: ['admin', 'user', 'provider'] },
  { label: 'Proxies', href: '/proxies', icon: Server,
    roles: ['admin'] },
  { label: 'API Keys', href: '/api-keys', icon: Key,
    roles: ['admin', 'user'] },
  { label: 'Stats', href: '/stats', icon: BarChart,
    roles: ['admin', 'user'] },
  { label: 'Model Mappings', href: '/model-mappings', icon: GitBranch,
    roles: ['admin', 'user'] },
  { label: 'OAuth', href: '/oauth', icon: Shield,
    roles: ['admin', 'user', 'provider'] },
]
```

---

## Component Patterns

### Data Table

Reusable table with:
- Server-side pagination
- Sortable columns
- Row actions (edit/delete)
- Toolbar (search + filters + add button)
- Loading & empty states

### Form Dialog

Reusable dialog shell:
- Title + description
- Form content slot
- Cancel + Submit buttons
- Loading state on submit

### Stat Card

Single metric display:
- Title + icon
- Large value
- Optional delta badge (% change)

### Chart Cards

Tremor chart wrappers:
- AreaChartCard - trends over time
- BarChartCard - comparisons
- DonutChartCard - distribution

---

## Error Handling

- Centralized `handleError()` function
- Axios interceptor for 401 → logout
- Sonner toast for notifications
- React Error Boundary for crashes
- Empty state component for no data

---

## Dependencies

```json
{
  "dependencies": {
    "react": "^18",
    "react-dom": "^18",
    "react-router-dom": "^6",
    "@tanstack/react-query": "^5",
    "@tanstack/react-table": "^8",
    "react-hook-form": "^7",
    "@hookform/resolvers": "^3",
    "zod": "^3",
    "zustand": "^4",
    "axios": "^1",
    "@tremor/react": "^3",
    "sonner": "^1",
    "lucide-react": "^0.300",
    "tailwindcss": "^3",
    "class-variance-authority": "^0.7",
    "clsx": "^2",
    "tailwind-merge": "^2"
  }
}
```

---

## Scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext ts,tsx",
    "types:generate": "openapi-typescript ../openapi/index.yaml -o src/types/api.d.ts"
  }
}
```

---

## Environment

```bash
VITE_API_URL=http://localhost:8088
VITE_APP_NAME=AIGateway
```
