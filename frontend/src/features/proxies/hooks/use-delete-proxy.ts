import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteProxy } from '../api'
import { handleError } from '@/lib/handle-error'
import { PROXIES_QUERY_KEY } from './use-proxies-query'

export function useDeleteProxy(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteProxy,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PROXIES_QUERY_KEY })
      toast.success('Proxy deleted successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
