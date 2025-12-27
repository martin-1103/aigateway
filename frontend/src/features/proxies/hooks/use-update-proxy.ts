import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { updateProxy } from '../api'
import { handleError } from '@/lib/handle-error'
import { PROXIES_QUERY_KEY } from './use-proxies-query'
import type { UpdateProxyRequest } from '../proxies.types'

export function useUpdateProxy(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateProxyRequest }) =>
      updateProxy(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: PROXIES_QUERY_KEY })
      toast.success('Proxy updated successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
