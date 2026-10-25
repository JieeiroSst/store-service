import { defineConfig, loadEnv, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'

function adsense(client: string): Plugin {
  const valid = /^ca-pub-\d{10,20}$/.test(client)
  return {
    name: 'keikibook-adsense',
    transformIndexHtml() {
      if (!valid) return
      return [
        { tag: 'meta', attrs: { name: 'google-adsense-account', content: client }, injectTo: 'head' },
        {
          tag: 'script',
          attrs: { async: true, src: `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${client}`, crossorigin: 'anonymous' },
          injectTo: 'head',
        },
      ]
    },
    generateBundle() {
      if (!valid) return
      this.emitFile({ type: 'asset', fileName: 'ads.txt', source: `google.com, ${client.replace('ca-', '')}, DIRECT, f08c47fec0942fa0\n` })
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, '.', 'VITE_')
  return {
    plugins: [react(), adsense(env.VITE_ADSENSE_CLIENT ?? '')],
    build: { outDir: 'build', sourcemap: false },
  }
})
