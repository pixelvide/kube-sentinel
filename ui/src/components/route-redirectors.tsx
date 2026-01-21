import { Navigate, useLocation } from 'react-router-dom'

import { useCluster } from '@/hooks/use-cluster'

export function RootRedirector() {
  const { currentCluster } = useCluster()
  // Wait for cluster to be resolved
  if (!currentCluster) {
    return null // generic loading or let the App's loading state handle it
  }
  return <Navigate to={`/c/${currentCluster}/dashboard`} replace />
}

export function ClusterRedirector() {
  const { currentCluster } = useCluster()
  const location = useLocation()

  if (!currentCluster) {
    return null
  }

  // Preserve the current path but prepend cluster
  // e.g. /pods -> /c/dev/pods
  return <Navigate to={`/c/${currentCluster}${location.pathname}`} replace />
}
