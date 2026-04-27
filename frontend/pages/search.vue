<script setup lang="ts">
const { me } = useAuth()
const route = useRoute()
const router = useRouter()

const query = ref("")
const results = ref<any[]>([])
const total = ref(0)
const isLoading = ref(false)
const errorMessage = ref("")

const num = 10
const offset = computed(() => Number(route.query.offset) || 0)
const searchQuery = computed(() => (route.query.q as string) || "")

onMounted(async () => {
  try {
    await me()
  } catch {
    navigateTo("/login")
  }
  if (searchQuery.value) {
    query.value = searchQuery.value
    await performSearch()
  }
})

watch(() => route.query.q, (newVal) => {
  if (newVal) {
    query.value = newVal as string
    performSearch()
  }
})

watch(() => route.query.offset, () => {
  performSearch()
})

async function performSearch() {
  if (!query.value.trim()) return
  isLoading.value = true
  errorMessage.value = ""
  try {
    // Используем $fetch к внутреннему API
    const data = await $fetch("/api/v1/search", {
      params: {
        q: query.value,
        num,
        offset: offset.value,
      },
    })
    results.value = data.results || []
    total.value = data.total || 0
  } catch (e: any) {
    errorMessage.value = "Search failed. Please try again."
    results.value = []
  } finally {
    isLoading.value = false
  }
}

function newSearch() {
  router.push({ path: "/search", query: { q: query.value, num, offset: 0 } })
}

function prevPage() {
  if (offset.value >= num) {
    router.push({ path: "/search", query: { q: query.value, num, offset: offset.value - num } })
  }
}

function nextPage() {
  if (offset.value + num < total.value) {
    router.push({ path: "/search", query: { q: query.value, num, offset: offset.value + num } })
  }
}
</script>

<template>
  <div class="search-page">
    <div class="search-header">
      <h1 class="site-title">
        <NuxtLink to="/">Search engine</NuxtLink>
      </h1>
      <div class="search-bar">
        <input
          v-model="query"
          placeholder="Search..."
          @keyup.enter="newSearch"
        />
        <button @click="newSearch">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
        </button>
      </div>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="isLoading" class="loading">Loading results...</div>

    <div v-else-if="results.length" class="results">
      <p class="count">About {{ total }} results (showing {{ offset + 1 }}–{{ Math.min(offset + num, total) }})</p>
      <ul>
        <li v-for="item in results" :key="item.url" class="result-item">
          <a :href="item.url" target="_blank" class="title">{{ item.title }}</a>
          <cite class="url">{{ item.url }}</cite>
          <p class="snippet">{{ item.snippet }}</p>
        </li>
      </ul>
      <div class="pagination">
        <button :disabled="offset === 0" @click="prevPage">← Previous</button>
        <button :disabled="offset + num >= total" @click="nextPage">Next →</button>
      </div>
    </div>

    <p v-else-if="!isLoading && searchQuery" class="no-results">No results found for "{{ searchQuery }}"</p>
  </div>
</template>

<style scoped>
.search-page {
  max-width: 720px;
  margin: 0 auto;
  padding: 1rem;
}

.search-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  border-bottom: 1px solid #eee;
  padding-bottom: 1rem;
}

.site-title {
  font-size: 1.5rem;
  margin: 0;
  white-space: nowrap;
}

.site-title a {
  color: inherit;
  text-decoration: none;
}

.search-bar {
  display: flex;
  flex: 1;
  border: 1px solid #ccc;
  border-radius: 24px;
  overflow: hidden;
}

.search-bar input {
  flex: 1;
  border: none;
  outline: none;
  padding: 10px 16px;
  font-size: 16px;
  background: transparent;
  color: inherit;
}

.search-bar button {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px 16px;
  border: none;
  background: transparent;
  cursor: pointer;
  color: inherit;
}

.error {
  color: #c33;
  background: #fee;
  padding: 8px;
  border-radius: 4px;
}

.loading {
  text-align: center;
  padding: 2rem;
}

.count {
  color: #666;
  font-size: 14px;
  margin-bottom: 1rem;
}

.results ul {
  list-style: none;
  padding: 0;
}

.result-item {
  margin-bottom: 1.5rem;
}

.title {
  font-size: 18px;
  color: #1a0dab;
  text-decoration: none;
  font-weight: 500;
}

.title:hover {
  text-decoration: underline;
}

.url {
  display: block;
  font-style: normal;
  color: #006621;
  font-size: 14px;
  margin-bottom: 4px;
}

.snippet {
  color: #545454;
  margin: 0;
  font-size: 14px;
  line-height: 1.4;
}

.pagination {
  display: flex;
  justify-content: space-between;
  margin-top: 2rem;
}

.pagination button {
  padding: 8px 16px;
  cursor: pointer;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.no-results {
  text-align: center;
  padding: 2rem;
  color: #666;
}
</style>