import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog } from '@/components/form/form-dialog'
import { FormField } from '@/components/form/form-field'
import { Input } from '@/components/ui/input'
import { useUpdateModelMapping } from '../hooks'
import {
  updateModelMappingSchema,
  type UpdateModelMappingFormData,
} from '../model-mappings.schemas'
import type { ModelMapping } from '../model-mappings.types'

interface ModelMappingEditDialogProps {
  mapping: ModelMapping | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelMappingEditDialog({
  mapping,
  open,
  onOpenChange,
}: ModelMappingEditDialogProps) {
  const [isSubmitting, setIsSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<UpdateModelMappingFormData>({
    resolver: zodResolver(updateModelMappingSchema),
    defaultValues: {
      target_model: '',
    },
  })

  useEffect(() => {
    if (mapping) {
      reset({ target_model: mapping.target_model })
    }
  }, [mapping, reset])

  const updateMutation = useUpdateModelMapping({
    onSuccess: () => {
      onOpenChange(false)
    },
  })

  const onSubmit = handleSubmit(async (data) => {
    if (!mapping) return
    setIsSubmitting(true)
    try {
      await updateMutation.mutateAsync({ alias: mapping.alias, data })
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
      title="Edit Model Mapping"
      description={`Update the target model for alias "${mapping?.alias}".`}
      isSubmitting={isSubmitting}
      onSubmit={onSubmit}
      submitLabel="Save"
    >
      <FormField label="Alias">
        <Input value={mapping?.alias ?? ''} disabled className="bg-muted" />
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
