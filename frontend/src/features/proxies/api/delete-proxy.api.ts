import { apiClient } from '@/lib/api-client'

interface DeleteResponse {
  message: string
}

export async function deleteProxy(id: number): Promise<DeleteResponse> {
  const response = await apiClient.delete<DeleteResponse>(`/api/v1/proxies/${id}`)
  return response.data
}
