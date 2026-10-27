/** Return a same-site relative URL from the current route to a site path. */
export function relativeLink(fromPath: string, target: string): string {
  const suffixAt = target.search(/[?#]/);
  const targetPath = suffixAt < 0 ? target : target.slice(0, suffixAt);
  const suffix = suffixAt < 0 ? '' : target.slice(suffixAt);
  const fromDirectory = fromPath.endsWith('/') ? fromPath : fromPath.slice(0, fromPath.lastIndexOf('/') + 1);
  const fromParts = fromDirectory.split('/').filter(Boolean);
  const targetParts = targetPath.split('/').filter(Boolean);

  let shared = 0;
  while (shared < fromParts.length && shared < targetParts.length && fromParts[shared] === targetParts[shared]) {
    shared++;
  }

  const relative = [
    ...Array.from({ length: fromParts.length - shared }, () => '..'),
    ...targetParts.slice(shared),
  ].join('/');

  return `${relative || '.'}${suffix}`;
}
