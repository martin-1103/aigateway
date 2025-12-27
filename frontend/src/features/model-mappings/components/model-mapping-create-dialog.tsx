import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form/form-dialog'
import { FormField } from '@/components/form/form-field'
import { Input } from '@/components/ui/input'
import { useCreateModelMapping } from '../hooks'
import {
  createModelMappingSchema,
  type CreateModelMappingFormData,
} from '../model-mappings.schemas'

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
      target_model: '',
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
      <FormField label="Target Model" error={errors.target_model?.message}>
        <Input
          {...register('target_model')}
          placeholder="e.g., claude-sonnet-4-20250514"
          autoComplete="off"
          aria-describedby={errors.target_model ? 'target-model-error' : undefined}
        />
      </FormField>
    </FormDialog>
  )
}
