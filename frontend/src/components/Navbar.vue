<template>
    <nav class="bg-gray-100 text-gray-900 shadow-lg sticky top-0 z-50">
        <div class="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
            <router-link to="/" class="text-2xl font-bold cursor-pointer hover:opacity-80 transition">
                <img src="/logo.png" alt="Logo" class="inline h-8 w-8 mr-2" />
                Sniffly
            </router-link>
            <ul class="hidden md:flex space-x-6">
                <li v-for="link in navLinks" :key="link.name">
                    <router-link
                        :to="{ name: link.routeName }"
                        class="hover:text-blue-200 transition duration-200"
                    >
                        {{ link.name }}
                    </router-link>
                </li>
            </ul>
            <!-- Mobile menu button -->
            <button class="md:hidden" @click="toggleMobileMenu">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
                </svg>
            </button>
        </div>
        <!-- Mobile menu -->
        <div v-if="mobileMenuOpen" class="md:hidden bg-blue-700">
            <ul class="px-4 py-2 space-y-2">
                <li v-for="link in navLinks" :key="link.name">
                    <router-link
                        :to="{ name: link.routeName }"
                        class="block py-2 hover:text-blue-200"
                        @click="mobileMenuOpen = false"
                    >
                        {{ link.name }}
                    </router-link>
                </li>
            </ul>
        </div>
    </nav>
</template>

<script setup>
import { ref } from 'vue'

const mobileMenuOpen = ref(false)

const navLinks = [
  { name: 'Home', routeName: 'Home' }
]

const toggleMobileMenu = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value
}
</script>