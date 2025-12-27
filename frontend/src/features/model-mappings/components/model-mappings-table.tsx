import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Card, CardContent } from '@/components/ui/card'
import { ModelMappingActions, formatDate } from './model-mappings-table-columns'
import type { ModelMapping } from '../model-mappings.types'

interface ModelMappingsTableProps {
  mappings: ModelMapping[]
  onEdit: (mapping: ModelMapping) => void
  onDelete: (mapping: ModelMapping) => void
}

export function ModelMappingsTable({ mappings, onEdit, onDelete }: ModelMappingsTableProps) {
  return (
    <Card>
      <CardContent className="p-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Alias</TableHead>
              <TableHead>Provider</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mappings.map((mapping) => (
              <TableRow key={mapping.id}>
                <TableCell className="font-mono font-medium">{mapping.alias}</TableCell>
                <TableCell className="text-muted-foreground">{mapping.provider_id}</TableCell>
                <TableCell className="font-mono text-muted-foreground">
                  {mapping.model_name}
                </TableCell>
                <TableCell className="text-muted-foreground">{mapping.description}</TableCell>
                <TableCell>
                  <span className={mapping.enabled ? 'text-green-600' : 'text-red-600'}>
                    {mapping.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </TableCell>
                <TableCell>{formatDate(mapping.created_at)}</TableCell>
                <TableCell>
                  <ModelMappingActions
                    mapping={mapping}
                    onEdit={onEdit}
                    onDelete={onDelete}
                  />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
