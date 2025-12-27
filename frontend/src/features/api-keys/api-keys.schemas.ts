import { z } from 'zod'

export const createApiKeySchema = z.object({
  label: z
    .string()
    .min(1, 'Label is required')
    .max(100, 'Label must be at most 100 characters'),
})

export type CreateApiKeyFormData = z.infer<typeof createApiKeySchema>
