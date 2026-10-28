# Proxyglass website

Astro static marketing site for [proxyglass.app](https://proxyglass.app), styled to match the
Proxyglass dashboard. It has no Easter eggs or external runtime asset dependencies.

## Local development

Requires Node.js 22.23.4+.

```sh
npm ci
npm run dev -- --background
```

The Astro dev server opens at <http://localhost:4321>. Manage it with `npx astro dev status`,
`npx astro dev logs` and `npx astro dev stop`.

## Build and check

```sh
npm run build
npm run smoke
npx wrangler deploy --dry-run
```

Astro writes a static site to `dist/`. Wrangler serves those generated assets from a Cloudflare
Worker; the Worker has no application data or server-side rendering.

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

`site/wrangler.jsonc` attaches a Cloudflare Worker Custom Domain at `proxyglass.app`, deploys
`dist/` as static assets, and returns the generated `404.html` for missing paths. Wrangler/Cloudflare
provisions the apex DNS record and TLS certificate for the custom domain.

For **Workers Builds** in the Cloudflare dashboard, use Worker name `proxyglass`, repository
`chrisvanmeer/proxyglass`, production branch `main`, root directory `site`, build command
`npm run build`, and deploy command `npx wrangler deploy`. Choose **None/Other** for the framework
preset (this is a Worker with static assets, not a Pages project); the Wrangler config supplies the
asset directory and custom domain.

## Cloudflare Email Routing

In the Cloudflare dashboard, open **Email → Email Routing**, add and verify a destination inbox,
then create the addresses you want to forward (for example, `hello@proxyglass.app`). Email Routing
shows and installs the exact MX and SPF records for your zone. Keep those mail records alongside the
Worker-managed website record; they serve different protocols. Email Routing handles incoming
forwarding only—it does not provide SMTP for sending mail.

The GitHub Actions workflow `.github/workflows/site-deploy.yml` supports manual deploys. To deploy
automatically on pushes to `main`, configure the two secrets above and set the repository variable
`PROXYGLASS_WORKER_DEPLOY=true`.

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
