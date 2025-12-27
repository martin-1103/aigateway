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
              <TableHead>Target Model</TableHead>
              <TableHead>Created</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="w-[70px]">
                <span className="sr-only">Actions</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {mappings.map((mapping) => (
              <TableRow key={mapping.alias}>
                <TableCell className="font-mono font-medium">{mapping.alias}</TableCell>
                <TableCell className="font-mono text-muted-foreground">
                  {mapping.target_model}
                </TableCell>
                <TableCell>{formatDate(mapping.created_at)}</TableCell>
                <TableCell>{formatDate(mapping.updated_at)}</TableCell>
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
