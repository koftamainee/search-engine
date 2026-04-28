<script setup lang="ts">
const { me } = useAuth()
const route = useRoute()
const router = useRouter()

const query = ref("")
const results = ref<any[]>([])
const total = ref(0)
const isLoading = ref(false)
const errorMessage = ref("")
const suggestions = ref<{ type: string; data: string }[]>([])
const showSuggestions = ref(false)
const selectedIndex = ref(-1)
const isFocused = ref(false)

const num = 10
const offset = computed(() => Number(route.query.offset) || 0)
const searchQuery = computed(() => (route.query.q as string) || "")

let debounceTimer: ReturnType<typeof setTimeout> | null = null

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

watch(query, (val) => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    fetchSuggestions(val)
  }, 150)
})

async function fetchSuggestions(q: string) {
  try {
    const res = await $fetch<{
      status: string
      data: { query: string; suggestions: { type: string; data: string }[] }
    }>(`/api/v1/suggest?q=${encodeURIComponent(q)}&num=10`)

    suggestions.value = res.data.suggestions ?? []

    if (isFocused.value) {
      showSuggestions.value = suggestions.value.length > 0
    }
    selectedIndex.value = -1
  } catch {
    suggestions.value = []
    showSuggestions.value = false
  }
}

async function performSearch() {
  if (!query.value.trim()) return
  isLoading.value = true
  errorMessage.value = ""
  showSuggestions.value = false
  try {
    const data = await $fetch<{
      query: string
      hits: { id: string; score: number; data: any }[]
      num: number
      total: number
      offset: number
    }>("/api/v1/search", {
      params: {
        q: query.value,
        num,
        offset: offset.value,
      },
    })
    results.value = data.hits || []
    total.value = data.total || 0
  } catch (e: any) {
    errorMessage.value = "Search failed. Please try again."
    results.value = []
  } finally {
    isLoading.value = false
  }
}

function selectSuggestion(suggestion: { type: string; data: string }) {
  query.value = suggestion.data
  newSearch()
}

function newSearch() {
  if (!query.value.trim()) return
  showSuggestions.value = false
  router.push({ path: "/search", query: { q: query.value, num, offset: 0 } })
}

function onKeyDown(e: KeyboardEvent) {
  if (!showSuggestions.value) return

  if (e.key === "ArrowDown") {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, suggestions.value.length - 1)
  } else if (e.key === "ArrowUp") {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, -1)
  } else if (e.key === "Enter" && selectedIndex.value >= 0) {
    e.preventDefault()
    selectSuggestion(suggestions.value[selectedIndex.value]!)
  }
}

function onFocus() {
  isFocused.value = true
  fetchSuggestions(query.value)
}

function onBlur() {
  isFocused.value = false
  setTimeout(() => {
    showSuggestions.value = false
  }, 200)
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
  <div class="wrapper">
    <NuxtLink to="/me" class="profile-icon" aria-label="Profile">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
    </NuxtLink>

    <div class="search-page">
      <div class="search-header">
        <h1 class="site-title">
          <NuxtLink to="/">Search engine</NuxtLink>
        </h1>
        <div class="search-container">
          <div class="search-bar">
            <input
              v-model="query"
              placeholder="Search..."
              @keyup.enter="newSearch"
              @keydown="onKeyDown"
              @focus="onFocus"
              @blur="onBlur"
            />
            <button @click="newSearch">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="11" cy="11" r="8" />
                <line x1="21" y1="21" x2="16.65" y2="16.65" />
              </svg>
            </button>
          </div>
          
          <div v-if="showSuggestions" class="suggestions-dropdown">
            <div
              v-for="(suggestion, index) in suggestions"
              :key="index"
              class="suggestion-item"
              :class="{ selected: index === selectedIndex }"
              @mousedown.prevent="selectSuggestion(suggestion)"
            >
              <svg v-if="suggestion.type === 'history'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="suggestion-icon">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
              <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="suggestion-icon">
                <circle cx="11" cy="11" r="8" />
                <line x1="21" y1="21" x2="16.65" y2="16.65" />
              </svg>
              <span>{{ suggestion.data }}</span>
            </div>
          </div>
        </div>
      </div>

      <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

      <div v-if="isLoading" class="loading">Loading results...</div>

      <div v-else-if="results.length" class="results">
        <p class="count">About {{ total }} results (showing {{ offset + 1 }}–{{ Math.min(offset + num, total) }})</p>
        <ul>
          <li v-for="item in results" :key="item.id" class="result-item">
            <a :href="item.data.url" target="_blank" class="title">{{ item.data.title }}</a>
            <cite class="url">{{ item.data.url }}</cite>
            <p class="snippet">{{ item.data.snippet }}</p>
          </li>
        </ul>
        <div class="pagination">
          <button :disabled="offset === 0" @click="prevPage">← Previous</button>
          <button :disabled="offset + num >= total" @click="nextPage">Next →</button>
        </div>
      </div>

      <p v-else-if="!isLoading && searchQuery" class="no-results">No results found for "{{ searchQuery }}"</p>
    </div>
  </div>
</template>

<style scoped>
.wrapper {
  position: relative;
  min-height: 100vh;
  padding-top: 80px;
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
  z-index: 10;
}

.profile-icon:hover {
  background-color: rgba(128, 128, 128, 0.2);
}

.search-page {
  max-width: 760px;
  margin: 0 auto;
  padding: 0 1.5rem 2rem;
}

.search-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(128, 128, 128, 0.2);
}

.site-title {
  font-size: 1.5rem;
  margin: 0;
  white-space: nowrap;
  flex-shrink: 0;
}

.site-title a {
  color: inherit;
  text-decoration: none;
}

.site-title a:hover {
  color: #646cff;
}

.search-container {
  position: relative;
  flex: 1;
}

.search-bar {
  display: flex;
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 24px;
  overflow: hidden;
  transition: border-color 0.2s;
}

.search-bar:focus-within {
  border-color: #646cff;
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
  transition: color 0.2s;
}

.search-bar button:hover {
  color: #646cff;
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 8px;
  background: var(--bg, #fff);
  border: 1px solid rgba(128, 128, 128, 0.3);
  border-radius: 12px;
  overflow: hidden;
  z-index: 10;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  cursor: pointer;
  transition: background-color 0.15s;
}

.suggestion-item:hover,
.suggestion-item.selected {
  background-color: rgba(128, 128, 128, 0.1);
}

.suggestion-icon {
  flex-shrink: 0;
  opacity: 0.6;
}

.error {
  color: #c33;
  background: #fee;
  padding: 10px 14px;
  border-radius: 6px;
  margin-bottom: 1.5rem;
}

.loading {
  text-align: center;
  padding: 3rem 0;
  color: #888;
}

.count {
  color: #888;
  font-size: 14px;
  margin-bottom: 1.5rem;
}

.results ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.result-item {
  margin-bottom: 1.8rem;
}

.title {
  font-size: 18px;
  color: #646cff;
  text-decoration: none;
  font-weight: 500;
  display: inline-block;
  margin-bottom: 2px;
}

.title:hover {
  text-decoration: underline;
}

.url {
  display: block;
  font-style: normal;
  color: #4a9e5c;
  font-size: 13px;
  margin-bottom: 6px;
  word-break: break-all;
}

.snippet {
  color: #999;
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
}

.pagination {
  display: flex;
  justify-content: center;
  gap: 1rem;
  margin-top: 2.5rem;
}

.pagination button {
  padding: 10px 20px;
  cursor: pointer;
  border-radius: 8px;
  font-size: 14px;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.no-results {
  text-align: center;
  padding: 3rem 0;
  color: #888;
  font-size: 15px;
}
</style>