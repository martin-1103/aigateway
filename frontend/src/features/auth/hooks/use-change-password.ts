import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { changePassword } from '../api'
import { handleError } from '@/lib/handle-error'

export function useChangePassword() {
  return useMutation({
    mutationFn: changePassword,
    onSuccess: () => {
      toast.success('Password changed successfully')
    },
    onError: handleError,
  })
}
