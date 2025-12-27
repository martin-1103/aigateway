import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { createProxy } from '../api'
import { handleError } from '@/lib/handle-error'
import { PROXIES_QUERY_KEY } from './use-proxies-query'

export function useCreateProxy(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createProxy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PROXIES_QUERY_KEY })
      toast.success('Proxy created successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
