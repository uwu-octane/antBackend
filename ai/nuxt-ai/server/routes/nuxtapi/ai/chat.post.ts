
type Msg = { role: string; content: string }

export default defineEventHandler(async (event) => {
  try {
    // 用 raw 规避解析器差异
    const raw = await readRawBody(event)          // Buffer | string | null
    let body: { messages?: Msg[] } | null = null

    if (raw && raw.length > 0) {
      // 统一按 utf-8 转字符串再解析
      const s = typeof raw === 'string' ? raw : (raw as Buffer).toString('utf-8')
      body = JSON.parse(s)
    } else {
      body = null
    }

    const last = body?.messages?.at(-1)?.content ?? ''
    return {
      id: crypto.randomUUID(),
      model: 'mock-llm',
      created: Date.now(),
      choices: [
        { index: 0, message: { role: 'assistant', content: `你说的是：${last}（mock）` } },
      ],
    }
  } catch (e: any) {
    setResponseStatus(event, 400)
    return { message: e?.message ?? 'bad request' }
  }
})