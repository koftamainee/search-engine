<script setup lang="ts">
import { onMounted, ref } from "vue";
import { me, logout } from "../api/auth";
import { useRouter } from "vue-router";

const user = ref<any>(null);
const router = useRouter();

onMounted(async () => {
  try {
    const res = await me();
    user.value = res.data;
  } catch {
    router.push("/login");
  }
});

async function onLogout() {
  await logout();
  router.push("/login");
}
</script>

<template>
  <div>
    <h1>Search engine</h1>

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
      <button @click="onLogout">Logout</button>
    </div>
  </div>
</template>