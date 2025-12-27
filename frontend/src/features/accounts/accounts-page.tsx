import { useState, useCallback } from 'react'
import { Plus } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { Button } from '@/components/ui/button'
import {
  AccountsTable,
  AccountCreateDialog,
  AccountEditDialog,
  AccountDeleteDialog,
} from './components'
import { useAccountsQuery } from './hooks'
import type { Account } from './accounts.types'

export function AccountsPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null)

  const { data, isLoading } = useAccountsQuery()

  const handleEdit = useCallback((account: Account) => {
    setSelectedAccount(account)
    setEditDialogOpen(true)
  }, [])

  const handleDelete = useCallback((account: Account) => {
    setSelectedAccount(account)
    setDeleteDialogOpen(true)
  }, [])

  return (
    <>
      <PageHeader
        title="Accounts"
        description="Manage provider accounts for the AI gateway."
        actions={
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Account
          </Button>
        }
      />

      <PageContent>
        <AccountsTable
          data={data?.data ?? []}
          isLoading={isLoading}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      </PageContent>

      <AccountCreateDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
      />

      <AccountEditDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        account={selectedAccount}
      />

      <AccountDeleteDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        account={selectedAccount}
      />
    </>
  )
}
