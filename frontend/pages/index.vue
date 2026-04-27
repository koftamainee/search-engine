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
    <!-- Here will be search history later -->
  </div>
</template>

<style scoped>
.page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 2rem;
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
  width: 100%;
  max-width: 560px;
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