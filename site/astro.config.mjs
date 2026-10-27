import { defineConfig } from 'astro/config';

export default defineConfig({
  site: 'https://proxyglass.app',
  output: 'static',
  trailingSlash: 'always',
  vite: {
    build: {
      // Keep all assets as files so Workers static assets and strict CSP can
      // serve them locally without inline base64 assets.
      assetsInlineLimit: 0,
    },
  },
});
