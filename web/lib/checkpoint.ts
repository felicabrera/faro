/**
 * Parsing for the signed-note checkpoints FARO publishes.
 *
 * A checkpoint is the log's claim about its own current state: an origin, a tree
 * size and a root hash, followed by one or more signatures. The format is the
 * signed note (golang.org/x/mod/sumdb/note) used by tlog-tiles logs generally,
 * which is why a verifier written against the C2SP specs can read FARO.
 *
 * What this module does NOT do, and must not be mistaken for: it does not verify
 * the signature, and it does not verify any inclusion or consistency proof. It
 * splits bytes into fields. Everything here is unverified input until a verifier
 * says otherwise.
 *
 * The verification layer is the next piece of work on this explorer, and it has
 * to run in the browser. A page that renders "verified" because the server said
 * so is worse than a page that renders nothing, because it manufactures
 * confidence that nobody checked.
 */

/** A parsed but *unverified* checkpoint. */
export interface Checkpoint {
  /** The log's identity, which a verifier must pin. */
  origin: string
  /** Number of entries in the tree. */
  size: bigint
  /** Base64 Merkle root hash, exactly as it appeared in the note. */
  rootHash: string
  /** The signature lines, unparsed and unchecked. */
  signatures: string[]
  /** The raw note, kept because it, not the parsed struct, is the evidence. */
  raw: string
}

export class CheckpointParseError extends Error {
  constructor(message: string) {
    super(`checkpoint: ${message}`)
    this.name = 'CheckpointParseError'
  }
}

/**
 * Parses a signed note into its checkpoint fields.
 *
 * The note body is the first three lines (origin, size, root hash), then a blank
 * line, then signature lines beginning with an em-dash marker.
 */
export function parseCheckpoint(raw: string): Checkpoint {
  const [body = '', sigBlock = ''] = splitNote(raw)
  const lines = body.split('\n')
  const [origin, sizeLine, rootHash] = lines

  if (!origin || !sizeLine || !rootHash) {
    throw new CheckpointParseError('expected an origin, a size and a root hash')
  }
  if (!/^\d+$/.test(sizeLine)) {
    throw new CheckpointParseError(`tree size ${JSON.stringify(sizeLine)} is not a number`)
  }

  const signatures = sigBlock
    .split('\n')
    .map((l) => l.trim())
    .filter((l) => l.length > 0)

  if (signatures.length === 0) {
    throw new CheckpointParseError('no signatures; an unsigned checkpoint proves nothing')
  }

  return { origin, size: BigInt(sizeLine), rootHash, signatures, raw }
}

/** Splits a note into its body and its signature block. */
function splitNote(raw: string): [string, string] {
  const marker = '\n\n'
  const at = raw.indexOf(marker)
  if (at < 0) {
    throw new CheckpointParseError('malformed note: no blank line separating body from signatures')
  }
  return [raw.slice(0, at), raw.slice(at + marker.length)]
}

/** The base URL of the log this build points at. */
export const faroURL: string = process.env.NEXT_PUBLIC_FARO_URL ?? 'http://localhost:2025'

/**
 * Fetches the current checkpoint from the log.
 *
 * `cache: 'no-store'` matches the Cache-Control the log sends: a stale
 * checkpoint would show a tree smaller than the real one, which looks exactly
 * like a log that stopped accepting entries.
 */
export async function fetchCheckpoint(baseURL: string = faroURL): Promise<Checkpoint> {
  const response = await fetch(new URL('checkpoint', ensureTrailingSlash(baseURL)), {
    cache: 'no-store',
  })
  if (!response.ok) {
    throw new Error(`checkpoint: log returned ${response.status} ${response.statusText}`)
  }
  return parseCheckpoint(await response.text())
}

function ensureTrailingSlash(url: string): string {
  return url.endsWith('/') ? url : `${url}/`
}
