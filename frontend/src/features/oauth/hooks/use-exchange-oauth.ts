import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { exchangeOAuth } from '../api'
import { handleError } from '@/lib/handle-error'

export function useExchangeOAuth(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: exchangeOAuth,
    onSuccess: (data) => {
      toast.success(`Account "${data.account.label}" connected successfully!`)
      queryClient.invalidateQueries({ queryKey: ['accounts'] })
      options?.onSuccess?.()
    },
    onError: (error) => {
      handleError(error)
    },
  })
}
