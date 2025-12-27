import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteModelMapping } from '../api'
import { handleError } from '@/lib/handle-error'
import { MODEL_MAPPINGS_QUERY_KEY } from './use-model-mappings-query'

export function useDeleteModelMapping(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteModelMapping,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: MODEL_MAPPINGS_QUERY_KEY })
      toast.success('Model mapping deleted successfully')
      options?.onSuccess?.()
    },
    onError: handleError,
  })
}
