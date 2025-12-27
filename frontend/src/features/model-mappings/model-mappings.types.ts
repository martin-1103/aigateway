export interface ModelMapping {
  id: number
  alias: string
  provider_id: string
  model_name: string
  description: string
  enabled: boolean
  priority: number
  created_at: string
  updated_at: string
}

export interface CreateModelMappingRequest {
  alias: string
  provider_id: string
  model_name: string
  description?: string
}

export interface UpdateModelMappingRequest {
  model_name: string
  description?: string
}

export interface ModelMappingsResponse {
  data: ModelMapping[]
  limit: number
  page: number
  total: number
}
