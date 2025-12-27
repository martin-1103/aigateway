import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form/form-dialog'
import { FormField } from '@/components/form/form-field'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { useCreateModelMapping } from '../hooks'
import {
  createModelMappingSchema,
  type CreateModelMappingFormData,
} from '../model-mappings.schemas'

const PROVIDERS = [
  { value: 'antigravity', label: 'Antigravity (Google Cloud)' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'glm', label: 'GLM (Chinese LLMs)' },
]

interface ModelMappingCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelMappingCreateDialog({ open, onOpenChange }: ModelMappingCreateDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateModelMappingFormData>({
    resolver: zodResolver(createModelMappingSchema),
    defaultValues: {
      alias: '',
      provider_id: 'antigravity',
      model_name: '',
      description: '',
    },
  })

  const createMutation = useCreateModelMapping({
    onSuccess: () => {
      reset()
      onOpenChange(false)
    },
  })

  const onSubmit = handleSubmit(async (data) => {
    setIsSubmitting(true)
    try {
      await createMutation.mutateAsync(data)
    } finally {
      setIsSubmitting(false)
    }
  })

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      reset()
    }
    onOpenChange(newOpen)
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="Create Model Mapping"
      description="Create a new alias that maps to an existing model."
      isSubmitting={isSubmitting}
      onSubmit={onSubmit}
      submitLabel="Create"
    >
      <FormField label="Alias" error={errors.alias?.message}>
        <Input
          {...register('alias')}
          placeholder="e.g., gpt-4"
          autoComplete="off"
          aria-describedby={errors.alias ? 'alias-error' : undefined}
        />
      </FormField>
      <FormField label="Provider" error={errors.provider_id?.message}>
        <Select {...register('provider_id')} defaultValue="antigravity">
          {PROVIDERS.map((provider) => (
            <option key={provider.value} value={provider.value}>
              {provider.label}
            </option>
          ))}
        </Select>
      </FormField>
      <FormField label="Model Name" error={errors.model_name?.message}>
        <Input
          {...register('model_name')}
          placeholder="e.g., claude-opus-4-5"
          autoComplete="off"
          aria-describedby={errors.model_name ? 'model-name-error' : undefined}
        />
      </FormField>
      <FormField label="Description" error={errors.description?.message}>
        <Input
          {...register('description')}
          placeholder="e.g., Latest Claude Opus"
          autoComplete="off"
          aria-describedby={errors.description ? 'description-error' : undefined}
        />
      </FormField>
    </FormDialog>
  )
}
