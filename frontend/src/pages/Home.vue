<script setup lang="ts">
import { onMounted, ref } from "vue";
import { me, logout } from "../api/auth";
import { useRouter } from "vue-router";

const user = ref<any>(null);
const errorMessage = ref("");
const isLoading = ref(false);
const router = useRouter();

onMounted(async () => {
  try {
    const res = await me();
    user.value = res.data;
  } catch (error: any) {
    if (error.response?.status === 403) {
      errorMessage.value = "Your account has been blocked. Please contact support.";
    } else {
      router.push("/login");
    }
  }
});

async function onLogout() {
  isLoading.value = true;
  try {
    await logout();
    router.push("/login");
  } catch (error: any) {
    errorMessage.value = "Logout error. Please try again.";
  } finally {
    isLoading.value = false;
  }
}

function clearError() {
  errorMessage.value = "";
}
</script>

<template>
  <div>
    <h1>Search engine</h1>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
      <button @click="clearError" class="close-btn">×</button>
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
      <button @click="onLogout" :disabled="isLoading">
        {{ isLoading ? "Logging out..." : "Logout" }}
      </button>
    </div>
  </div>
</template>

<style scoped>
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
}

ul {
  list-style: none;
  padding: 0;
}

li {
  margin: 8px 0;
}

button {
  padding: 8px 16px;
  cursor: pointer;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>