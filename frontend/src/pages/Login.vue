<script setup lang="ts">
import { ref } from "vue";
import { login } from "../api/auth";
import { useRouter } from "vue-router";

const email = ref("");
const password = ref("");
const router = useRouter();

async function onLogin() {
  await login(email.value, password.value);
  router.push("/");
}

function goToRegister() {
  router.push("/register");
}
</script>

<template>
  <div>
    <h1>Login</h1>

    <input v-model="email" placeholder="email" />
    <input v-model="password" type="password" placeholder="password" />

    <button @click="onLogin">Login</button>
    <button @click="goToRegister">Register</button>
  </div>
</template>