import { useMemo } from 'react'
import { IconLoader } from '@tabler/icons-react'
import { Pod } from 'kubernetes-types/core/v1'
import { Link } from 'react-router-dom'

import { MetricsData, PodWithMetrics } from '@/types/api'
import { PodStatus } from '@/types/k8s'
import { getPodStatus } from '@/lib/k8s'
import { formatDate } from '@/lib/utils'

import { MetricCell } from './metrics-cell'
import { PodStatusIcon } from './pod-status-icon'
import { Column, SimpleTable } from './simple-table'
import { Badge } from './ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'

export function PodTable(props: {
  pods?: PodWithMetrics[]
  labelSelector?: string
  isLoading?: boolean
  hiddenNode?: boolean
}) {
  const { pods, isLoading } = props

  // Optimization: Memoize pod status calculation to avoid repeated calls in column accessors.
  // getPodStatus is computationally expensive (O(C) where C is container count).
  // Without this cache, it runs 3 times per row. With cache, it runs 1 time per row.
  const podStatusCache = useMemo(() => {
    const cache = new Map<string, PodStatus>()
    if (pods) {
      pods.forEach((pod) => {
        if (pod.metadata?.uid) {
          cache.set(pod.metadata.uid, getPodStatus(pod))
        }
      })
    }
    return cache
  }, [pods])

  // Pod table columns
  const podColumns = useMemo(
    (): Column<PodWithMetrics>[] => [
      {
        header: 'Name',
        accessor: (pod: Pod) => pod.metadata,
        cell: (value: unknown) => {
          const meta = value as Pod['metadata']
          return (
            <div className="font-medium text-blue-500 hover:underline">
              <Link to={`/pods/${meta!.namespace}/${meta!.name}`}>
                {meta!.name}
              </Link>
            </div>
          )
        },
        align: 'left' as const,
      },
      {
        header: 'Ready',
        accessor: (pod: Pod) => {
          const status =
            podStatusCache.get(pod.metadata?.uid || '') || getPodStatus(pod)
          return `${status.readyContainers} / ${status.totalContainers}`
        },
        cell: (value: unknown) => value as string,
      },
      {
        header: 'Restart',
        accessor: (pod: Pod) => {
          const status =
            podStatusCache.get(pod.metadata?.uid || '') || getPodStatus(pod)
          return status.restartString || '0'
        },
        cell: (value: unknown) => {
          return (
            <span className="text-muted-foreground text-sm">
              {value as number}
            </span>
          )
        },
      },
      {
        header: 'Status',
        accessor: (pod: Pod) => pod,
        cell: (value: unknown) => {
          const pod = value as Pod
          const status =
            podStatusCache.get(pod.metadata?.uid || '') || getPodStatus(pod)
          return (
            <Badge variant="outline" className="text-muted-foreground px-1.5">
              <PodStatusIcon status={status.reason} />
              {status.reason}
            </Badge>
          )
        },
      },
      {
        header: 'CPU',
        accessor: (pod: PodWithMetrics) => {
          return pod.metrics
        },
        cell: (value: unknown) => {
          return <MetricCell type="cpu" metrics={value as MetricsData} />
        },
      },
      {
        header: 'Memory',
        accessor: (pod: PodWithMetrics) => {
          return pod.metrics
        },
        cell: (value: unknown) => {
          return <MetricCell type="memory" metrics={value as MetricsData} />
        },
      },
      {
        header: 'IP',
        accessor: (pod: Pod) => pod.status?.podIP || '-',
        cell: (value: unknown) => (
          <span className="text-sm text-muted-foreground font-mono">
            {value as string}
          </span>
        ),
      },
      ...(props.hiddenNode
        ? []
        : [
            {
              header: 'Node',
              accessor: (pod: Pod) => pod.spec?.nodeName || '-',
              cell: (value: unknown) => (
                <Link
                  to={`/nodes/${value}`}
                  className="text-blue-600 hover:text-blue-800 hover:underline"
                >
                  {value as string}
                </Link>
              ),
            },
          ]),
      {
        header: 'Created',
        accessor: (pod: Pod) => pod.metadata?.creationTimestamp || '',
        cell: (value: unknown) => {
          return (
            <span className="text-muted-foreground text-sm">
              {formatDate(value as string, true)}
            </span>
          )
        },
      },
    ],
    [props.hiddenNode, podStatusCache]
  )

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <IconLoader className="animate-spin mr-2" />
        Loading pods...
      </div>
    )
  }
  return (
    <Card>
      <CardHeader>
        <CardTitle>Pods</CardTitle>
      </CardHeader>
      <CardContent>
        <SimpleTable
          data={pods || []}
          columns={podColumns}
          emptyMessage="No pods found"
          pagination={{
            enabled: true,
            pageSize: 20,
            showPageInfo: true,
          }}
        />
      </CardContent>
    </Card>
  )
}
