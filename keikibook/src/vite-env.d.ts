/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_APK_URL?: string
  readonly VITE_ADSENSE_CLIENT?: string
  readonly VITE_ADSENSE_SLOT?: string
}

interface Window {
  adsbygoogle?: unknown[]
}
