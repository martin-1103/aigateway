import { apiClient } from '@/lib/api-client'
import type { ModelMappingsResponse } from '../model-mappings.types'

export async function getModelMappings(): Promise<ModelMappingsResponse> {
  const response = await apiClient.get<ModelMappingsResponse>('/api/v1/model-mappings')
  return response.data
}
