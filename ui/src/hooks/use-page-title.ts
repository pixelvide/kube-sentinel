import { useEffect } from 'react'

export function usePageTitle(title: string) {
  useEffect(() => {
    const previousTitle = document.title

    if (title) {
      document.title = `${title} - Kube Sentinel`
    }

    return () => {
      document.title = previousTitle
    }
  }, [title])
}
