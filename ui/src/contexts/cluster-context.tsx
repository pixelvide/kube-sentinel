/* eslint-disable react-refresh/only-export-components */
import React, { createContext, useEffect, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

import { Cluster } from '@/types/api'
import { apiClient } from '@/lib/api-client'
import { withSubPath } from '@/lib/subpath'

import { useAuth } from './auth-context'

interface ClusterContextType {
  clusters: Cluster[]
  currentCluster: string | null
  setCurrentCluster: (clusterName: string) => void
  isLoading: boolean
  isSwitching?: boolean
  error: Error | null
}

export const ClusterContext = createContext<ClusterContextType | undefined>(
  undefined
)

export const ClusterProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const params = useParams<{ cluster?: string }>()
  const navigate = useNavigate()
  const location = useLocation()
  const queryClient = useQueryClient()
  const [isSwitching, setIsSwitching] = useState(false)
  const { user, isLoading: isAuthLoading } = useAuth()

  // Determine current cluster: URL param usually takes precedence, but if missing (global routes), use storage
  const urlCluster = params.cluster
  const [storedCluster, setStoredCluster] = useState<string | null>(
    localStorage.getItem('current-cluster')
  )

  const currentCluster = urlCluster || storedCluster

  // Fetch clusters from API (this request shouldn't need cluster header)
  const {
    data: clusters = [],
    isLoading,
    error,
  } = useQuery<Cluster[]>({
    queryKey: ['clusters'],
    queryFn: async () => {
      const response = await fetch(withSubPath('/api/v1/clusters'), {
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (response.status === 403) {
        const errorData = await response.json().catch(() => ({}))
        const redirectUrl = response.headers.get('Location')
        if (redirectUrl) {
          window.location.href = redirectUrl
        }
        throw new Error(`${errorData.error || response.status}`)
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(`${errorData.error || response.status}`)
      }

      return response.json()
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  })

  // Sync state with validation
  useEffect(() => {
    if (clusters.length > 0) {
      if (!currentCluster) {
        // No cluster selected or in URL -> Find default
        const defaultCluster = clusters.find((c) => c.isDefault) || clusters[0]
        setStoredCluster(defaultCluster.name)
        localStorage.setItem('current-cluster', defaultCluster.name)
        // We don't force navigate here, the RootRedirector or user action will handle it
      } else if (!clusters.some((c) => c.name === currentCluster)) {
        // Current cluster (URL or Stored) is invalid
        if (storedCluster === currentCluster) {
          setStoredCluster(null)
          localStorage.removeItem('current-cluster')
        }

        // If URL has an invalid cluster, redirect to default
        if (urlCluster) {
          const defaultCluster =
            clusters.find((c) => c.isDefault) || clusters[0]
          if (defaultCluster) {
            navigate(`/c/${defaultCluster.name}/dashboard`, { replace: true })
          }
        }
      } else {
        // Valid current cluster, ensure cookies/storage match
        if (currentCluster !== storedCluster) {
          setStoredCluster(currentCluster)
          localStorage.setItem('current-cluster', currentCluster)
        }
        document.cookie = `x-cluster-name=${currentCluster}; path=/`
      }
    }
  }, [clusters, currentCluster, storedCluster, navigate, urlCluster])

  // If admin is logged in and no clusters are found, redirect to settings
  useEffect(() => {
    if (
      !isLoading &&
      !isAuthLoading &&
      clusters.length === 0 &&
      user?.isAdmin()
    ) {
      // Avoid infinite redirect loop if already on settings
      if (!location.pathname.startsWith('/settings')) {
        navigate('/settings?tab=clusters')
      }
    }
  }, [isLoading, isAuthLoading, clusters, user, location.pathname, navigate])

  const setCurrentCluster = (clusterName: string) => {
    if (clusterName === currentCluster) return

    try {
      setIsSwitching(true)

      // Update storage/cookies immediately
      localStorage.setItem('current-cluster', clusterName)
      setStoredCluster(clusterName)
      document.cookie = `x-cluster-name=${clusterName}; path=/`

      // Explicitly update API client provider to use the new cluster for pending requests (invalidation)
      apiClient.setClusterProvider(() => clusterName)

      setTimeout(async () => {
        await queryClient.invalidateQueries({
          predicate: (query) => {
            const key = query.queryKey[0] as string
            return !['user', 'auth', 'clusters'].includes(key)
          },
        })

        setIsSwitching(false)
        toast.success(`Switched to cluster: ${clusterName}`, {
          id: 'cluster-switch',
        })

        // Navigate
        if (location.pathname.startsWith('/c/')) {
          // Replace cluster part
          // /c/old/foo -> /c/new/foo
          const newPath = location.pathname.replace(
            /^\/c\/[^/]+/,
            `/c/${clusterName}`
          )
          navigate(newPath)
        } else {
          // From global page -> go to dashboard of new cluster
          navigate(`/c/${clusterName}/dashboard`)
        }
      }, 300)
    } catch (error) {
      console.error('Failed to switch cluster:', error)
      setIsSwitching(false)
      toast.error('Failed to switch cluster', {
        id: 'cluster-switch',
      })
    }
  }

  const value: ClusterContextType = {
    clusters,
    currentCluster,
    setCurrentCluster,
    isLoading,
    isSwitching,
    error: error as Error | null,
  }

  return (
    <ClusterContext.Provider value={value}>{children}</ClusterContext.Provider>
  )
}
