<script setup lang="ts">
import Navbar from './components/Navbar.vue'
</script>

<template>
  <div class="h-screen flex flex-col">
    <Navbar />
    <div class="flex-1 min-h-0">
      <router-view />
    </div>
  </div>
</template>

<style>
  @import "tailwindcss";
  @custom-variant dark (&:where([data-theme=dark], [data-theme=dark] *));
</style>
