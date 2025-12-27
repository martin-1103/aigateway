import { useState } from 'react'
import { Plus, GitBranch } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { Button } from '@/components/ui/button'
import { LoadingSpinner } from '@/components/feedback/loading-spinner'
import { EmptyState } from '@/components/feedback/empty-state'
import { useModelMappingsQuery } from './hooks'
import {
  ModelMappingsTable,
  ModelMappingCreateDialog,
  ModelMappingEditDialog,
  ModelMappingDeleteDialog,
} from './components'
import type { ModelMapping } from './model-mappings.types'

export function ModelMappingsPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editMapping, setEditMapping] = useState<ModelMapping | null>(null)
  const [deleteMapping, setDeleteMapping] = useState<ModelMapping | null>(null)

  const { data, isLoading, error } = useModelMappingsQuery()

  const handleEdit = (mapping: ModelMapping) => {
    setEditMapping(mapping)
  }

  const handleDelete = (mapping: ModelMapping) => {
    setDeleteMapping(mapping)
  }

  const mappings = data?.data ?? []

  return (
    <>
      <PageHeader
        title="Model Mappings"
        description="Create aliases that map to actual AI models"
        actions={
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="h-4 w-4" />
            Add Mapping
          </Button>
        }
      />
      <PageContent>
        {isLoading ? (
          <div className="flex justify-center py-12">
            <LoadingSpinner size="lg" />
          </div>
        ) : error ? (
          <div className="text-center py-12 text-destructive">
            Failed to load model mappings. Please try again.
          </div>
        ) : mappings.length === 0 ? (
          <EmptyState
            icon={GitBranch}
            title="No model mappings"
            description="Create your first model mapping to get started."
            action={
              <Button onClick={() => setCreateDialogOpen(true)}>
                <Plus className="h-4 w-4" />
                Add Mapping
              </Button>
            }
          />
        ) : (
          <ModelMappingsTable
            mappings={mappings}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        )}
      </PageContent>

      <ModelMappingCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
      />
      <ModelMappingEditDialog
        mapping={editMapping}
        open={editMapping !== null}
        onOpenChange={(open) => !open && setEditMapping(null)}
      />
      <ModelMappingDeleteDialog
        mapping={deleteMapping}
        open={deleteMapping !== null}
        onOpenChange={(open) => !open && setDeleteMapping(null)}
      />
    </>
  )
}
