export default defineNuxtRouteMiddleware((to) => {
  // Skip middleware on server side
  if (process.server) return

  const api = useApi()

  // Load token from localStorage if not already loaded
  api.loadToken()

  // If not authenticated and trying to access a protected route, stay on index (login page)
  // The index page handles showing login vs dashboard based on auth state
  if (!api.isAuthenticated.value && to.path !== '/') {
    return navigateTo('/')
  }
})
