import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog, FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import { useUpdateUser } from '../hooks'
import { updateUserSchema, type UpdateUserFormData } from '../users.schemas'
import type { User } from '../users.types'

interface UserEditDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  user: User | null
}

export function UserEditDialog({ open, onOpenChange, user }: UserEditDialogProps) {
  const { mutate, isPending } = useUpdateUser(() => {
    onOpenChange(false)
  })

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateUserFormData>({
    resolver: zodResolver(updateUserSchema),
  })

  useEffect(() => {
    if (user) {
      reset({
        username: user.username,
        password: '',
        role: user.role,
        is_active: user.is_active,
      })
    }
  }, [user, reset])

  const onSubmit = (data: UpdateUserFormData) => {
    if (!user) return
    const payload = {
      username: data.username,
      role: data.role,
      is_active: data.is_active,
      ...(data.password ? { password: data.password } : {}),
    }
    mutate({ id: user.id, data: payload })
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Edit User"
      description="Update user information."
      isSubmitting={isPending}
      onSubmit={handleSubmit(onSubmit)}
      submitLabel="Save"
    >
      <FormField label="Username" error={errors.username?.message}>
        <Input {...register('username')} placeholder="Enter username" />
      </FormField>

      <FormField label="Password" error={errors.password?.message}>
        <Input
          {...register('password')}
          type="password"
          placeholder="Leave blank to keep current"
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
