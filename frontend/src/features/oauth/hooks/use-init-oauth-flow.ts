import { useMutation } from '@tanstack/react-query'
import { toast } from 'sonner'
import { initOAuthFlow } from '../api'
import { handleError } from '@/lib/handle-error'
import { logger } from '@/lib/logger'

export function useInitOAuthFlow() {
  return useMutation({
    mutationFn: async (data: Parameters<typeof initOAuthFlow>[0]) => {
      logger.oauth.log('Sending init request', data)
      try {
        const response = await initOAuthFlow(data)
        logger.oauth.info('Init success', response)
        return response
      } catch (error) {
        logger.oauth.error('Init failed', error)
        throw error
      }
    },
    onSuccess: (data, variables) => {
      logger.oauth.log('OAuth init success', data.auth_url)
      if (variables.flow_type === 'manual') {
        toast.success('OAuth flow initiated. Complete login in the new tab.')
      } else {
        toast.success('OAuth flow initiated. Redirecting...')
        window.location.href = data.auth_url
      }
    },
    onError: (error) => {
      logger.oauth.error('Mutation error', error)
      handleError(error)
    },
  })
}
