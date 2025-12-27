import { apiClient } from '@/lib/api-client'

export async function deleteModelMapping(alias: string): Promise<void> {
  await apiClient.delete(`/api/v1/model-mappings/${encodeURIComponent(alias)}`)
}
