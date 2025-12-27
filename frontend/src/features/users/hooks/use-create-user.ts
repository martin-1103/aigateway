import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { createUser } from '../api'
import { usersQueryKey } from './use-users-query'
import { handleError } from '@/lib/handle-error'

export function useCreateUser(onSuccess?: () => void) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createUser,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: usersQueryKey })
      toast.success('User created successfully')
      onSuccess?.()
    },
    onError: handleError,
  })
}
