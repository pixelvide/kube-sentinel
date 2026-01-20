import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Check, Loader2, Plus, Trash2, X } from 'lucide-react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { UserGitlabConfig } from '@/types/api'

import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog'
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import {
    deleteUserGitlabConfig,
    upsertUserGitlabConfig,
    useGitlabHosts,
    useUserGitlabConfigs,
    validateUserGitlabConfig,
} from '@/lib/api'
import { DeleteConfirmationDialog } from '@/components/delete-confirmation-dialog'

interface ConfigFormValues {
    gitlab_host_id: string
    token: string
}

export function GitlabConfigManagement() {
    const { t } = useTranslation()
    const queryClient = useQueryClient()
    const [isCreateOpen, setIsCreateOpen] = useState(false)
    const [editingConfig, setEditingConfig] = useState<{ id: number; hostId: string; token: string } | null>(null)
    const [deleteConfigId, setDeleteConfigId] = useState<number | null>(null)
    const [isValidating, setIsValidating] = useState<number | null>(null)

    const { data: configs, isLoading } = useUserGitlabConfigs()
    const { data: hosts } = useGitlabHosts()

    const form = useForm<ConfigFormValues>({
        defaultValues: {
            gitlab_host_id: '',
            token: '',
        },
    })

    const onSubmit = async (data: ConfigFormValues) => {
        try {
            await upsertUserGitlabConfig({
                gitlab_host_id: parseInt(data.gitlab_host_id),
                token: data.token,
            })
            toast.success(
                editingConfig
                    ? t('settings.gitlab.updated', 'Configuration updated')
                    : t('settings.gitlab.created', 'Configuration created')
            )
            setIsCreateOpen(false)
            setEditingConfig(null)
            form.reset()
            queryClient.invalidateQueries({ queryKey: ['user-gitlab-configs'] })
        } catch (_error) {
            toast.error(t('settings.gitlab.error', 'Failed to save configuration'))
        }
    }

    const handleDelete = async () => {
        if (!deleteConfigId) return
        try {
            await deleteUserGitlabConfig(deleteConfigId)
            toast.success(t('settings.gitlab.deleted', 'Configuration deleted'))
            setDeleteConfigId(null)
            queryClient.invalidateQueries({ queryKey: ['user-gitlab-configs'] })
        } catch (_error) {
            toast.error(t('settings.gitlab.deleteError', 'Failed to delete configuration'))
        }
    }

    const handleValidate = async (id: number) => {
        setIsValidating(id)
        try {
            await validateUserGitlabConfig(id)
            toast.success(t('settings.gitlab.valid', 'Configuration is valid'))
            queryClient.invalidateQueries({ queryKey: ['user-gitlab-configs'] })
        } catch (_error) {
            toast.error(t('settings.gitlab.invalid', 'Configuration is invalid'))
        } finally {
            setIsValidating(null)
        }
    }

    const openEdit = (config: UserGitlabConfig) => {
        setEditingConfig({
            id: config.id, // Not used in upsert but useful for state
            hostId: config.gitlab_host_id.toString(),
            token: config.token,
        })
        form.reset({
            gitlab_host_id: config.gitlab_host_id.toString(),
            token: config.token,
        })
        setIsCreateOpen(true)
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <div className="space-y-1">
                    <h3 className="text-lg font-medium">{t('settings.gitlab.title', 'GitLab Configuration')}</h3>
                    <p className="text-sm text-muted-foreground">
                        {t('settings.gitlab.subtitle', 'Manage your GitLab tokens for various hosts.')}
                    </p>
                </div>

                <Dialog open={isCreateOpen} onOpenChange={(open) => {
                    setIsCreateOpen(open)
                    if (!open) {
                        setEditingConfig(null)
                        form.reset()
                    }
                }}>
                    <DialogTrigger asChild>
                        <Button>
                            <Plus className="mr-2 h-4 w-4" />
                            {t('settings.gitlab.add', 'Add Configuration')}
                        </Button>
                    </DialogTrigger>
                    <DialogContent>
                        <DialogHeader>
                            <DialogTitle>
                                {editingConfig ? t('settings.gitlab.edit', 'Edit Configuration') : t('settings.gitlab.add', 'Add Configuration')}
                            </DialogTitle>
                            <DialogDescription>
                                {t('settings.gitlab.description', 'Enter your GitLab Personal Access Token.')}
                            </DialogDescription>
                        </DialogHeader>
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                                <FormField
                                    control={form.control}
                                    name="gitlab_host_id"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>{t('settings.gitlab.host', 'GitLab Host')}</FormLabel>
                                            <Select
                                                onValueChange={field.onChange}
                                                defaultValue={field.value}
                                                disabled={!!editingConfig} // Primary key part, maybe shouldn't change
                                            >
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder="Select a host" />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectContent>
                                                    {hosts?.map((host) => (
                                                        <SelectItem key={host.id} value={host.id.toString()}>
                                                            {host.gitlab_host}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="token"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>{t('settings.gitlab.token', 'Token')}</FormLabel>
                                            <FormControl>
                                                <Input type="password" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <DialogFooter>
                                    <Button type="submit">
                                        {t('common.save', 'Save')}
                                    </Button>
                                </DialogFooter>
                            </form>
                        </Form>
                    </DialogContent>
                </Dialog>
            </div>

            <div className="border rounded-md">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>{t('settings.gitlab.host', 'Host')}</TableHead>
                            <TableHead>{t('settings.gitlab.status', 'Status')}</TableHead>
                            <TableHead className="w-[150px] text-right">{t('common.actions', 'Actions')}</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {isLoading ? (
                            <TableRow>
                                <TableCell colSpan={3} className="text-center py-4">
                                    <Loader2 className="h-6 w-6 animate-spin mx-auto" />
                                </TableCell>
                            </TableRow>
                        ) : configs?.length === 0 ? (
                            <TableRow>
                                <TableCell colSpan={3} className="text-center h-24 text-muted-foreground">
                                    {t('settings.gitlab.empty', 'No configurations found')}
                                </TableCell>
                            </TableRow>
                        ) : (
                            configs?.map((config) => (
                                <TableRow key={config.id}>
                                    <TableCell className="font-medium">
                                        {config.gitlab_host.gitlab_host}
                                    </TableCell>
                                    <TableCell>
                                        {config.is_validated ? (
                                            <Badge variant="default" className="bg-green-500/15 text-green-700 hover:bg-green-500/25 border-green-200">
                                                <Check className="mr-1 h-3 w-3" />
                                                {t('settings.gitlab.validated', 'Validated')}
                                            </Badge>
                                        ) : (
                                            <Badge variant="secondary" className="text-muted-foreground">
                                                <X className="mr-1 h-3 w-3" />
                                                {t('settings.gitlab.unvalidated', 'Not Validated')}
                                            </Badge>
                                        )}
                                    </TableCell>
                                    <TableCell className="text-right space-x-2">
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => handleValidate(config.id)}
                                            disabled={isValidating === config.id}
                                        >
                                            {isValidating === config.id ? (
                                                <Loader2 className="h-3 w-3 animate-spin" />
                                            ) : config.is_validated ? (
                                                t('common.revalidate', 'Re-validate')
                                            ) : (
                                                t('common.validate', 'Validate')
                                            )}
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            onClick={() => openEdit(config)}
                                        >
                                            <span className="sr-only">Edit</span>
                                            <svg
                                                width="15"
                                                height="15"
                                                viewBox="0 0 15 15"
                                                fill="none"
                                                xmlns="http://www.w3.org/2000/svg"
                                                className="h-4 w-4"
                                            >
                                                <path
                                                    d="M11.8536 1.14645C11.6583 0.951184 11.3417 0.951184 11.1464 1.14645L3.71455 8.57829C3.64582 8.64701 3.59754 8.73579 3.57539 8.8319L3.00792 11.2372L5.39808 11.801L6.68069 6.42787L11.8536 1.14645ZM6 13.5L1.5 13.5V9L4.5 9L6 13.5ZM13.8536 3.14645C14.0488 2.95118 14.0488 2.63461 13.8536 2.43934L12.5607 1.14645C12.3654 0.951184 12.0488 0.951184 11.8536 1.14645L12.8536 2.14645L13.8536 3.14645Z"
                                                    fill="currentColor"
                                                    fillRule="evenodd"
                                                    clipRule="evenodd"
                                                ></path>
                                            </svg>
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="text-destructive hover:text-destructive"
                                            onClick={() => setDeleteConfigId(config.id)}
                                        >
                                            <Trash2 className="h-4 w-4" />
                                        </Button>
                                    </TableCell>
                                </TableRow>
                            ))
                        )}
                    </TableBody>
                </Table>
            </div>

            <DeleteConfirmationDialog
                open={!!deleteConfigId}
                onOpenChange={(open) => !open && setDeleteConfigId(null)}
                onConfirm={handleDelete}
                title={t('settings.gitlab.deleteTitle', 'Delete Configuration')}
                description={t(
                    'settings.gitlab.deleteDescription',
                    'Are you sure you want to delete this GitLab configuration? This action cannot be undone.'
                )}
            />
        </div>
    )
}
