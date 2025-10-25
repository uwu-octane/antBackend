type Msg = { role: string; content: string }

export default defineEventHandler(async (event) => {
  setResponseHeader(event, 'Content-Type', 'text/event-stream; charset=utf-8')
  setResponseHeader(event, 'Cache-Control', 'no-cache, no-transform')
  setResponseHeader(event, 'Connection', 'keep-alive')

  const body = await readBody<{ messages?: Msg[] }>(event)
  const last = body?.messages?.at(-1)?.content ?? ''
  const te = new TextEncoder()

  const stream = new ReadableStream({
    start(controller) {
      const chunks = [`你说的是：${last}`, ' 这是', ' 流式', ' mock', ' ✅']
      let i = 0
      const timer = setInterval(() => {
        if (i >= chunks.length) {
          controller.enqueue(te.encode('event: done\ndata: {}\n\n'))
          clearInterval(timer)
          controller.close()
          return
        }
        controller.enqueue(te.encode(`data: ${JSON.stringify({ delta: chunks[i++] })}\n\n`))
      }, 150)
    },
  })
  return sendStream(event, stream)
})