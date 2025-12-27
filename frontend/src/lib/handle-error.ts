import { toast } from 'sonner'
import { AxiosError } from 'axios'

interface ApiError {
  error?: string
  message?: string
}

export function handleError(error: unknown): void {
  if (error instanceof AxiosError) {
    const data = error.response?.data as ApiError | undefined
    const message = data?.error ?? data?.message ?? 'Request failed'
    toast.error(message)
    return
  }

  if (error instanceof Error) {
    toast.error(error.message)
    return
  }

  toast.error('Something went wrong')
}
