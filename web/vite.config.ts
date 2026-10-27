import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/health': 'http://localhost:8080',
      '/catalog': 'http://localhost:8080',
      '/updates': 'http://localhost:8080',
      '/demo': 'http://localhost:8080',
    },
  },
  test: { environment: 'jsdom', setupFiles: './src/testSetup.ts' },
})
