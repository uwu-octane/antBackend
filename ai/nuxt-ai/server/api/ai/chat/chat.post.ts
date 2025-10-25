type Msg = { role: string; content: string }

export default defineEventHandler(async (event) => {
  try {
    const body = await readBody<{ messages?: Msg[] }>(event)
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