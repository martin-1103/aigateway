import { useQuery } from '@tanstack/react-query'
import { getModelMappings } from '../api'

export const MODEL_MAPPINGS_QUERY_KEY = ['model-mappings']

export function useModelMappingsQuery() {
  return useQuery({
    queryKey: MODEL_MAPPINGS_QUERY_KEY,
    queryFn: getModelMappings,
  })
}
