import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { FormDialog, FormField } from '@/components/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { useCreateProxy } from '../hooks'
import { proxySchema, type ProxyFormData } from '../proxies.schemas'

interface ProxyCreateDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ProxyCreateDialog({ open, onOpenChange }: ProxyCreateDialogProps) {
  const { mutate, isPending } = useCreateProxy(() => onOpenChange(false))

  const {
    register,
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ProxyFormData>({
    resolver: zodResolver(proxySchema),
    defaultValues: {
      url: '',
      protocol: 'http',
      is_active: true,
      max_accounts: 10,
      priority: 0,
      weight: 1,
      max_failures: 5,
    },
  })

  const onSubmit = (data: ProxyFormData) => {
    mutate(data)
  }

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      reset()
    }
    onOpenChange(isOpen)
  }

  return (
    <FormDialog
      open={open}
      onOpenChange={handleOpenChange}
      title="Create Proxy"
      description="Add a new proxy server to the pool."
      isSubmitting={isPending}
      onSubmit={handleSubmit(onSubmit)}
      submitLabel="Create"
    >
      <FormField label="URL" error={errors.url?.message}>
        <Input
          {...register('url')}
          placeholder="http://proxy.example.com:8080"
        />
      </FormField>

      <FormField label="Protocol" error={errors.protocol?.message}>
        <Controller
          name="protocol"
          control={control}
          render={({ field }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger>
                <SelectValue placeholder="Select protocol" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="http">HTTP</SelectItem>
                <SelectItem value="https">HTTPS</SelectItem>
                <SelectItem value="socks5">SOCKS5</SelectItem>
              </SelectContent>
            </Select>
          )}
        />
      </FormField>

      <div className="grid grid-cols-2 gap-4">
        <FormField label="Max Accounts" error={errors.max_accounts?.message}>
          <Input type="number" {...register('max_accounts')} />
        </FormField>
        <FormField label="Max Failures" error={errors.max_failures?.message}>
          <Input type="number" {...register('max_failures')} />
        </FormField>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <FormField label="Priority" error={errors.priority?.message}>
          <Input type="number" {...register('priority')} />
        </FormField>
        <FormField label="Weight" error={errors.weight?.message}>
          <Input type="number" {...register('weight')} />
        </FormField>
      </div>

      <div className="flex items-center space-x-2">
        <Controller
          name="is_active"
          control={control}
          render={({ field }) => (
            <Switch
              id="is_active"
              checked={field.value}
              onCheckedChange={field.onChange}
            />
          )}
        />
        <Label htmlFor="is_active">Active</Label>
      </div>
    </FormDialog>
  )
}
