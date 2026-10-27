# Proxyglass website

Astro static marketing site for [proxyglass.app](https://proxyglass.app), styled to match the Proxyglass dashboard. It has no Easter eggs or external runtime asset dependencies.

## Local development

Requires Node.js 22.12+.

```sh
npm ci
npm run dev -- --background
```

The Astro dev server opens at <http://localhost:4321>. Manage it with `npx astro dev status`, `npx astro dev logs` and `npx astro dev stop`.

## Build and check

```sh
npm run build
npm run smoke
npx wrangler deploy --dry-run
```

Astro writes a static site to `dist/`. Wrangler serves those generated assets from a Cloudflare Worker; the Worker has no application data or server-side rendering.

## Preview the Cloudflare Worker locally

```sh
npm run build
npm run worker:dev -- --local --ip 127.0.0.1 --port 8788
```

Open <http://127.0.0.1:8788>. The Worker uses the same static asset binding as production.

## Deploy to `proxyglass.app`

1. Add `proxyglass.app` to the Cloudflare account and ensure its DNS is active there.
2. Create an API token with permission to deploy Workers and edit the zone's Worker routes.
3. Set `CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` in the environment, then run:

   ```sh
   npm run deploy
   ```

`site/wrangler.jsonc` attaches the Worker to `proxyglass.app/*` and `www.proxyglass.app/*`, deploys `dist/` as static assets, and returns the generated `404.html` for missing paths. TLS and edge delivery are handled by Cloudflare.

The GitHub Actions workflow `.github/workflows/site-deploy.yml` supports manual deploys. To deploy automatically on pushes to `main`, configure the two secrets above and set the repository variable `PROXYGLASS_WORKER_DEPLOY=true`.

## Structure

```text
astro.config.mjs       static output and canonical site URL
wrangler.jsonc         Worker asset binding and custom-domain routes
src/worker/index.js    small Worker entrypoint serving the asset binding
src/pages/             landing page and 404 page
src/components/        hero, dashboard preview, features and deployment sections
src/styles/global.css  dark glass visual system shared with the app
public/                local logos, favicon, social card and crawler metadata
```
