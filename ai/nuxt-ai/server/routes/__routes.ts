export default defineEventHandler(() => {
    const app = useNitroApp()
    const router = app.router as any  
    return (router.routes || []).map((r: any) => ({
      path: r.path,
      methods: Object.keys(r.handlers || {}),
      file: r.file,
    }))
  })