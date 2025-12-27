import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useDeleteModelMapping } from '../hooks'
import type { ModelMapping } from '../model-mappings.types'

interface ModelMappingDeleteDialogProps {
  mapping: ModelMapping | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function ModelMappingDeleteDialog({
  mapping,
  open,
  onOpenChange,
}: ModelMappingDeleteDialogProps) {
  const deleteMutation = useDeleteModelMapping({
    onSuccess: () => {
      onOpenChange(false)
    },
  })

  const handleDelete = async () => {
    if (!mapping) return
    await deleteMutation.mutateAsync(mapping.alias)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Model Mapping</DialogTitle>
          <DialogDescription>
            Are you sure you want to delete the mapping for alias{' '}
            <span className="font-mono font-medium text-foreground">"{mapping?.alias}"</span>?
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
