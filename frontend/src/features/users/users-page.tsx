import { useState } from 'react'
import { Plus } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { Button } from '@/components/ui/button'
import {
  UsersTable,
  UserCreateDialog,
  UserEditDialog,
  UserDeleteDialog,
} from './components'
import { useUsersQuery } from './hooks'
import type { User } from './users.types'

export function UsersPage() {
  const { data, isLoading } = useUsersQuery()
  const [createOpen, setCreateOpen] = useState(false)
  const [editUser, setEditUser] = useState<User | null>(null)
  const [deleteUser, setDeleteUser] = useState<User | null>(null)

  return (
    <>
      <PageHeader
        title="Users"
        description="Manage system users and their access roles."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add User
          </Button>
        }
      />
      <PageContent>
        {isLoading ? (
          <div className="flex items-center justify-center h-32">
            <p className="text-muted-foreground">Loading users...</p>
          </div>
        ) : (
          <UsersTable
            data={data?.data ?? []}
            onEdit={setEditUser}
            onDelete={setDeleteUser}
          />
        )}
      </PageContent>

      <UserCreateDialog open={createOpen} onOpenChange={setCreateOpen} />
      <UserEditDialog
        open={!!editUser}
        onOpenChange={(open) => !open && setEditUser(null)}
        user={editUser}
      />
      <UserDeleteDialog
        open={!!deleteUser}
        onOpenChange={(open) => !open && setDeleteUser(null)}
        user={deleteUser}
      />
    </>
  )
}
