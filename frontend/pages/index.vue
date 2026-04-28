<script setup lang="ts">
const { me } = useAuth()
const router = useRouter()
const query = ref("")
const suggestions = ref<{ type: string; data: string }[]>([])
const showSuggestions = ref(false)
const selectedIndex = ref(-1)
const isFocused = ref(false)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  try {
    await me()
  } catch (err) {
    navigateTo("/login")
  }
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
  } catch (err) {
    suggestions.value = []
    showSuggestions.value = false
  }
}

watch(query, (val) => {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
  }
  debounceTimer = setTimeout(() => {
    fetchSuggestions(val)
  }, 150)
})

function onSearch() {
  if (!query.value.trim()) return
  showSuggestions.value = false
  router.push({ path: "/search", query: { q: query.value, num: 10, offset: 0 } })
}

function selectSuggestion(suggestion: { type: string; data: string }) {
  query.value = suggestion.data
  showSuggestions.value = false
  onSearch()
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
      <div class="search-bar-wrapper">
        <div class="search-bar">
          <input
            v-model="query"
            placeholder="Type your search..."
            @keyup.enter="onSearch"
            @keydown="onKeyDown"
            @focus="onFocus"
            @blur="onBlur"
          />
          <button @click="onSearch">
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

.search-bar-wrapper {
  position: relative;
  width: 560px;
  max-width: 90vw;
}

.search-bar {
  display: flex;
  align-items: center;
  border: 1px solid #ccc;
  border-radius: 24px;
  overflow: hidden;
  width: 100%;
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

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  margin-top: 8px;
  background: var(--bg, #fff);
  border: 1px solid #ccc;
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
</style>