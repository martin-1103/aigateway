import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form'
import { FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { useCreateAccount } from '../hooks'
import { createAccountSchema, type CreateAccountFormData } from '../accounts.schemas'
import type { Provider } from '../accounts.types'
import { apiClient } from '@/lib/api-client'
import { Loader2, ExternalLink } from 'lucide-react'
import { useExchangeOAuth } from '@/features/oauth/hooks'

interface AccountCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

type OAuthStep = 'idle' | 'started' | 'callback'

export function AccountCreateDialog({ open, onOpenChange }: AccountCreateDialogProps) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [providersLoading, setProvidersLoading] = useState(false)
  const [selectedProvider, setSelectedProvider] = useState<Provider | null>(null)
  const [selectedAuthType, setSelectedAuthType] = useState<'oauth' | 'api_key' | ''>('')
  const [oauthLoading, setOauthLoading] = useState(false)
  const [projectId, setProjectId] = useState('')
  const [oauthStep, setOauthStep] = useState<OAuthStep>('idle')
  const [authUrl, setAuthUrl] = useState('')
  const [callbackUrl, setCallbackUrl] = useState('')

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

  const exchangeMutation = useExchangeOAuth({
    onSuccess: () => {
      handleClose()
    },
  })

  useEffect(() => {
    if (open && providers.length === 0) {
      fetchProviders()
    }
  }, [open])

  useEffect(() => {
    const provider = providers.find((p) => p.id === providerIdValue)
    setSelectedProvider(provider || null)
    setSelectedAuthType('')
    setValue('auth_data', '')
    setOauthStep('idle')
    setAuthUrl('')
    setCallbackUrl('')
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
      handleClose()
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
    if (!selectedProvider) return
    // Only require projectId for antigravity
    if (requiresProjectId && !projectId.trim()) return

    try {
      setOauthLoading(true)

      const initPayload: Record<string, string> = {
        provider: selectedProvider.id,
        flow_type: 'manual',
      }
      // Only include project_id if provider requires it
      if (requiresProjectId) {
        initPayload.project_id = projectId.trim()
      }

      const initResponse = await apiClient.post<{ auth_url: string }>('/api/v1/oauth/init', initPayload)

      setAuthUrl(initResponse.data.auth_url)
      setOauthStep('callback')
      window.open(initResponse.data.auth_url, '_blank')
    } catch (error) {
      console.error('Failed to initiate OAuth:', error)
    } finally {
      setOauthLoading(false)
    }
  }

  const handleCallbackSubmit = () => {
    if (!callbackUrl.trim()) return
    exchangeMutation.mutate({ callback_url: callbackUrl.trim() })
  }

  const handleClose = () => {
    reset()
    setSelectedProvider(null)
    setSelectedAuthType('')
    setProjectId('')
    setOauthStep('idle')
    setAuthUrl('')
    setCallbackUrl('')
    onOpenChange(false)
  }

  const supportsOAuth = selectedProvider?.supported_auth_types.includes('oauth') ?? false
  const supportsAPIKey = selectedProvider?.supported_auth_types.includes('api_key') ?? false
  const requiresProjectId = selectedProvider?.id === 'antigravity'

  return (
    <FormDialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) handleClose()
        else onOpenChange(isOpen)
      }}
      title="Create Account"
      description="Add a new provider account to the gateway."
      onSubmit={onSubmit}
      isSubmitting={createMutation.isPending}
      submitLabel="Create"
      hideSubmit={selectedAuthType === 'oauth'}
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
                            setOauthStep('idle')
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
                            setOauthStep('idle')
                          }}
                          className="h-4 w-4"
                        />
                        <span className="text-sm">API Key</span>
                      </label>
                    )}
                  </div>
                </FormField>
              )}

              {/* OAuth Flow - Step 1: Start */}
              {selectedAuthType === 'oauth' && oauthStep === 'idle' && (
                <>
                  {requiresProjectId && (
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
                  )}
                  <FormField label="OAuth Authentication">
                    <Button
                      type="button"
                      onClick={handleOAuthClick}
                      disabled={oauthLoading || (requiresProjectId && !projectId.trim())}
                      className="w-full"
                    >
                      {oauthLoading ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Starting...
                        </>
                      ) : (
                        <>
                          <ExternalLink className="mr-2 h-4 w-4" />
                          Start OAuth Flow
                        </>
                      )}
                    </Button>
                  </FormField>
                </>
              )}

              {/* OAuth Flow - Step 2: Callback URL Input */}
              {selectedAuthType === 'oauth' && oauthStep === 'callback' && (
                <>
                  <div className="rounded-md bg-muted p-3 text-sm">
                    <p className="font-medium">Instructions:</p>
                    <ol className="mt-2 list-decimal pl-4 space-y-1 text-muted-foreground">
                      <li>Complete Google login in the new tab</li>
                      <li>After login, copy the URL from address bar</li>
                      <li>Paste it below and click Submit</li>
                    </ol>
                  </div>

                  {authUrl && (
                    <div className="space-y-2">
                      <Label className="text-sm">Auth URL</Label>
                      <div className="flex gap-2">
                        <Input value={authUrl} disabled className="text-xs flex-1" />
                        <Button
                          type="button"
                          variant="outline"
                          size="icon"
                          onClick={() => window.open(authUrl, '_blank')}
                        >
                          <ExternalLink className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  )}

                  <div className="space-y-2">
                    <Label className="text-sm">Callback URL</Label>
                    <textarea
                      value={callbackUrl}
                      onChange={(e) => setCallbackUrl(e.target.value)}
                      placeholder="Paste the full callback URL here..."
                      rows={3}
                      className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                    />
                  </div>

                  <div className="flex gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => {
                        setOauthStep('idle')
                        setCallbackUrl('')
                      }}
                      className="flex-1"
                    >
                      Back
                    </Button>
                    <Button
                      type="button"
                      onClick={handleCallbackSubmit}
                      disabled={exchangeMutation.isPending || !callbackUrl.trim()}
                      className="flex-1"
                    >
                      {exchangeMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Submitting...
                        </>
                      ) : (
                        'Submit Callback URL'
                      )}
                    </Button>
                  </div>
                </>
              )}

              {/* API Key Flow */}
              {selectedAuthType === 'api_key' && (
                <>
                  <FormField label="Account Label" error={errors.label?.message}>
                    <Input
                      {...register('label')}
                      type="text"
                      placeholder="My Account"
                      autoComplete="off"
                    />
                  </FormField>
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
                </>
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
