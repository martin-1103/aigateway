import { useState, useEffect } from 'react'
import { getLiteMe, type LiteUser } from '../api/lite-me.api'
import { getAccessKey } from '../api/lite.client'

interface UseLiteAuthResult {
  user: LiteUser | null
  isLoading: boolean
  error: string | null
  accessKey: string | null
}

export function useLiteAuth(): UseLiteAuthResult {
  const [user, setUser] = useState<LiteUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const accessKey = getAccessKey()

  useEffect(() => {
    if (!accessKey) {
      setError('Access key required')
      setIsLoading(false)
      return
    }

    getLiteMe()
      .then(setUser)
      .catch((err) => {
        setError(err.response?.data?.error || 'Invalid access key')
      })
      .finally(() => setIsLoading(false))
  }, [accessKey])

  return { user, isLoading, error, accessKey }
}
