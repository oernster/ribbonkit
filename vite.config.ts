/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The kit's web half is tested here on its own; an application tests its own page with its own
// config, naming the kit's setup file (@oernster/ribbonkit/test-setup) so both run under one setup.
export default defineConfig({
  plugins: [react()],
  test: {
    include: ['web/**/*.test.{ts,tsx}'],
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./web/testing/setup.ts'],
  },
})
