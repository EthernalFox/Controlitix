import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  build: {
    rollupOptions: {
      input: {
        main: 'index.html',
      },
      output: {
        assetFileNames: chunkInfo => {
          const isCss = chunkInfo.name?.includes('.css')
          return isCss ? `assets/[name]-[hash].[ext]` : `assets/[name].[ext]`
        },
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@modules': path.resolve(__dirname, './src/shared/modules'),
      '@api': path.resolve(__dirname, './src/shared/modules/Net/Api'),
      '@socket': path.resolve(__dirname, './src/shared/modules/Net/Socket'),
      '@utils': path.resolve(__dirname, './src/shared/utils'),
      '@styles': path.resolve(__dirname, './src/styles'),
      '@app': path.resolve(__dirname, './src/app'),
      '@entities': path.resolve(__dirname, './src/entities'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@features': path.resolve(__dirname, './src/features'),
      '@widgets': path.resolve(__dirname, './src/widgets'),
      '@shared': path.resolve(__dirname, './src/shared'),
      '@store': path.resolve(__dirname, './src/shared/store'),
    },
  },
  plugins: [react()],

})
