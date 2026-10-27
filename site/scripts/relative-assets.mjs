import { readdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, relative, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const dist = fileURLToPath(new URL('../dist/', import.meta.url));
const toPosix = (value) => value.split(sep).join('/');

async function walk(directory) {
	const entries = await readdir(directory, { withFileTypes: true });
	const files = [];
	for (const entry of entries) {
		const path = `${directory}/${entry.name}`;
		if (entry.isDirectory()) files.push(...await walk(path));
		else if (entry.isFile() && entry.name.endsWith('.html')) files.push(path);
	}
	return files;
}

for (const file of await walk(dist)) {
	const pagePath = toPosix(relative(dist, file));
	const pageDirectory = dirname(pagePath);
	let html = await readFile(file, 'utf8');
	html = html.replace(/\b(href|src)=(['"])\/([^'"]*)\2/g, (_match, attribute, quote, rawTarget) => {
		const suffixAt = rawTarget.search(/[?#]/);
		const target = suffixAt < 0 ? rawTarget : rawTarget.slice(0, suffixAt);
		const suffix = suffixAt < 0 ? '' : rawTarget.slice(suffixAt);
		const fromDirectory = pageDirectory === '.' ? '' : pageDirectory;
		let targetRelative = relative(fromDirectory, target).split(sep).join('/');
		if (!targetRelative.startsWith('.')) targetRelative = `./${targetRelative}`;
		return `${attribute}=${quote}${targetRelative}${suffix}${quote}`;
	});
	await writeFile(file, html);
}
