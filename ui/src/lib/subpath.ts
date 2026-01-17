declare global {
    interface Window {
        __dynamic_base__?: string
    }
}

export function getSubPath(): string {
    if (import.meta.env.DEV) {
        return (import.meta.env.VITE_CLOUD_SENTINEL_K8S_BASE as string) || ''
    }
    return window.__dynamic_base__ || ''
}

export function withSubPath(path: string): string {
    const subPath = getSubPath()
    if (!subPath) return path

    if (path.startsWith('/')) {
        return `${subPath}${path}`
    }
    return `${subPath}/${path}`
}
