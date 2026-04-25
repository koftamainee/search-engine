<script setup lang="ts">
import { ref } from "vue";
import { register } from "../api/auth";
import { useRouter } from "vue-router";

const email = ref("");
const password = ref("");
const router = useRouter();

async function onRegister() {
  await register(email.value, password.value);
  router.push("/login");
}

function goToLogin() {
  router.push("/login");
}
</script>

<template>
  <div>
    <h1>Register</h1>

    <input v-model="email" placeholder="email" />
    <input v-model="password" type="password" placeholder="password" />

    <button @click="onRegister">Create account</button>
    <button @click="goToLogin">Login</button>
  </div>
</template>