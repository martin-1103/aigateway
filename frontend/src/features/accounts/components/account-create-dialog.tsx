import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form'
import { FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { useCreateAccount } from '../hooks'
import { createAccountSchema, type CreateAccountFormData } from '../accounts.schemas'

const PROVIDERS = ['antigravity', 'openai', 'glm'] as const

interface AccountCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function AccountCreateDialog({ open, onOpenChange }: AccountCreateDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateAccountFormData>({
    resolver: zodResolver(createAccountSchema),
    defaultValues: {
      provider_id: '',
      label: '',
      auth_data: '',
      is_active: true,
    },
  })

  const createMutation = useCreateAccount({
    onSuccess: () => {
      reset()
      onOpenChange(false)
    },
  })

  const onSubmit = handleSubmit((data) => {
    createMutation.mutate(data)
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={(isOpen) => {
        if (!isOpen) reset()
        onOpenChange(isOpen)
      }}
      title="Create Account"
      description="Add a new provider account to the gateway."
      onSubmit={onSubmit}
      isSubmitting={createMutation.isPending}
      submitLabel="Create"
    >
      <FormField label="Provider" error={errors.provider_id?.message}>
        <select
          {...register('provider_id')}
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
          aria-label="Select provider"
        >
          <option value="">Select a provider</option>
          {PROVIDERS.map((provider) => (
            <option key={provider} value={provider}>
              {provider.charAt(0).toUpperCase() + provider.slice(1)}
            </option>
          ))}
        </select>
      </FormField>

      <FormField label="Account Label" error={errors.label?.message}>
        <Input
          {...register('label')}
          type="text"
          placeholder="My OpenAI Account"
          autoComplete="off"
        />
      </FormField>

      <FormField label="Credentials (JSON)" error={errors.auth_data?.message}>
        <textarea
          {...register('auth_data')}
          rows={4}
          placeholder='{"client_id": "...", "client_secret": "..."}'
          className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
        />
      </FormField>

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
    </FormDialog>
  )
}
