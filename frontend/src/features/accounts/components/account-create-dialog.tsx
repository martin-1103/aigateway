import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form'
import { FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useCreateAccount } from '../hooks'
import { createAccountSchema, type CreateAccountFormData } from '../accounts.schemas'
import type { Provider } from '../accounts.types'
import { apiClient } from '@/lib/api-client'
import { Loader2 } from 'lucide-react'

interface AccountCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AccountCreateDialog({ open, onOpenChange }: AccountCreateDialogProps) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [providersLoading, setProvidersLoading] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null)
  const [selectedAuthType, setSelectedAuthType] = useState<'oauth' | 'api_key' | ''>('')
  const [oauthLoading, setOauthLoading] = useState(false)
  const [projectId, setProjectId] = useState('')

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
    setValue,
    watch,
  } = useForm<CreateAccountFormData>({
    resolver: zodResolver(createAccountSchema),
    defaultValues: {
      provider_id: '',
      label: '',
      auth_data: '',
      is_active: true,
    },
  })

  const providerIdValue = watch('provider_id')

  // Fetch providers on dialog open
  useEffect(() => {
    if (open && providers.length === 0) {
      fetchProviders()
    }
  }, [open])

  // Update selected provider and auth types when provider_id changes
  useEffect(() => {
    const provider = providers.find((p) => p.id === providerIdValue)
    setSelectedProvider(provider || null)
    setSelectedAuthType('')
    setValue('auth_data', '')
  }, [providerIdValue, providers, setValue])

  const fetchProviders = async () => {
    try {
      setProvidersLoading(true)
      const response = await apiClient.get<{ providers: Provider[] }>('/api/v1/providers')
      setProviders(response.data.providers || [])
    } catch (error) {
      console.error('Failed to fetch providers:', error)
    } finally {
      setProvidersLoading(false)
    }
  }

  const createMutation = useCreateAccount({
    onSuccess: () => {
      reset()
      setSelectedProvider(null)
      setSelectedAuthType('')
      setProjectId('')
      onOpenChange(false)
    },
  })

  const onSubmit = handleSubmit((data) => {
    if (!selectedAuthType) {
      console.error('Auth type not selected')
      return
    }

    const payload = {
      ...data,
      auth_type: selectedAuthType,
    }
    createMutation.mutate(payload as any)
  })

  const handleOAuthClick = async () => {
    if (!selectedProvider || !projectId.trim()) return

    try {
      setOauthLoading(true)

      // Generate PKCE codes and get auth URL
      const initResponse = await apiClient.post<{ auth_url: string }>('/api/v1/oauth/init', {
        provider: selectedProvider.id,
        project_id: projectId.trim(),
        flow_type: 'auto',
      })

      // Open OAuth in popup
      const width = 600
      const height = 700
      const left = window.innerWidth / 2 - width / 2
      const top = window.innerHeight / 2 - height / 2

      const popup = window.open(
        initResponse.data.auth_url,
        'oauth-popup',
        `width=${width},height=${height},left=${left},top=${top}`
      )

      if (!popup) {
        alert('Popup was blocked. Please allow popups for this site.')
        return
      }

      // Listen for OAuth success message
      const handleOAuthMessage = (event: MessageEvent) => {
        if (event.data.type === 'oauth_success') {
          setSelectedAuthType('oauth')
          // Auth data will be stored from the OAuth callback
          // The account is already created by the backend
          popup.close()
          setOauthLoading(false)
          // Close dialog on success
          setTimeout(() => onOpenChange(false), 500)
        } else if (event.data.type === 'oauth_error') {
          console.error('OAuth error:', event.data.error)
          setOauthLoading(false)
        }
      }

      window.addEventListener('message', handleOAuthMessage)
      return () => window.removeEventListener('message', handleOAuthMessage)
    } catch (error) {
      console.error('Failed to initiate OAuth:', error)
      setOauthLoading(false)
    }
  }

  const supportsOAuth = selectedProvider?.supported_auth_types.includes('oauth') ?? false
  const supportsAPIKey = selectedProvider?.supported_auth_types.includes('api_key') ?? false

  return (
    <FormDialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          reset()
          setSelectedProvider(null)
          setSelectedAuthType('')
          setProjectId('')
        }
        onOpenChange(isOpen)
      }}
      title="Create Account"
      description="Add a new provider account to the gateway."
      onSubmit={onSubmit}
      isSubmitting={createMutation.isPending}
      submitLabel="Create"
    >
      {providersLoading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-6 w-6 animate-spin" />
        </div>
      ) : (
        <>
          <FormField label="Provider" error={errors.provider_id?.message}>
            <select
              {...register('provider_id')}
              className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
              aria-label="Select provider"
            >
              <option value="">Select a provider</option>
              {providers.map((provider) => (
                <option key={provider.id} value={provider.id}>
                  {provider.name}
                </option>
              ))}
            </select>
          </FormField>

          {selectedProvider && (
            <>
              {/* Auth Type Selection - shown first */}
              {(supportsOAuth || supportsAPIKey) && (
                <FormField label="Authentication Method">
                  <div className="space-y-2">
                    {supportsOAuth && (
                      <label className="flex items-center gap-2">
                        <input
                          type="radio"
                          name="auth_type"
                          value="oauth"
                          checked={selectedAuthType === 'oauth'}
                          onChange={(e) => {
                            setSelectedAuthType(e.target.value as 'oauth')
                            setValue('auth_data', '')
                          }}
                          className="h-4 w-4"
                        />
                        <span className="text-sm">OAuth</span>
                      </label>
                    )}
                    {supportsAPIKey && (
                      <label className="flex items-center gap-2">
                        <input
                          type="radio"
                          name="auth_type"
                          value="api_key"
                          checked={selectedAuthType === 'api_key'}
                          onChange={(e) => {
                            setSelectedAuthType(e.target.value as 'api_key')
                            setProjectId('')
                          }}
                          className="h-4 w-4"
                        />
                        <span className="text-sm">API Key</span>
                      </label>
                    )}
                  </div>
                </FormField>
              )}

              {/* OAuth Flow */}
              {selectedAuthType === 'oauth' && (
                <>
                  <FormField label="Project ID">
                    <Input
                      value={projectId}
                      onChange={(e) => setProjectId(e.target.value)}
                      placeholder="e.g., my-project, project-123, etc."
                      autoComplete="off"
                    />
                    <p className="text-xs text-muted-foreground mt-1">
                      Used for quota tracking
                    </p>
                  </FormField>
                  <FormField label="OAuth Authentication">
                    <Button
                      type="button"
                      onClick={handleOAuthClick}
                      disabled={oauthLoading || !projectId.trim()}
                      className="w-full"
                    >
                      {oauthLoading ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Initiating...
                        </>
                      ) : (
                        'Start OAuth Flow'
                      )}
                    </Button>
                    <p className="text-xs text-muted-foreground mt-2">
                      Click to open OAuth login in a popup window
                    </p>
                  </FormField>
                </>
              )}

              {/* Account Label - only shown for API Key auth */}
              {selectedAuthType === 'api_key' && (
                <FormField label="Account Label" error={errors.label?.message}>
                  <Input
                    {...register('label')}
                    type="text"
                    placeholder="My Account"
                    autoComplete="off"
                  />
                </FormField>
              )}

              {/* API Key Input */}
              {selectedAuthType === 'api_key' && (
                <FormField label="Credentials (JSON)" error={errors.auth_data?.message}>
                  <textarea
                    {...register('auth_data')}
                    rows={4}
                    placeholder={
                      selectedProvider.id === 'openai'
                        ? '{"api_key": "sk-..."}'
                        : selectedProvider.id === 'glm'
                          ? '{"api_key": "..."}'
                          : '{"api_key": "..."}'
                    }
                    className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                  />
                </FormField>
              )}
            </>
          )}

          <FormField label="Status">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                {...register('is_active')}
                className="h-4 w-4 rounded border-input"
              />
              <span className="text-sm">Active</span>
            </label>
          </FormField>
        </>
      )}
    </FormDialog>
  )
}
