import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Copy, Check } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateApiKey } from '../hooks'
import { createApiKeySchema, type CreateApiKeyFormData } from '../api-keys.schemas'
import type { CreateApiKeyResponse } from '../api-keys.types'

interface ApiKeyCreateDialogProps {
  open: boolean
  onClose: () => void
}

export function ApiKeyCreateDialog({ open, onClose }: ApiKeyCreateDialogProps) {
  const { mutate, isPending } = useCreateApiKey()
  const [createdKey, setCreatedKey] = useState<CreateApiKeyResponse | null>(null)
  const [copied, setCopied] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateApiKeyFormData>({
    resolver: zodResolver(createApiKeySchema),
  })

  const handleClose = () => {
    reset()
    setCreatedKey(null)
    setCopied(false)
    onClose()
  }

  const onSubmit = (data: CreateApiKeyFormData) => {
    mutate(data, {
      onSuccess: (response) => {
        setCreatedKey(response)
      },
    })
  }

  const copyToClipboard = async () => {
    if (!createdKey) return
    await navigator.clipboard.writeText(createdKey.key)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {createdKey ? 'API Key Created' : 'Create API Key'}
          </DialogTitle>
          <DialogDescription>
            {createdKey
              ? 'Copy your API key now. You will not be able to see it again.'
              : 'Enter a name for your new API key.'}
          </DialogDescription>
        </DialogHeader>

        {createdKey ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>API Key</Label>
              <div className="flex items-center gap-2">
                <code className="flex-1 px-3 py-2 bg-muted rounded text-sm break-all">
                  {createdKey.key}
                </code>
                <Button
                  variant="outline"
                  size="icon"
                  onClick={copyToClipboard}
                  aria-label="Copy API key"
                >
                  {copied ? (
                    <Check className="h-4 w-4 text-green-600" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
            <DialogFooter>
              <Button onClick={handleClose}>Done</Button>
            </DialogFooter>
          </div>
        ) : (
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                placeholder="My API Key"
                {...register('name')}
              />
              {errors.name && (
                <p className="text-sm text-destructive">{errors.name.message}</p>
              )}
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button type="submit" disabled={isPending}>
                {isPending ? 'Creating...' : 'Create'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
