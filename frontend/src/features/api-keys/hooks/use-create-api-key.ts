import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { createApiKey } from '../api'
import { handleError } from '@/lib/handle-error'
import { API_KEYS_QUERY_KEY } from './use-api-keys-query'

export function useCreateApiKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createApiKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: API_KEYS_QUERY_KEY })
      toast.success('API key created successfully')
    },
    onError: handleError,
  })
}
