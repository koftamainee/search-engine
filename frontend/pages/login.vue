<script setup lang="ts">
const { login, me } = useAuth()

const email = ref("")
const password = ref("")
const errorMessage = ref("")
const isLoading = ref(false)
const isCheckingAuth = ref(true)

onMounted(async () => {
  try {
    await me()
    navigateTo("/")
  } catch {
    isCheckingAuth.value = false
  }
})

async function onLogin() {
  errorMessage.value = ""
  isLoading.value = true

  try {
    await login(email.value, password.value)
    navigateTo("/")
  } catch (error: any) {
    if (error.response?.status === 403) {
      errorMessage.value = "Your account has been blocked. Please contact support."
    } else if (error.response?.status === 401) {
      errorMessage.value = "Invalid email or password."
    } else {
      errorMessage.value = "Login error. Please try again later."
    }
  } finally {
    isLoading.value = false
  }
}

function clearError() {
  errorMessage.value = ""
}
</script>

<template>
  <div v-if="isCheckingAuth" class="page">
    <p>Checking authentication...</p>
  </div>

  <div v-else class="page">
    <h1>Login</h1>

    <div v-if="errorMessage" class="error-message">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
      </svg>
      {{ errorMessage }}
      <button @click="clearError" class="close-btn">&times;</button>
    </div>

    <div class="input-group">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
        <polyline points="22,6 12,13 2,6"/>
      </svg>
      <input
        v-model="email"
        placeholder="Email"
        type="email"
        :disabled="isLoading"
      />
    </div>

    <div class="input-group">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
        <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
      </svg>
      <input
        v-model="password"
        type="password"
        placeholder="Password"
        :disabled="isLoading"
      />
    </div>

    <button @click="onLogin" :disabled="isLoading" class="login-btn">
      <svg v-if="isLoading" class="spinner" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" opacity="0.25" />
        <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" fill="currentColor" opacity="0.75" />
      </svg>
      <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4"/>
        <polyline points="10 17 15 12 10 7"/>
        <line x1="15" y1="12" x2="3" y2="12"/>
      </svg>
      {{ isLoading ? "Logging in..." : "Login" }}
    </button>

    <NuxtLink to="/register" class="register-btn" :class="{ disabled: isLoading }" @click.prevent="isLoading ? null : navigateTo('/register')">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
        <circle cx="8.5" cy="7" r="4"/>
        <line x1="20" y1="8" x2="20" y2="14"/>
        <line x1="23" y1="11" x2="17" y2="11"/>
      </svg>
      Register
    </NuxtLink>
  </div>
</template>

<style scoped>
.page {
  text-align: center;
  max-width: 360px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 20px;
}

.error-message {
  background-color: #fee;
  color: #c33;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  border: 1px solid #fcc;
  display: flex;
  align-items: center;
  gap: 8px;
  text-align: left;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #c33;
  padding: 0 5px;
  margin: 0;
  margin-left: auto;
}

.close-btn:hover {
  color: #a00;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  padding: 0 10px;
  margin-bottom: 10px;
}

.input-group svg {
  flex-shrink: 0;
  color: #888;
}

.input-group input {
  border: none;
  outline: none;
  padding: 10px 0;
  width: 100%;
  background: transparent;
  color: inherit;
  font-size: 14px;
}

.login-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  cursor: pointer;
  width: 100%;
  justify-content: center;
  margin-top: 5px;
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.register-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 24px;
  cursor: pointer;
  width: 100%;
  justify-content: center;
  text-decoration: none;
  color: inherit;
  background: transparent;
  border: 1px solid #ccc;
  border-radius: 8px;
  margin-top: 8px;
  font-size: 14px;
}

.register-btn:hover {
  border-color: #646cff;
  color: #646cff;
}

.register-btn.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>