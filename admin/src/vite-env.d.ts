/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APP_PORT: string
  readonly VITE_API_BASE_URL: string
  readonly DEV: boolean
  readonly PROD: boolean
  readonly SSR: boolean
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

