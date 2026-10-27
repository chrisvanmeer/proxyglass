import assert from 'node:assert/strict';
import { readFile, access } from 'node:fs/promises';
import { URL } from 'node:url';

const html = await readFile(new URL('../dist/index.html', import.meta.url), 'utf8');
for (const text of [
	'proxyglass',
	'See every proxy.',
	'Change nothing.',
	'Read-only by design',
	'Native binary + systemd',
	'Container + Nomad',
	'Last check',
	'GitHub',
]) {
	assert.ok(html.toLowerCase().includes(text.toLowerCase()), `landing page should include: ${text}`);
}

for (const asset of ['favicon.svg', 'og.svg', 'robots.txt', 'sitemap.xml', '404.html', 'assets/proxyglass-mark.svg']) {
	await access(new URL(`../dist/${asset}`, import.meta.url));
}

const oidcDocs = await readFile(new URL('../dist/docs/oidc/index.html', import.meta.url), 'utf8');
assert.ok(oidcDocs.includes('client_secret'), 'OIDC docs should include the config sample');
assert.ok(oidcDocs.includes('/ui/oidc/callback'), 'OIDC docs should explain the callback URL');
for (const route of ['docs/index.html', 'docs/configuration/index.html', 'docs/deployment/index.html']) {
	await access(new URL(`../dist/${route}`, import.meta.url));
}

const localPages = [
	['/', html],
	['/docs/', await readFile(new URL('../dist/docs/index.html', import.meta.url), 'utf8')],
	['/docs/configuration/', await readFile(new URL('../dist/docs/configuration/index.html', import.meta.url), 'utf8')],
	['/docs/oidc/', oidcDocs],
	['/docs/deployment/', await readFile(new URL('../dist/docs/deployment/index.html', import.meta.url), 'utf8')],
];
for (const [path, page] of localPages) {
	for (const [, href] of page.matchAll(/\b(?:href|src)="([^"]+)"/g)) {
		if (href.startsWith('https://github.com/chrisvanmeer/proxyglass')) continue;
		assert.ok(!/^(?:[a-z]+:|\/)/i.test(href), `internal href/src should be relative for local testing: ${href}`);
		assert.equal(new URL(href, `https://local.test${path}`).hostname, 'local.test');
	}
}

assert.ok(!/unpkg\.com|fonts\.googleapis\.com|fonts\.gstatic\.com/i.test(html), 'page should not depend on third-party browser assets');
console.log('Proxyglass site smoke test passed.');
