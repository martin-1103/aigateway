import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useRevokeApiKey } from '../hooks'
import type { ApiKey } from '../api-keys.types'

interface ApiKeyRevokeDialogProps {
  apiKey: ApiKey | null
  onClose: () => void
}

export function ApiKeyRevokeDialog({ apiKey, onClose }: ApiKeyRevokeDialogProps) {
  const { mutate, isPending } = useRevokeApiKey()

  const handleRevoke = () => {
    if (!apiKey) return
    mutate(apiKey.id, {
      onSuccess: () => {
        onClose()
      },
    })
  }

  return (
    <Dialog open={!!apiKey} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke API Key</DialogTitle>
          <DialogDescription>
            Are you sure you want to revoke the API key "{apiKey?.name}"? This
            action cannot be undone and any applications using this key will
            stop working.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleRevoke}
            disabled={isPending}
          >
            {isPending ? 'Revoking...' : 'Revoke'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
