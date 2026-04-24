import { defineConfig } from 'vite'
import inject from '@rollup/plugin-inject'

export default defineConfig({
  build: {
    outDir: 'assets',
    emptyOutDir: true,
    assetsInlineLimit: 0,
    rollupOptions: {
      input: {
        stylesheets: 'web/packs/stylesheets.js',
        application: 'web/packs/application.js',
        index: 'web/packs/index.js',
        useflags: 'web/packs/useflags.js',
      },
      output: {
        entryFileNames: '[name].js',
        chunkFileNames: '[name].js',
        assetFileNames: '[name][extname]',
      },
      plugins: [
        inject({
          $: 'jquery',
          jQuery: 'jquery',
          'window.jQuery': 'jquery',
        }),
      ],
    },
  },
})
