import Consul from 'consul'

export default defineNitroPlugin(async () => {
    const host = process.env.HOST || 'localhost'
    const port = Number(process.env.PORT) || 7780
    const consulHost = process.env.CONSUL_HOST || 'localhost'
    const consulPort = Number(process.env.CONSUL_PORT) || 8500
    const consul = new Consul({
        host: consulHost,
        port: consulPort,
    })
    const id = `nuxt-ai-${host}-${port}`
    const svc = {
        id,
        name: 'nuxt-ai',
        address: 'host.docker.internal',
        port: port,
        tags:['gateway', 'ai', 'http'],
        check: {
            http: `http://host.docker.internal:${port}/api/health`,
            interval: '10s',
            timeout: '5s',
            deregistercriticalserviceafter: '1m',
        },
    } as const 
    console.log('🔌 Nitro plugin bootstrap: server/plugins loaded')
    try {
        await consul.agent.service.register(svc as any)
        console.log(` Registered nuxt-ai to Consul at ${consulHost}:${consulPort} as ${id}`)
    } catch (error) {
        console.error(`Failed to register nuxt-ai to Consul:`, error)
    }

    const cleanup = async () => {
        try {
            await consul.agent.service.deregister(id)
            console.log(` Deregistered nuxt-ai from Consul at ${consulHost}:${consulPort} as ${id}`)
        } catch (error) {
            
        } finally {
            process.exit(0)
        }
    }

    process.on('SIGINT', cleanup)
    process.on('SIGTERM', cleanup)
})