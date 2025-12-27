import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { updateUser } from '../api'
import { usersQueryKey } from './use-users-query'
import { handleError } from '@/lib/handle-error'
import type { UpdateUserRequest } from '../users.types'

export function useUpdateUser(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateUserRequest }) =>
      updateUser(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: usersQueryKey })
      toast.success('User updated successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
