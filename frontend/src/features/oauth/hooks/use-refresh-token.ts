import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { refreshToken } from '../api'
import { handleError } from '@/lib/handle-error'

export function useRefreshToken() {
  return useMutation({
    mutationFn: refreshToken,
    onSuccess: (data) => {
      toast.success(data.message || 'Token refreshed successfully')
    },
    onError: handleError,
  })
}
