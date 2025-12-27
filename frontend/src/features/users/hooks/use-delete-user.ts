import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { deleteUser } from '../api'
import { usersQueryKey } from './use-users-query'
import { handleError } from '@/lib/handle-error'

export function useDeleteUser(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: usersQueryKey })
      toast.success('User deleted successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
