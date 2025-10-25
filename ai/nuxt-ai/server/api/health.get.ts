export default defineEventHandler(async (event) => {
    return {
        ok: true,
        service: useRuntimeConfig().public.serviceName || 'nuxt-ai',
        ts:Date.now(),
        auth: getHeader(event, 'Authorization') ? 'present' : 'absent',
    }
})