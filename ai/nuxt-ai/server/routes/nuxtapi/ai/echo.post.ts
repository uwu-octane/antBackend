export default defineEventHandler(async (event) => {
    const raw = await readRawBody(event).catch(() => null)  // Buffer|string|null
    let parsed: any = null
    try {
      parsed = raw ? JSON.parse(typeof raw === 'string' ? raw : (raw as Buffer).toString('utf-8')) : null
    } catch (e) {
      parsed = { __parseError: String(e) }
    }
    return {
      ok: true,
      path: event.path,
      method: event.method,
      headers: getRequestHeaders(event),
      rawLen: raw ? (typeof raw === 'string' ? raw.length : (raw as Buffer).byteLength) : 0,
      parsed,
    }
  })