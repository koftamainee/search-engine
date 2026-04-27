export default defineNuxtRouteMiddleware((to) => {
    const token = useCookie<string | null>('access_token')

    if (!token.value) {
        if (to.path !== '/login' && to.path !== '/register') {
            return navigateTo('/login')
        }
    }
})