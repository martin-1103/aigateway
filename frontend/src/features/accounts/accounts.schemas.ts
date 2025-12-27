import { z } from 'zod'

export const createAccountSchema = z.object({
  provider: z.string().min(1, 'Provider is required'),
  email: z.string().email('Invalid email address'),
  credentials: z.string().min(1, 'Credentials are required'),
  is_active: z.boolean().default(true),
})

export const updateAccountSchema = z.object({
  provider: z.string().min(1, 'Provider is required'),
  email: z.string().email('Invalid email address'),
  credentials: z.string().optional(),
  is_active: z.boolean(),
})

export type CreateAccountFormData = z.infer<typeof createAccountSchema>
export type UpdateAccountFormData = z.infer<typeof updateAccountSchema>
