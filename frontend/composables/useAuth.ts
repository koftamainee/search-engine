export const useAuth = () => {
    const { $api } = useNuxtApp()

    const login = (email: string, password: string) => {
        return $api('/auth/login', {
            method: 'POST',
            body: { email, password }
        })
    }

    const register = (email: string, password: string) => {
        return $api('/auth/register', {
            method: 'POST',
            body: { email, password }
        })
    }

    const me = () => $api('/me')

    const logout = () => $api('/auth/logout', {
        method: 'POST'
    })

    return { login, register, me, logout }
}