import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { createAccount } from '../api'
import { handleError } from '@/lib/handle-error'
import { ACCOUNTS_QUERY_KEY } from './use-accounts-query'

export function useCreateAccount(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY })
      toast.success('Account created successfully')
      options?.onSuccess?.()
    },
    onError: handleError,
  })
}
