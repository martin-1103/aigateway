import { apiClient } from '@/lib/api-client'
import type { UpdateModelMappingRequest, ModelMapping } from '../model-mappings.types'

export async function updateModelMapping(
  alias: string,
  data: UpdateModelMappingRequest
): Promise<ModelMapping> {
  const response = await apiClient.put<ModelMapping>(
    `/api/v1/model-mappings/${encodeURIComponent(alias)}`,
    data
  )
  return response.data
}
