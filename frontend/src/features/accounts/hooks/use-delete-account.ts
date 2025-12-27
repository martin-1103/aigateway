import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteAccount } from '../api'
import { handleError } from '@/lib/handle-error'
import { ACCOUNTS_QUERY_KEY } from './use-accounts-query'

export function useDeleteAccount(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY })
      toast.success('Account deleted successfully')
      options?.onSuccess?.()
    },
    onError: handleError,
  })
}
