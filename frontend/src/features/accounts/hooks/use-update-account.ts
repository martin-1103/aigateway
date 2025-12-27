import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { updateAccount } from '../api'
import { handleError } from '@/lib/handle-error'
import { ACCOUNTS_QUERY_KEY } from './use-accounts-query'
import type { UpdateAccountRequest } from '../accounts.types'

export function useUpdateAccount(options?: { onSuccess?: () => void }) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateAccountRequest }) =>
      updateAccount(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY })
      toast.success('Account updated successfully')
      options?.onSuccess?.()
    },
    onError: handleError,
  })
}
