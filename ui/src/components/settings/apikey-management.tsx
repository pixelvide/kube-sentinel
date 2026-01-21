import { useCallback, useMemo, useState } from 'react'
import {
  IconCheck,
  IconCopy,
  IconKey,
  IconPlus,
  IconShieldCheck,
  IconTrash,
} from '@tabler/icons-react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { PersonalAccessToken } from '@/types/api'
import { createAPIKey, deleteAPIKey, useAPIKeyList } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { DeleteConfirmationDialog } from '@/components/delete-confirmation-dialog'

import { Action, ActionTable } from '../action-table'
import { APIKeyDialog } from './apikey-dialog'
import UserRoleAssignment from './user-role-assignment'

export function APIKeyManagement() {
  const { t } = useTranslation()
  const queryClient = useQueryClient()

  const { data: apiKeys = [], isLoading, error } = useAPIKeyList()

  const [showDialog, setShowDialog] = useState(false)
  const [deletingKey, setDeletingKey] = useState<PersonalAccessToken | null>(
    null
  )
  const [assigningKey, setAssigningKey] = useState<PersonalAccessToken | null>(
    null
  )
  const [generatedToken, setGeneratedToken] = useState<{
    name: string
    token: string
  } | null>(null)
  const [search, setSearch] = useState('')

  const filteredKeys = useMemo(() => {
    if (!search.trim()) return apiKeys
    const s = search.toLowerCase()
    return apiKeys.filter(
      (k) =>
        k.name.toLowerCase().includes(s) ||
        k.user?.username.toLowerCase().includes(s) ||
        String(k.userId).includes(s)
    )
  }, [apiKeys, search])

  const copyToClipboard = useCallback(
    (text: string) => {
      navigator.clipboard.writeText(text)
      toast.success(t('common.copied', 'Copied to clipboard'))
    },
    [t]
  )

  const columns = useMemo<ColumnDef<PersonalAccessToken>[]>(
    () => [
      {
        id: 'name',
        header: t('apikeyManagement.table.name', 'Name'),
        cell: ({ row: { original: apiKey } }) => (
          <div className="flex flex-col">
            <span className="font-medium">{apiKey.name}</span>
            <span className="text-xs text-muted-foreground">
              Owner: {apiKey.user?.username || apiKey.userId}
            </span>
          </div>
        ),
      },
      {
        id: 'prefix',
        header: t('apikeyManagement.table.prefix', 'Prefix'),
        cell: ({ row: { original: apiKey } }) => (
          <code className="text-sm bg-muted px-2 py-1 rounded">
            {apiKey.prefix}...
          </code>
        ),
      },
      {
        id: 'lastUsedAt',
        header: t('apikeyManagement.table.lastActive', 'Last Active'),
        cell: ({ row: { original: apiKey } }) => (
          <div className="flex flex-col gap-1">
            {apiKey.lastUsedAt ? (
              <span className="text-sm">
                {new Date(apiKey.lastUsedAt).toLocaleString()}
              </span>
            ) : (
              <Badge variant="secondary">
                {t('apikeyManagement.neverUsed', 'Never')}
              </Badge>
            )}
            {apiKey.lastUsedIP && (
              <span className="text-xs text-muted-foreground font-mono">
                IP: {apiKey.lastUsedIP.split(',').pop()}
              </span>
            )}
          </div>
        ),
      },
      {
        id: 'expiresAt',
        header: t('apikeyManagement.table.expiresAt', 'Expires At'),
        cell: ({ row: { original: apiKey } }) => {
          if (!apiKey.expiresAt) return <span className="text-sm">-</span>
          const expiry = new Date(apiKey.expiresAt)
          const isExpired = expiry < new Date()
          return (
            <span
              className={`text-sm ${isExpired ? 'text-destructive font-medium' : ''}`}
            >
              {expiry.toLocaleDateString()}
              {isExpired && ` (${t('apikeyManagement.expired', 'Expired')})`}
            </span>
          )
        },
      },
      {
        id: 'createdAt',
        header: t('common.createdAt', 'Created At'),
        cell: ({ row: { original: apiKey } }) => (
          <span className="text-sm text-muted-foreground">
            {new Date(apiKey.createdAt).toLocaleString()}
          </span>
        ),
      },
      {
        id: 'roles',
        header: t('apikeyManagement.table.roles', 'Roles'),
        cell: ({ row: { original: apiKey } }) => (
          <div className="text-sm text-muted-foreground">
            {apiKey.roles?.map((r) => r.name).join(', ') || '-'}
          </div>
        ),
      },
    ],
    [t]
  )

  const actions = useMemo<Action<PersonalAccessToken>[]>(
    () => [
      {
        label: (
          <>
            <IconShieldCheck className="h-4 w-4" />
            {t('common.assign', 'Assign')}
          </>
        ),
        onClick: (apiKey) => setAssigningKey(apiKey),
      },
      {
        label: (
          <div className="inline-flex items-center gap-2 text-destructive">
            <IconTrash className="h-4 w-4" />
            {t('common.delete', 'Delete')}
          </div>
        ),
        onClick: (apiKey) => setDeletingKey(apiKey),
      },
    ],
    [t]
  )

  const createMutation = useMutation({
    mutationFn: createAPIKey,
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['apikey-list'] })
      setShowDialog(false)
      setGeneratedToken({ name: data.apiKey.name, token: data.token })
      toast.success(
        t('apikeyManagement.messages.created', 'API Key created successfully')
      )
    },
    onError: () => {
      toast.error(
        t(
          'apikeyManagement.messages.createError',
          'Failed to create API Key. Please try again.'
        )
      )
    },
  })

  const deleteMutation = useMutation({
    mutationFn: deleteAPIKey,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apikey-list'] })
      setDeletingKey(null)
      toast.success(
        t('apikeyManagement.messages.deleted', 'API Key deleted successfully')
      )
    },
    onError: () => {
      toast.error(
        t(
          'apikeyManagement.messages.deleteError',
          'Failed to delete API Key. Please try again.'
        )
      )
    },
  })

  const handleCreate = useCallback(
    (data: { name: string; expiresAt?: string }) => {
      createMutation.mutate(data)
    },
    [createMutation]
  )

  const handleDelete = useCallback(() => {
    if (deletingKey) {
      deleteMutation.mutate(deletingKey.id)
    }
  }, [deletingKey, deleteMutation])

  if (error) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <p className="text-destructive">
            {t('apikeyManagement.errors.loadFailed', 'Failed to load API Keys')}
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <>
      {generatedToken && (
        <Card className="mb-6 border-primary bg-primary/5">
          <CardHeader>
            <CardTitle className="text-primary flex items-center gap-2">
              <IconCheck className="h-5 w-5" />
              {t(
                'apikeyManagement.generated.title',
                'Key Generated Successfully'
              )}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm">
              {t(
                'apikeyManagement.generated.description',
                "Please copy your API key now. You won't be able to see it again!"
              )}
            </p>
            <div className="flex items-center gap-2 p-3 bg-background border rounded-md font-mono text-sm break-all">
              {generatedToken.token}
              <Button
                variant="ghost"
                size="sm"
                className="ml-auto flex-shrink-0"
                onClick={() => copyToClipboard(generatedToken.token)}
              >
                <IconCopy className="h-4 w-4" />
              </Button>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setGeneratedToken(null)}
            >
              {t('common.dismiss', 'Dismiss')}
            </Button>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center gap-2">
                  <IconKey className="h-5 w-5" />
                  {t('apikeyManagement.title', 'API Key')}
                </CardTitle>
                <p className="text-sm text-muted-foreground mt-1">
                  {t(
                    'apikeyManagement.description',
                    'Manage API keys for programmatic access'
                  )}
                </p>
              </div>
              <Button onClick={() => setShowDialog(true)}>
                <IconPlus className="mr-2 h-4 w-4" />
                {t('apikeyManagement.actions.add', 'Add API Key')}
              </Button>
            </div>
            {apiKeys.length > 0 && (
              <div className="flex items-center gap-2">
                <Input
                  placeholder={t(
                    'apikeyManagement.searchPlaceholder',
                    'Filter by name or owner...'
                  )}
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  className="max-w-sm"
                />
              </div>
            )}
          </div>
        </CardHeader>
        <CardContent>
          {apiKeys.length === 0 && !isLoading ? (
            <div className="text-center py-8 text-muted-foreground">
              <IconKey className="h-12 w-12 mx-auto mb-4 opacity-50" />
              <p className="text-lg font-medium">
                {t('apikeyManagement.empty.title', 'No API keys configured')}
              </p>
              <p className="text-sm">
                {t(
                  'apikeyManagement.empty.description',
                  'Create an API key to get started with programmatic access.'
                )}
              </p>
            </div>
          ) : (
            <ActionTable
              columns={columns}
              data={filteredKeys}
              actions={actions}
            />
          )}
        </CardContent>
      </Card>

      <APIKeyDialog
        open={showDialog}
        onOpenChange={setShowDialog}
        onSubmit={handleCreate}
        isLoading={createMutation.isPending}
      />

      <UserRoleAssignment
        open={!!assigningKey}
        onOpenChange={(open: boolean) => !open && setAssigningKey(null)}
        subject={
          assigningKey
            ? {
                type: 'user',
                name:
                  assigningKey.user?.username || String(assigningKey.userId),
              }
            : undefined
        }
      />

      <DeleteConfirmationDialog
        open={!!deletingKey}
        onOpenChange={(open: boolean) => !open && setDeletingKey(null)}
        onConfirm={handleDelete}
        resourceName={deletingKey?.name || ''}
        resourceType="API Key"
        isDeleting={deleteMutation.isPending}
      />
    </>
  )
}
