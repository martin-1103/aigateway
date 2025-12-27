import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form'
import { FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { useUpdateAccount } from '../hooks'
import { updateAccountSchema, type UpdateAccountFormData } from '../accounts.schemas'
import type { Account } from '../accounts.types'

interface AccountEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  account: Account | null
}

export function AccountEditDialog({ open, onOpenChange, account }: AccountEditDialogProps) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateAccountFormData>({
    resolver: zodResolver(updateAccountSchema),
  })

  useEffect(() => {
    if (account) {
      reset({
        label: account.label,
        auth_data: '',
        is_active: account.is_active,
      })
    }
  }, [account, reset])

  const updateMutation = useUpdateAccount({
    onSuccess: () => {
      onOpenChange(false)
    },
  })

  const onSubmit = handleSubmit((data) => {
    if (!account) return

    const payload = {
      label: data.label,
      is_active: data.is_active,
      ...(data.auth_data ? { auth_data: data.auth_data } : {}),
    }

    updateMutation.mutate({ id: account.id, data: payload })
  })

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Edit Account"
      description="Update the account details."
      onSubmit={onSubmit}
      isSubmitting={updateMutation.isPending}
      submitLabel="Save Changes"
    >
      <FormField label="Provider">
        <Input
          value={account?.provider_id || ''}
          disabled
          type="text"
          placeholder="Provider"
        />
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
          placeholder="Leave empty to keep existing credentials"
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
