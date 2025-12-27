import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useDeleteProxy } from '../hooks'
import type { Proxy } from '../proxies.types'

interface ProxyDeleteDialogProps {
  proxy: Proxy | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ProxyDeleteDialog({
  proxy,
  open,
  onOpenChange,
}: ProxyDeleteDialogProps) {
  const { mutate, isPending } = useDeleteProxy(() => onOpenChange(false))

  const handleDelete = () => {
    if (!proxy) return
    mutate(proxy.id)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Proxy</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this proxy? This action cannot be
            undone.
          </DialogDescription>
        </DialogHeader>
        {proxy && (
          <div className="rounded-md bg-muted p-4">
            <p className="font-mono text-sm">{proxy.url}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Protocol: {proxy.protocol.toUpperCase()} | Accounts:{' '}
              {proxy.current_accounts}
            </p>
          </div>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={isPending}
          >
            {isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
