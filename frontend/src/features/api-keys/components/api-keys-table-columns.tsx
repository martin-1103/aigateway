import type { ApiKey } from '../api-keys.types'

export interface ApiKeysTableColumn {
  key: keyof ApiKey | 'actions'
  label: string
  className?: string
}

export const apiKeysTableColumns: ApiKeysTableColumn[] = [
  { key: 'label', label: 'Label' },
  { key: 'key_prefix', label: 'Key Prefix' },
  { key: 'is_active', label: 'Status' },
  { key: 'created_at', label: 'Created' },
  { key: 'last_used_at', label: 'Last Used' },
  { key: 'actions', label: '', className: 'w-[100px]' },
]
