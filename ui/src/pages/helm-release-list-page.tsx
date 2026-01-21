import { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  ColumnDef,
  getCoreRowModel,
  getPaginationRowModel,
  getFilteredRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { Search } from 'lucide-react'

import { HelmRelease, useHelmReleases } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ResourceTableView } from '@/components/resource-table-view'
import { ErrorMessage } from '@/components/error-message'
import { NamespaceSelector } from '@/components/selector/namespace-selector'

export function HelmReleaseListPage() {
  const { t } = useTranslation()
  const [selectedNamespace, setSelectedNamespace] = useState<string | undefined>('_all')
  const [searchQuery, setSearchQuery] = useState('')
  const [pagination, setPagination] = useState({
    pageIndex: 0,
    pageSize: 10,
  })

  const { data: releases = [], isLoading, error, refetch } = useHelmReleases(
    selectedNamespace
  )

  const columns = useMemo<ColumnDef<HelmRelease>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('common.name'),
        cell: (info) => <div className="font-medium">{info.getValue() as string}</div>,
      },
      {
        accessorKey: 'namespace',
        header: t('common.namespace'),
        cell: (info) => (
          <Badge variant="outline" className="ml-2">
            {info.getValue() as string}
          </Badge>
        ),
      },
      {
        accessorKey: 'revision',
        header: 'Revision',
      },
      {
        accessorKey: 'status',
        header: t('common.status'),
        cell: (info) => (
          <Badge variant={info.getValue() === 'deployed' ? 'default' : 'secondary'}>
            {info.getValue() as string}
          </Badge>
        ),
      },
      {
        accessorKey: 'chart',
        header: 'Chart',
      },
      {
        accessorKey: 'app_version',
        header: 'App Version',
      },
      {
        accessorKey: 'updated',
        header: 'Updated',
      },
    ],
    [t]
  )

  const filteredData = useMemo(() => {
    if (!searchQuery) return releases
    return releases.filter(
      (item) =>
        item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.chart.toLowerCase().includes(searchQuery.toLowerCase())
    )
  }, [releases, searchQuery])

  const table = useReactTable({
    data: filteredData,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    state: {
      pagination,
    },
    onPaginationChange: setPagination,
  })

  const handleNamespaceChange = (value: string) => {
    setSelectedNamespace(value)
  }

  if (error) {
    return <ErrorMessage error={error} refetch={refetch} />
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Helm Releases</h1>
        </div>
        <div className="flex items-center gap-2">
          <NamespaceSelector
            selectedNamespace={selectedNamespace}
            handleNamespaceChange={handleNamespaceChange}
            showAll={true}
          />
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search releases..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 w-[200px]"
            />
          </div>
          <Button variant="outline" onClick={() => refetch()}>
            {t('common.refresh')}
          </Button>
        </div>
      </div>

      <ResourceTableView
        table={table}
        columnCount={columns.length}
        isLoading={isLoading}
        data={filteredData}
        emptyState={
          <div className="text-center p-8 text-muted-foreground">
            No releases found
          </div>
        }
        hasActiveFilters={!!searchQuery}
        filteredRowCount={filteredData.length}
        totalRowCount={releases.length}
        searchQuery={searchQuery}
        pagination={pagination}
        setPagination={setPagination}
      />
    </div>
  )
}
