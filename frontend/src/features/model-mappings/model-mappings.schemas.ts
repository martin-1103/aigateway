import { z } from 'zod'

export const createModelMappingSchema = z.object({
  alias: z
    .string()
    .min(1, 'Alias is required')
    .max(100, 'Alias must be 100 characters or less')
    .regex(/^[a-zA-Z0-9-_.]+$/, 'Alias can only contain letters, numbers, hyphens, underscores, and dots'),
  target_model: z
    .string()
    .min(1, 'Target model is required')
    .max(100, 'Target model must be 100 characters or less'),
})

export const updateModelMappingSchema = z.object({
  target_model: z
    .string()
    .min(1, 'Target model is required')
    .max(100, 'Target model must be 100 characters or less'),
})

export type CreateModelMappingFormData = z.infer<typeof createModelMappingSchema>
export type UpdateModelMappingFormData = z.infer<typeof updateModelMappingSchema>
