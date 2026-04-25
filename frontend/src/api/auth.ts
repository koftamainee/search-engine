import { apiFetch } from "./client";

export function login(email: string, password: string) {
    return apiFetch("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    });
}

export function register(email: string, password: string) {
    return apiFetch("/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password }),
    });
}

export function me() {
    return apiFetch("/me");
}

export function logout() {
    return apiFetch("/auth/logout", {
        method: "POST",
    });
}