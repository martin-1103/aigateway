import { z } from 'zod'

export const createAccountSchema = z.object({
  provider_id: z.string().min(1, 'Provider is required'),
  label: z.string().min(1, 'Account label is required'),
  auth_data: z.string().min(1, 'Credentials are required'),
  is_active: z.boolean().default(true),
})

export const updateAccountSchema = z.object({
  provider_id: z.string().min(1, 'Provider is required'),
  label: z.string().min(1, 'Account label is required'),
  auth_data: z.string().optional(),
  is_active: z.boolean(),
})

export type CreateAccountFormData = z.infer<typeof createAccountSchema>
export type UpdateAccountFormData = z.infer<typeof updateAccountSchema>
