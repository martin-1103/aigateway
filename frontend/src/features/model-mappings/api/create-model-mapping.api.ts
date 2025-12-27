import { apiClient } from '@/lib/api-client'
import type { CreateModelMappingRequest, ModelMapping } from '../model-mappings.types'

export async function createModelMapping(data: CreateModelMappingRequest): Promise<ModelMapping> {
  const response = await apiClient.post<ModelMapping>('/api/v1/model-mappings', data)
  return response.data
}
