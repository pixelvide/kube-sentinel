/// <reference types="vite/client" />

interface ImportMetaEnv {
    readonly VITE_CLOUD_SENTINEL_K8S_BASE: string
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}
