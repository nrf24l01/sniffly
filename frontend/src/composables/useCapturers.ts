import { ref, computed, onMounted } from 'vue'
import { capturesService, type Capture } from '@/service/captures'

export function useCapturers() {
  const captures = ref<Capture[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const tokenVisibility = ref<Record<string, boolean>>({})

  const showCreate = ref(false)
  const showEdit = ref(false)
  const showDelete = ref(false)
  const showRegenerate = ref(false)

  const formName = ref('')
  const formEnabled = ref(true)
  const working = ref(false)

  const selected = ref<Capture | null>(null)

  const sortedCaptures = computed(() => {
    // Defensive: ensure array and string names
    const arr = Array.isArray(captures.value) ? captures.value : []
    return [...arr].sort((a, b) => (a?.name ?? '').localeCompare(b?.name ?? ''))
  })

  function resetForm() {
    formName.value = ''
    formEnabled.value = true
  }

  async function loadData() {
    loading.value = true
    error.value = null
    try {
      const data = await capturesService.list()
      captures.value = data
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      loading.value = false
    }
  }

  function toggleToken(id: string) {
    tokenVisibility.value = { ...tokenVisibility.value, [id]: !tokenVisibility.value[id] }
  }

  function openCreate() {
    resetForm()
    selected.value = null
    showCreate.value = true
  }

  function openEdit(cap: Capture) {
    selected.value = cap
    formName.value = cap.name
    formEnabled.value = cap.enabled
    showEdit.value = true
  }

  function openDelete(cap: Capture) {
    selected.value = cap
    showDelete.value = true
  }

  function openRegenerate(cap: Capture) {
    selected.value = cap
    showRegenerate.value = true
  }

  async function submitCreate() {
    if (!formName.value.trim()) return
    working.value = true
    error.value = null
    try {
      await capturesService.create({ name: formName.value.trim(), enabled: formEnabled.value })
      // Reload list from server to ensure fields are populated correctly
      await loadData()
      showCreate.value = false
      resetForm()
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      working.value = false
    }
  }

  async function submitEdit() {
    if (!selected.value) return
    if (!formName.value.trim()) return
    working.value = true
    error.value = null
    try {
      const updated = await capturesService.update(selected.value.uuid, { name: formName.value.trim(), enabled: formEnabled.value })
      captures.value = captures.value.map(c => (c.uuid === updated.uuid ? updated : c))
      showEdit.value = false
      selected.value = null
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      working.value = false
    }
  }

  async function confirmDelete() {
    if (!selected.value) return
    working.value = true
    error.value = null
    try {
      await capturesService.remove(selected.value.uuid)
      captures.value = captures.value.filter(c => c.uuid !== selected.value!.uuid)
      showDelete.value = false
      selected.value = null
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      working.value = false
    }
  }

  async function confirmRegenerate() {
    if (!selected.value) return
    working.value = true
    error.value = null
    try {
      const updated = await capturesService.regenerate(selected.value.uuid)
      captures.value = captures.value.map(c => (c.uuid === updated.uuid ? updated : c))
      tokenVisibility.value = { ...tokenVisibility.value, [updated.uuid]: true }
      showRegenerate.value = false
      selected.value = null
    } catch (e: any) {
      error.value = e?.response?.data?.message ?? e?.message ?? String(e)
    } finally {
      working.value = false
    }
  }

  onMounted(() => {
    void loadData()
  })

  return {
    // state
    captures,
    loading,
    error,
    tokenVisibility,
    showCreate,
    showEdit,
    showDelete,
    showRegenerate,
    formName,
    formEnabled,
    working,
    selected,
    sortedCaptures,

    // actions
    loadData,
    toggleToken,
    openCreate,
    openEdit,
    openDelete,
    openRegenerate,
    submitCreate,
    submitEdit,
    confirmDelete,
    confirmRegenerate,
    resetForm
  }
}
