import { onMounted } from 'vue'
import { useThemeStore } from '../stores/theme'
import { storeToRefs } from 'pinia'

export function useTheme() {
  const store = useThemeStore()

  // initialize theme on first use (safe to call multiple times)
  onMounted(() => {
    store.initTheme()
  })

  // Use storeToRefs to return stable refs that keep reactivity when destructured
  const { theme } = storeToRefs(store)

  return {
    theme,
    toggleTheme: store.toggleTheme,
    setTheme: store.setTheme,
  }
}
