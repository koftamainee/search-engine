<script setup lang="ts">
const { me, logout } = useAuth()

const user = ref<any>(null)
const errorMessage = ref("")
const isLoading = ref(false)

onMounted(async () => {
  try {
    const res = await me()
    user.value = res.data
  } catch (error: any) {
    if (error.response?.status === 403) {
      errorMessage.value = "Your account has been blocked. Please contact support."
    } else {
      navigateTo("/login")
    }
  }
})

async function onLogout() {
  isLoading.value = true
  try {
    await logout()
    navigateTo("/login")
  } catch (error: any) {
    errorMessage.value = "Logout error. Please try again."
  } finally {
    isLoading.value = false
  }
}

function clearError() {
  errorMessage.value = ""
}
</script>

<template>
  <div class="wrapper">
    <NuxtLink to="/" class="home-icon" aria-label="Home">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
      </svg>
    </NuxtLink>

    <div class="page">
      <h1>Profile</h1>

      <div v-if="errorMessage" class="error-message">
        {{ errorMessage }}
        <button @click="clearError" class="close-btn">&times;</button>
      </div>

      <div v-if="user">
        <div class="user-info">
          <h3>User Information</h3>
          <ul>
            <li><strong>ID:</strong> {{ user.id }}</li>
            <li><strong>Email:</strong> {{ user.email }}</li>
            <li><strong>Admin:</strong> {{ user.is_admin }}</li>
            <li><strong>Banned:</strong> {{ user.is_banned }}</li>
            <li><strong>Created at:</strong> {{ user.created_at }}</li>
          </ul>
        </div>
        <button @click="onLogout" :disabled="isLoading" class="logout-btn">
          <svg v-if="isLoading" class="spinner" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" opacity="0.25" />
            <path d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" fill="currentColor" opacity="0.75" />
          </svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
            <polyline points="16 17 21 12 16 7" />
            <line x1="21" y1="12" x2="9" y2="12" />
          </svg>
          {{ isLoading ? "Logging out..." : "Logout" }}
        </button>
      </div>

      <p v-else-if="!errorMessage">Loading...</p>
    </div>
  </div>
</template>

<style scoped>
.wrapper {
  position: relative;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}

.home-icon {
  position: absolute;
  top: 1.5rem;
  left: 1.5rem;
  color: inherit;
  text-decoration: none;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  transition: background-color 0.2s;
}

.home-icon:hover {
  background-color: rgba(128, 128, 128, 0.2);
}

.page {
  text-align: center;
  max-width: 500px;
  margin: 0 auto;
}

.error-message {
  background-color: #fee;
  color: #c33;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
  border: 1px solid #fcc;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #c33;
  padding: 0 5px;
  margin: 0;
}

.close-btn:hover {
  color: #a00;
}

.user-info {
  margin-bottom: 20px;
  padding: 15px;
  border: 1px solid #ddd;
  border-radius: 4px;
  text-align: left;
}

ul {
  list-style: none;
  padding: 0;
}

li {
  margin: 8px 0;
}

.logout-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
}

.logout-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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