import { useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { login } from '../api'
import { useAuthStore } from '../auth.store'
import { handleError } from '@/lib/handle-error'

export function useLogin() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  return useMutation({
    mutationFn: login,
    onSuccess: (data) => {
      setAuth(data.token, data.user)
      toast.success('Login successful')
      navigate('/dashboard')
    },
    onError: handleError,
  })
}
