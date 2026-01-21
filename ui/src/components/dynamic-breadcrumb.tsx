import { useTranslation } from 'react-i18next'
import { Link, useLocation } from 'react-router-dom'

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'

interface BreadcrumbSegment {
  label: string
  href?: string
}

export function DynamicBreadcrumb() {
  const location = useLocation()
  const { t } = useTranslation()

  const generateBreadcrumbs = (): BreadcrumbSegment[] => {
    const pathSegments = location.pathname.split('/').filter(Boolean)
    const breadcrumbs: BreadcrumbSegment[] = []

    if (pathSegments.length === 0) {
      return breadcrumbs
    }

    // Resource name mappings
    const resourceLabels: Record<string, string> = {
      pods: t('nav.pods'),
      deployments: t('nav.deployments'),
      services: t('nav.services'),
      configmaps: t('nav.configMaps'),
      secrets: t('nav.secrets'),
      ingresses: t('nav.ingresses'),
      gateways: t('nav.gateways'),
      httproutes: t('nav.httproutes'),
      jobs: t('nav.jobs'),
      daemonsets: t('nav.daemonsets'),
      statefulsets: t('nav.statefulsets'),
      namespaces: t('nav.namespaces'),
      pvcs: t('sidebar.short.pvcs'),
      crds: t('nav.crds'),
      crs: t('nav.customResources'),
      horizontalpodautoscalers: t('nav.horizontalpodautoscalers'),
    }

    // Helper function to create breadcrumb item
    const createBreadcrumb = (
      label: string,
      href?: string
    ): BreadcrumbSegment => ({
      label: resourceLabels[label] || label,
      href,
    })

    // Check if we are in cluster context
    const isClusterContext = pathSegments[0] === 'c'

    pathSegments.forEach((segment, index) => {
      // Skip 'c' segment
      if (isClusterContext && index === 0) return

      // Handle Cluster Name segment
      if (isClusterContext && index === 1) {
        breadcrumbs.push(createBreadcrumb(segment, `/c/${segment}/dashboard`))
        return
      }

      const isLastSegment = index === pathSegments.length - 1
      if (isLastSegment) {
        breadcrumbs.push(createBreadcrumb(segment, undefined))
        return
      }

      let href: string | undefined =
        `/${pathSegments.slice(0, index + 1).join('/')}`

      // Special handling for nested routes
      if (isClusterContext) {
        const relativeIndex = index - 2 // 0 for 'pods', 1 for 'namespace', etc.
        const relativeSegments = pathSegments.slice(2)

        if (relativeSegments[0] === 'crds') {
          if (relativeIndex === 0) href = `/c/${pathSegments[1]}/crds`
          else if (relativeIndex === 1)
            href = `/c/${pathSegments[1]}/crds/${pathSegments[3]}`
          else if (relativeIndex === 2)
            href = `/c/${pathSegments[1]}/crds/${pathSegments[3]}`
          else href = undefined
        } else {
          // Regular resources
          // index 2 is resource (pods), index 3 is namespace
          // (index 4 is resourceName [not linked], checked by isLastSegment above)
          const isNamespace = pathSegments.length === 5 && index === 3
          if (isNamespace) {
            href = `/${pathSegments.slice(0, 3).join('/')}`
          }
        }
      } else {
        // Legacy/Global logic (if any)
        if (pathSegments[0] === 'crds') {
          if (index === 0) href = '/crds'
          else if (index === 1) href = `/crds/${pathSegments[1]}`
          else if (index === 2) href = `/crds/${pathSegments[1]}`
          else href = undefined
        } else {
          const isNamespace = pathSegments.length === 3 && index === 1
          if (isNamespace) href = `/${pathSegments[0]}`
        }
      }

      breadcrumbs.push(createBreadcrumb(segment, href))
    })

    return breadcrumbs
  }

  const breadcrumbs = generateBreadcrumbs()

  return (
    <Breadcrumb className="hidden md:block">
      <BreadcrumbList>
        {breadcrumbs.map((crumb, index) => (
          <div key={index} className="flex items-center">
            {index > 0 && <BreadcrumbSeparator />}
            <BreadcrumbItem>
              {crumb.href ? (
                <BreadcrumbLink asChild>
                  <Link to={crumb.href}>{crumb.label}</Link>
                </BreadcrumbLink>
              ) : (
                <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </div>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
