import { AlertCircle, AlertTriangle, Info, Loader2 } from 'lucide-react'

import { Anomaly } from '@/types/api'
import { useResourceAnalysis } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

interface AnomalyTableProps {
  anomalies: Anomaly[]
}

const severityConfig = {
  critical: {
    variant: 'destructive',
    icon: AlertCircle,
    className:
      'bg-red-100 text-red-800 hover:bg-red-100/80 border-red-200 dark:bg-red-900/30 dark:text-red-300 dark:border-red-800',
  },
  high: {
    variant: 'destructive',
    icon: AlertCircle,
    className:
      'bg-orange-100 text-orange-800 hover:bg-orange-100/80 border-orange-200 dark:bg-orange-900/30 dark:text-orange-300 dark:border-orange-800',
  },
  medium: {
    variant: 'outline',
    icon: AlertTriangle,
    className:
      'bg-yellow-100 text-yellow-800 hover:bg-yellow-100/80 border-yellow-200 dark:bg-yellow-900/30 dark:text-yellow-300 dark:border-yellow-800',
  },
  low: {
    variant: 'outline',
    icon: Info,
    className:
      'bg-blue-100 text-blue-800 hover:bg-blue-100/80 border-blue-200 dark:bg-blue-900/30 dark:text-blue-300 dark:border-blue-800',
  },
  info: {
    variant: 'outline',
    icon: Info,
    className:
      'bg-gray-100 text-gray-800 hover:bg-gray-100/80 border-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700',
  },
} as const

export function AnomalyTable({ anomalies }: AnomalyTableProps) {
  if (!anomalies || anomalies.length === 0) {
    return (
      <Card>
        <CardContent className="pt-6 text-center text-muted-foreground">
          No anomalies detected.
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Detected Anomalies ({anomalies.length})</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[100px]">Severity</TableHead>
              <TableHead className="w-[200px]">Check</TableHead>
              <TableHead>Message</TableHead>
              <TableHead>Remediation</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {anomalies.map((anomaly, index) => {
              // Fallback for safety
              const config =
                severityConfig[anomaly.severity] || severityConfig.info
              const Icon = config.icon

              return (
                <TableRow key={index}>
                  <TableCell>
                    <Badge
                      variant="outline"
                      className={`flex w-fit items-center gap-1 ${config.className}`}
                    >
                      <Icon className="h-3 w-3" />
                      <span className="capitalize">{anomaly.severity}</span>
                    </Badge>
                  </TableCell>
                  <TableCell className="font-medium">{anomaly.title}</TableCell>
                  <TableCell>{anomaly.message}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">
                    {anomaly.remediation || '-'}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}

export function ResourceAnomalies(props: {
  resourceType: string
  name: string
  namespace?: string
}) {
  const {
    data: analysis,
    isLoading,
    error,
  } = useResourceAnalysis(props.resourceType, props.name, props.namespace)

  if (isLoading) {
    return (
      <div className="flex h-32 items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <Card>
        <CardContent className="pt-6 text-center text-destructive">
          Failed to load anomalies: {(error as Error).message}
        </CardContent>
      </Card>
    )
  }

  return <AnomalyTable anomalies={analysis?.anomalies || []} />
}
