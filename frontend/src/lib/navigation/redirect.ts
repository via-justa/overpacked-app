/**
 * Returns the given path only if it is a safe, same-origin internal path;
 * otherwise returns null.
 *
 * Guards the post-login redirect against open-redirect attacks: a value like
 * `//evil.com` or `/\evil.com` passes a naive `startsWith('/')` check but is
 * treated by browsers as a protocol-relative absolute URL, bouncing the user
 * off-origin after sign-in.
 */
export function safeRedirectPath(path: unknown): string | null {
  if (typeof path !== 'string') {
    return null
  }
  // Must be root-relative, but not protocol-relative ("//host" or "/\host").
  if (!path.startsWith('/') || path.startsWith('//') || path.startsWith('/\\')) {
    return null
  }
  return path
}
