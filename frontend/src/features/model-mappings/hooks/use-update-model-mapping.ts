import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { updateModelMapping } from '../api'
import { handleError } from '@/lib/handle-error'
import { MODEL_MAPPINGS_QUERY_KEY } from './use-model-mappings-query'
import type { UpdateModelMappingRequest } from '../model-mappings.types'

export function useUpdateModelMapping(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ alias, data }: { alias: string; data: UpdateModelMappingRequest }) =>
      updateModelMapping(alias, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: MODEL_MAPPINGS_QUERY_KEY })
      toast.success('Model mapping updated successfully')
      options?.onSuccess?.()
    },
    onError: handleError,
  })
}
