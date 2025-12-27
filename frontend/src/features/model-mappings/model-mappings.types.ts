export interface ModelMapping {
  alias: string
  target_model: string
  created_at: string
  updated_at: string
}

export interface CreateModelMappingRequest {
  alias: string
  target_model: string
}

export interface UpdateModelMappingRequest {
  target_model: string
}

export interface ModelMappingsResponse {
  mappings: ModelMapping[]
}
