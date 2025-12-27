import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { initOAuthFlow } from '../api'
import { handleError } from '@/lib/handle-error'

export function useInitOAuthFlow() {
  return useMutation({
    mutationFn: initOAuthFlow,
    onSuccess: (data) => {
      toast.success('OAuth flow initiated. Redirecting...')
      window.location.href = data.auth_url
    },
    onError: handleError,
  })
}
