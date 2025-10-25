// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },
  devServer: {
    host: process.env.HOST || 'localhost',
    port: Number(process.env.PORT) || 7780,
  },
  runtimeConfig: {
    ai: {apiKey : ''},
    public: {serviceName: 'nuxt-ai'},
  },
  vite: {
    server: {
      allowedHosts: ['host.docker.internal', 'localhost', '127.0.0.1'],
    },
  },
})
