import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog, FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { useCreateUser } from '../hooks'
import { createUserSchema, type CreateUserFormData } from '../users.schemas'

interface UserCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function UserCreateDialog({ open, onOpenChange }: UserCreateDialogProps) {
  const { mutate, isPending } = useCreateUser(() => {
    onOpenChange(false)
    reset()
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateUserFormData>({
    resolver: zodResolver(createUserSchema),
    defaultValues: {
      username: '',
      password: '',
      role: 'user',
      is_active: true,
    },
  })

  const onSubmit = (data: CreateUserFormData) => mutate(data)

  return (
    <FormDialog
      open={open}
      onOpenChange={(value) => {
        onOpenChange(value)
        if (!value) reset()
      }}
      title="Create User"
      description="Add a new user to the system."
      isSubmitting={isPending}
      onSubmit={handleSubmit(onSubmit)}
      submitLabel="Create"
    >
      <FormField label="Username" error={errors.username?.message}>
        <Input {...register('username')} placeholder="Enter username" />
      </FormField>

      <FormField label="Password" error={errors.password?.message}>
        <Input
          {...register('password')}
          type="password"
          placeholder="Enter password"
        />
      </FormField>

      <FormField label="Role" error={errors.role?.message}>
        <select
          {...register('role')}
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <option value="user">User</option>
          <option value="admin">Admin</option>
          <option value="provider">Provider</option>
        </select>
      </FormField>

      <FormField label="Status">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            {...register('is_active')}
            className="h-4 w-4 rounded border-gray-300"
          />
          <span className="text-sm">Active</span>
        </label>
      </FormField>
    </FormDialog>
  )
}
