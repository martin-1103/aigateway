import { z } from 'zod'

export const createModelMappingSchema = z.object({
  alias: z
    .string()
    .min(1, 'Alias is required')
    .max(100, 'Alias must be 100 characters or less')
    .regex(/^[a-zA-Z0-9-_.]+$/, 'Alias can only contain letters, numbers, hyphens, underscores, and dots'),
  provider_id: z
    .string()
    .min(1, 'Provider is required'),
  model_name: z
    .string()
    .min(1, 'Model name is required')
    .max(100, 'Model name must be 100 characters or less'),
  description: z
    .string()
    .max(500, 'Description must be 500 characters or less')
    .optional(),
})

export const updateModelMappingSchema = z.object({
  model_name: z
    .string()
    .min(1, 'Model name is required')
    .max(100, 'Model name must be 100 characters or less'),
  description: z
    .string()
    .max(500, 'Description must be 500 characters or less')
    .optional(),
})

export type CreateModelMappingFormData = z.infer<typeof createModelMappingSchema>
export type UpdateModelMappingFormData = z.infer<typeof updateModelMappingSchema>
