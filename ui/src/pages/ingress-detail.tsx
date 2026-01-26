
import { useState } from 'react'
import {
    IconLoader,
    IconRefresh,
    IconTrash,
    IconExternalLink,
} from '@tabler/icons-react'
import * as yaml from 'js-yaml'
import { Ingress } from 'kubernetes-types/networking/v1'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import {
    updateResource,
    useResource,
    useResourceAnalysis,
} from '@/lib/api'
import { formatDate, translateError } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { ResponsiveTabs } from '@/components/ui/responsive-tabs'
import { DescribeDialog } from '@/components/describe-dialog'
import { ErrorMessage } from '@/components/error-message'
import { EventTable } from '@/components/event-table'
import { LabelsAnno } from '@/components/lables-anno'
import { RelatedResourcesTable } from '@/components/related-resource-table'
import { ResourceDeleteConfirmationDialog } from '@/components/resource-delete-confirmation-dialog'
import { YamlEditor } from '@/components/yaml-editor'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { ResourceAnomalies } from '@/components/anomaly-table'
import { ResourceHistoryTable } from '@/components/resource-history-table'
import { SecurityTab } from '@/components/security/security-tab'

export function IngressDetail(props: { namespace: string; name: string }) {
    const { namespace, name } = props
    const [yamlContent, setYamlContent] = useState('')
    const [isSavingYaml, setIsSavingYaml] = useState(false)
    const [refreshKey, setRefreshKey] = useState(0)
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
    const { t } = useTranslation()

    const {
        data: ingress,
        isLoading,
        isError,
        error,
        refetch,
    } = useResource('ingresses', name, namespace)

    const { data: analysis } = useResourceAnalysis('ingresses', name, namespace)

    const handleRefresh = () => {
        setRefreshKey((prev) => prev + 1)
        refetch()
    }

    const handleSaveYaml = async (content: Ingress) => {
        setIsSavingYaml(true)
        try {
            await updateResource('ingresses', name, namespace, content)
            toast.success('YAML saved successfully')
        } catch (error) {
            console.error('Failed to save YAML:', error)
            toast.error(translateError(error, t))
        } finally {
            setIsSavingYaml(false)
        }
    }

    const handleYamlChange = (content: string) => {
        setYamlContent(content)
    }

    if (isLoading) {
        return (
            <div className="p-6">
                <Card>
                    <CardContent className="pt-6">
                        <div className="flex items-center justify-center gap-2">
                            <IconLoader className="animate-spin" />
                            <span>Loading ingress details...</span>
                        </div>
                    </CardContent>
                </Card>
            </div>
        )
    }

    if (isError || !ingress) {
        return (
            <ErrorMessage
                resourceName={'Ingress'}
                error={error}
                refetch={handleRefresh}
            />
        )
    }

    return (
        <div className="space-y-2">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-lg font-bold">{name}</h1>
                    <p className="text-muted-foreground">
                        Namespace: <span className="font-medium">{namespace}</span>
                    </p>
                </div>
                <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={handleRefresh}>
                        <IconRefresh className="w-4 h-4" />
                        Refresh
                    </Button>
                    <DescribeDialog
                        resourceType="ingresses"
                        namespace={namespace}
                        name={name}
                    />
                    <Button
                        variant="destructive"
                        size="sm"
                        onClick={() => setIsDeleteDialogOpen(true)}
                    >
                        <IconTrash className="w-4 h-4" />
                        Delete
                    </Button>
                </div>
            </div>

            <ResponsiveTabs
                tabs={[
                    {
                        value: 'overview',
                        label: 'Overview',
                        content: (
                            <div className="space-y-4">
                                <Card>
                                    <CardHeader>
                                        <CardTitle>Ingress Information</CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                                            <div>
                                                <Label className="text-xs text-muted-foreground">
                                                    Created
                                                </Label>
                                                <p className="text-sm">
                                                    {formatDate(
                                                        ingress.metadata?.creationTimestamp || '',
                                                        true
                                                    )}
                                                </p>
                                            </div>
                                            <div>
                                                <Label className="text-xs text-muted-foreground">
                                                    Class
                                                </Label>
                                                <p className="text-sm">
                                                    {ingress.spec?.ingressClassName || 'N/A'}
                                                </p>
                                            </div>
                                            <div>
                                                <Label className="text-xs text-muted-foreground">
                                                    Load Balancer
                                                </Label>
                                                <p className="text-sm">
                                                    {ingress.status?.loadBalancer?.ingress
                                                        ?.map((i) => i.ip || i.hostname)
                                                        .join(', ') || 'N/A'}
                                                </p>
                                            </div>
                                        </div>
                                        <LabelsAnno
                                            labels={ingress.metadata?.labels || {}}
                                            annotations={ingress.metadata?.annotations || {}}
                                        />
                                    </CardContent>
                                </Card>

                                <Card>
                                    <CardHeader>
                                        <CardTitle>Rules</CardTitle>
                                    </CardHeader>
                                    <CardContent>
                                        <Table>
                                            <TableHeader>
                                                <TableRow>
                                                    <TableHead>Host</TableHead>
                                                    <TableHead>Path</TableHead>
                                                    <TableHead>Service</TableHead>
                                                    <TableHead>Port</TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {ingress.spec?.rules?.map((rule, ruleIndex) => (
                                                    rule.http?.paths.map((path, pathIndex) => {
                                                        const isTLS = ingress.spec?.tls?.some((t) =>
                                                            t.hosts?.some((h) => h === rule.host)
                                                        )
                                                        const protocol = isTLS ? 'https' : 'http'
                                                        const url = rule.host ? `${protocol}://${rule.host}${path.path}` : undefined

                                                        return (
                                                            <TableRow key={`${ruleIndex}-${pathIndex}`}>
                                                                <TableCell className="font-medium">
                                                                    {url ? (
                                                                        <a
                                                                            href={url}
                                                                            target="_blank"
                                                                            rel="noopener noreferrer"
                                                                            className="text-blue-500 hover:underline flex items-center gap-1"
                                                                        >
                                                                            {rule.host}
                                                                            <IconExternalLink className="w-3 h-3" />
                                                                        </a>
                                                                    ) : (
                                                                        rule.host || '*'
                                                                    )}
                                                                </TableCell>
                                                                <TableCell>{path.path}</TableCell>
                                                                <TableCell>{path.backend?.service?.name || 'N/A'}</TableCell>
                                                                <TableCell>{path.backend?.service?.port?.number || path.backend?.service?.port?.name || 'N/A'}</TableCell>
                                                            </TableRow>
                                                        )
                                                    })
                                                ))}
                                                {(!ingress.spec?.rules || ingress.spec.rules.length === 0) && (
                                                    <TableRow>
                                                        <TableCell colSpan={4} className="text-center text-muted-foreground">
                                                            No rules defined
                                                        </TableCell>
                                                    </TableRow>
                                                )}
                                            </TableBody>
                                        </Table>
                                    </CardContent>
                                </Card>
                            </div>
                        ),
                    },
                    {
                        value: 'yaml',
                        label: 'YAML',
                        content: (
                            <YamlEditor<'ingresses'>
                                key={refreshKey}
                                value={yamlContent || yaml.dump(ingress, { indent: 2 })}
                                title="YAML Configuration"
                                onSave={handleSaveYaml}
                                onChange={handleYamlChange}
                                isSaving={isSavingYaml}
                            />
                        ),
                    },
                    {
                        value: 'events',
                        label: 'Events',
                        content: (
                            <EventTable
                                resource="ingresses"
                                name={name}
                                namespace={namespace}
                            />
                        ),
                    },
                    {
                        value: 'history',
                        label: 'History',
                        content: (
                            <ResourceHistoryTable
                                resourceType="ingresses"
                                name={name}
                                namespace={namespace}
                                currentResource={ingress}
                            />
                        ),
                    },
                    {
                        value: 'related',
                        label: 'Related',
                        content: (
                            <RelatedResourcesTable
                                resource="ingresses"
                                name={name}
                                namespace={namespace}
                            />
                        ),
                    },
                    {
                        value: 'security',
                        label: 'Security',
                        content: (
                            <SecurityTab namespace={namespace} kind="Ingress" name={name} />
                        ),
                    },
                    {
                        value: 'anomalies',
                        label: (
                            <>
                                Anomalies
                                {analysis?.anomalies && analysis.anomalies.length > 0 && (
                                    <Badge
                                        variant="secondary"
                                        className="ml-1 bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300"
                                    >
                                        {analysis.anomalies.length}
                                    </Badge>
                                )}
                            </>
                        ),
                        content: (
                            <ResourceAnomalies
                                resourceType="ingresses"
                                name={name}
                                namespace={namespace}
                            />
                        ),
                    },
                ]}
            />

            <ResourceDeleteConfirmationDialog
                open={isDeleteDialogOpen}
                onOpenChange={setIsDeleteDialogOpen}
                resourceName={name}
                resourceType="ingresses"
                namespace={namespace}
            />
        </div>
    )
}
