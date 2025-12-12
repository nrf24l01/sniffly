<template>
  <nav class="relative bg-gray-100 text-gray-900 shadow-lg sticky top-0 z-50 dark:bg-gray-900 dark:text-gray-100">
    <!-- Logo — вынесен в левый край экрана -->
    <router-link to="/" class="absolute left-0 top-1/2 transform -translate-y-1/2 ml-4 text-2xl font-bold cursor-pointer hover:opacity-80 transition">
      <img src="/logo.png" alt="Logo" class="inline h-8 w-8 mr-2" />
      Sniffly
    </router-link>

    <!-- Центральный контейнер: все остальные элементы справа -->
    <div class="mx-auto px-4 py-4 flex justify-end items-center">
      <ul class="hidden md:flex space-x-6 items-center">
        <li v-for="link in navLinks" :key="link.name">
          <router-link
            :to="{ name: link.routeName }"
            class="hover:text-blue-200 transition duration-200"
          >
            {{ link.name }}
          </router-link>
        </li>
      </ul>
      <!-- Theme toggle (desktop) -->
      <button
        class="hidden md:inline-flex items-center ml-4 p-2 rounded hover:bg-gray-200 dark:hover:bg-gray-800 transition"
        @click="toggleTheme"
        :aria-label="`Toggle theme, current: ${theme}`"
      >
        <OutlineMoonIcon v-if="theme === 'light'" class="h-5 w-5" />
        <SolidMoonIcon v-else class="h-5 w-5" />
      </button>
      <!-- Mobile menu button -->
      <button class="md:hidden ml-2" @click="toggleMobileMenu">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
        </svg>
      </button>
    </div>

    <!-- Mobile menu -->
    <div v-if="mobileMenuOpen" class="md:hidden bg-blue-700 dark:bg-blue-900">
      <ul class="px-4 py-2 space-y-2">
        <li v-for="link in navLinks" :key="link.name">
          <router-link
            :to="{ name: link.routeName }"
            class="block py-2 hover:text-blue-200 dark:hover:text-blue-300"
            @click="mobileMenuOpen = false"
          >
            {{ link.name }}
          </router-link>
        </li>
      </ul>
      <div class="px-4 py-3 border-t border-blue-600">
        <button
          class="w-full flex items-center justify-center gap-2 px-3 py-2 rounded bg-white bg-opacity-10 hover:bg-opacity-20 transition text-black dark:bg-gray-900 dark:bg-opacity-5 dark:text-white"
          @click="toggleTheme(); mobileMenuOpen = false"
          :aria-label="`Toggle theme, current: ${theme}`"
        >
          <OutlineMoonIcon v-if="theme === 'light'" class="h-5 w-5" />
          <SolidMoonIcon v-else class="h-5 w-5" />
          <span class="text-sm">Toggle theme</span>
        </button>
      </div>
    </div>
  </nav>
</template>

<script setup>
import { ref } from 'vue'
import { useTheme } from '../composables/useTheme'
import { MoonIcon as OutlineMoonIcon } from "@heroicons/vue/24/outline"
import { MoonIcon as SolidMoonIcon } from "@heroicons/vue/24/solid"

const mobileMenuOpen = ref(false)

const { theme, toggleTheme } = useTheme()

const navLinks = [
    { name: 'Home', routeName: 'Home' }
]

const toggleMobileMenu = () => {
    mobileMenuOpen.value = !mobileMenuOpen.value
}
</script>