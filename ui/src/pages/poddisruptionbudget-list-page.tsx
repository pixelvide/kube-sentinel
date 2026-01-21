import { useCallback, useMemo } from 'react'
import { createColumnHelper } from '@tanstack/react-table'
import { PodDisruptionBudget } from 'kubernetes-types/policy/v1'
import { Link } from 'react-router-dom'

import { formatDate } from '@/lib/utils'
import { ResourceTable } from '@/components/resource-table'

export function PodDisruptionBudgetListPage() {
  const columnHelper = createColumnHelper<PodDisruptionBudget>()

  const columns = useMemo(
    () => [
      columnHelper.accessor('metadata.name', {
        header: 'Name',
        cell: ({ row }) => (
          <div className="font-medium text-blue-500 hover:underline">
            <Link
              to={`/poddisruptionbudgets/${row.original.metadata!.namespace}/${
                row.original.metadata!.name
              }`}
            >
              {row.original.metadata!.name}
            </Link>
          </div>
        ),
      }),
      columnHelper.accessor((row) => row.spec?.minAvailable, {
        id: 'minAvailable',
        header: 'Min Available',
        cell: ({ getValue }) => {
          const val = getValue()
          return val !== undefined ? val.toString() : '-'
        },
      }),
      columnHelper.accessor((row) => row.spec?.maxUnavailable, {
        id: 'maxUnavailable',
        header: 'Max Unavailable',
        cell: ({ getValue }) => {
          const val = getValue()
          return val !== undefined ? val.toString() : '-'
        },
      }),
      columnHelper.accessor((row) => row.status?.disruptionsAllowed, {
        id: 'disruptionsAllowed',
        header: 'Allowed Disruptions',
        cell: ({ getValue }) => getValue() ?? 0,
      }),
      columnHelper.accessor((row) => row.status?.currentHealthy, {
        id: 'currentHealthy',
        header: 'Current Healthy',
        cell: ({ getValue }) => getValue() ?? 0,
      }),
      columnHelper.accessor((row) => row.status?.desiredHealthy, {
        id: 'desiredHealthy',
        header: 'Desired Healthy',
        cell: ({ getValue }) => getValue() ?? 0,
      }),
      columnHelper.accessor('metadata.creationTimestamp', {
        header: 'Created',
        cell: ({ getValue }) => {
          const dateStr = formatDate(getValue() || '')

          return (
            <span className="text-muted-foreground text-sm">{dateStr}</span>
          )
        },
      }),
    ],
    [columnHelper]
  )

  const filter = useCallback(
    (item: PodDisruptionBudget, query: string) => {
      const queryLower = query.toLowerCase()
      return (
        item.metadata!.name!.toLowerCase().includes(queryLower) ||
        (item.metadata!.namespace?.toLowerCase() || '').includes(queryLower)
      )
    },
    []
  )

  return (
    <ResourceTable
      resourceName="poddisruptionbudgets"
      columns={columns}
      searchQueryFilter={filter}
    />
  )
}
