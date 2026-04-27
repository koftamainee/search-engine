<script setup lang="ts">
const { me } = useAuth()
const router = useRouter()
const query = ref("")

onMounted(async () => {
  try {
    await me()
  } catch {
    navigateTo("/login")
  }
})

function onSearch() {
  if (!query.value.trim()) return
  router.push({ path: "/search", query: { q: query.value, num: 10, offset: 0 } })
}
</script>

<template>
  <div class="wrapper">
    <NuxtLink to="/me" class="profile-icon" aria-label="Profile">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
    </NuxtLink>

    <div class="page">
      <h1 class="title">Search engine</h1>
      <div class="search-bar">
        <input
          v-model="query"
          placeholder="Type your search..."
          @keyup.enter="onSearch"
        />
        <button @click="onSearch">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </button>
      </div>
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

.profile-icon {
  position: absolute;
  top: 1.5rem;
  right: 1.5rem;
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

.profile-icon:hover {
  background-color: rgba(128, 128, 128, 0.2);
}

.page {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  margin-bottom: 20vh;
}

.title {
  font-size: 2.5rem;
  margin: 0;
}

.search-bar {
  display: flex;
  align-items: center;
  border: 1px solid #ccc;
  border-radius: 24px;
  overflow: hidden;
  width: 560px;
  max-width: 90vw;
  background: transparent;
}

.search-bar input {
  flex: 1;
  border: none;
  outline: none;
  padding: 12px 16px;
  font-size: 16px;
  background: transparent;
  color: inherit;
}

.search-bar button {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 16px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: inherit;
}

.search-bar button:hover {
  color: #646cff;
}
</style>