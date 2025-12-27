import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { revokeApiKey } from '../api'
import { handleError } from '@/lib/handle-error'
import { API_KEYS_QUERY_KEY } from './use-api-keys-query'

export function useRevokeApiKey() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: revokeApiKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: API_KEYS_QUERY_KEY })
      toast.success('API key revoked successfully')
    },
    onError: handleError,
  })
}
