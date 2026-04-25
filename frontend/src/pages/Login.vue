<script setup lang="ts">
import { ref } from "vue";
import { login } from "../api/auth";
import { useRouter } from "vue-router";

const email = ref("");
const password = ref("");
const errorMessage = ref("");
const isLoading = ref(false);
const router = useRouter();

async function onLogin() {
  errorMessage.value = "";
  isLoading.value = true;

  try {
    await login(email.value, password.value);
    router.push("/");
  } catch (error: any) {
    if (error.response?.status === 403) {
      errorMessage.value = "Your account has been blocked. Please contact support.";
    } else if (error.response?.status === 401) {
      errorMessage.value = "Invalid email or password.";
    } else {
      errorMessage.value = "Login error. Please try again later.";
    }
  } finally {
    isLoading.value = false;
  }
}

function goToRegister() {
  router.push("/register");
}

function clearError() {
  errorMessage.value = "";
}
</script>

<template>
  <div>
    <h1>Login</h1>

    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
      <button @click="clearError" class="close-btn">×</button>
    </div>

    <input
      v-model="email"
      placeholder="Email"
      type="email"
      :disabled="isLoading"
    />

    <input
      v-model="password"
      type="password"
      placeholder="Password"
      :disabled="isLoading"
    />

    <button @click="onLogin" :disabled="isLoading">
      {{ isLoading ? "Logging in..." : "Login" }}
    </button>

    <button @click="goToRegister" :disabled="isLoading">
      Register
    </button>
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

input, button {
  display: block;
  margin: 10px 0;
  padding: 8px 12px;
}

button {
  cursor: pointer;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>