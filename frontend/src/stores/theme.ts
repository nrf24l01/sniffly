import { defineStore } from 'pinia'
import { ref } from 'vue'

type Theme = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  const theme = ref<Theme>('light')

  function applyTheme(t: Theme) {
    const html = document.documentElement
    html.setAttribute('data-theme', t)

    try {
      localStorage.setItem('theme', t)
    } catch (e) {
      // ignore
    }

    theme.value = t
  }

  function toggleTheme() {
    applyTheme(theme.value === 'dark' ? 'light' : 'dark')
  }

  function initTheme() {
    // prefer saved value, otherwise follow system preference
    try {
      const saved = localStorage.getItem('theme') as Theme | null
      if (saved) {
        applyTheme(saved)
        return
      }
    } catch (e) {
      // ignore
    }
  }

  return { theme, toggleTheme, setTheme: applyTheme, initTheme }
})
