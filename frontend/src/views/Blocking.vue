<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ArrowPathIcon, NoSymbolIcon, TrashIcon } from '@heroicons/vue/24/outline'

import { blockingService, type BlockRuleTargetType, type DeviceBlockRuleItem } from '@/service/blocking'
import { devicesService, type DeviceListItem } from '@/service/devices'

const devices = ref<DeviceListItem[]>([])
const rules = ref<DeviceBlockRuleItem[]>([])
const selectedDeviceId = ref('')

const loading = ref(false)
const saving = ref(false)
const removingRuleId = ref<string | null>(null)
const error = ref<string | null>(null)
const formError = ref<string | null>(null)

const targetType = ref<BlockRuleTargetType>('ip')
const targetValue = ref('')
const enabled = ref(true)

const targetTypeOptions: Array<{ value: BlockRuleTargetType; label: string; hint: string }> = [
  { value: 'ip', label: 'IP-адрес', hint: 'Например, 8.8.8.8' },
  { value: 'cidr', label: 'Подсеть', hint: 'Например, 10.0.0.0/24' },
  { value: 'sni', label: 'SNI / Host', hint: 'Например, api.example.com' }
]

const sortedDevices = computed(() => {
  return [...devices.value].sort((a, b) => {
    const left = (a.user_label || '').trim() || a.mac
    const right = (b.user_label || '').trim() || b.mac
    return left.localeCompare(right)
  })
})

const selectedDevice = computed(() => sortedDevices.value.find(device => device.uuid === selectedDeviceId.value) ?? null)

const filteredRules = computed(() => {
  return rules.value
    .filter(rule => !selectedDeviceId.value || rule.device_uuid === selectedDeviceId.value)
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
})

const activeRulesCount = computed(() => filteredRules.value.filter(rule => rule.enabled).length)

watch(sortedDevices, next => {
  if (!next.length) {
    selectedDeviceId.value = ''
    return
  }
  const firstDevice = next[0]
  if (!firstDevice) {
    selectedDeviceId.value = ''
    return
  }
  if (!selectedDeviceId.value || !next.some(device => device.uuid === selectedDeviceId.value)) {
    selectedDeviceId.value = firstDevice.uuid
  }
}, { immediate: true })

function normalizeTargetValue(value: string) {
  const trimmed = value.trim()
  return targetType.value === 'sni' ? trimmed.toLowerCase() : trimmed
}

async function loadData() {
  loading.value = true
  error.value = null
  try {
    const [devicesData, rulesData] = await Promise.all([devicesService.list(), blockingService.list()])
    devices.value = Array.isArray(devicesData) ? devicesData : []
    rules.value = Array.isArray(rulesData) ? rulesData : []
  } catch (e: any) {
    error.value = e?.response?.data?.message ?? e?.message ?? String(e)
  } finally {
    loading.value = false
  }
}

async function createRule() {
  if (!selectedDeviceId.value) {
    formError.value = 'Сначала выберите устройство.'
    return
  }
  const normalizedValue = normalizeTargetValue(targetValue.value)
  if (!normalizedValue) {
    formError.value = 'Укажите значение правила.'
    return
  }

  saving.value = true
  formError.value = null
  try {
    const created = await blockingService.create(selectedDeviceId.value, {
      target_type: targetType.value,
      target_value: normalizedValue,
      enabled: enabled.value
    })
    rules.value = [created, ...rules.value.filter(rule => rule.uuid !== created.uuid)]
    targetValue.value = ''
    enabled.value = true
  } catch (e: any) {
    formError.value = e?.response?.data?.message ?? e?.message ?? String(e)
  } finally {
    saving.value = false
  }
}

async function toggleRule(rule: DeviceBlockRuleItem, next: boolean) {
  try {
    const updated = await blockingService.update(rule.device_uuid, rule.uuid, { enabled: next })
    rules.value = rules.value.map(item => (item.uuid === updated.uuid ? updated : item))
  } catch (e: any) {
    error.value = e?.response?.data?.message ?? e?.message ?? String(e)
  }
}

async function removeRule(rule: DeviceBlockRuleItem) {
  removingRuleId.value = rule.uuid
  error.value = null
  try {
    await blockingService.remove(rule.device_uuid, rule.uuid)
    rules.value = rules.value.filter(item => item.uuid !== rule.uuid)
  } catch (e: any) {
    error.value = e?.response?.data?.message ?? e?.message ?? String(e)
  } finally {
    removingRuleId.value = null
  }
}

function typeLabel(type: BlockRuleTargetType) {
  return targetTypeOptions.find(option => option.value === type)?.label ?? type
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <main class="min-h-full overflow-auto bg-slate-50 dark:bg-slate-900">
    <div class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <header class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <p class="text-xs font-semibold uppercase tracking-[0.24em] text-green-700 dark:text-green-300">Traffic control</p>
          <h1 class="mt-2 text-3xl font-semibold tracking-tight text-slate-900 dark:text-slate-50">Блокировка трафика устройств</h1>
        </div>

        <button
          class="inline-flex items-center justify-center gap-2 self-start rounded-2xl border border-slate-200/70 bg-white/80 px-4 py-2 text-sm font-semibold text-slate-900 shadow-sm backdrop-blur transition hover:bg-white disabled:cursor-not-allowed disabled:opacity-60 dark:border-slate-800 dark:bg-slate-900/60 dark:text-slate-50 dark:hover:bg-slate-900"
          :disabled="loading"
          @click="loadData"
        >
          <ArrowPathIcon class="h-4 w-4" :class="loading ? 'animate-spin' : ''" />
          {{ loading ? 'Обновление…' : 'Обновить правила' }}
        </button>
      </header>

      <div v-if="error" class="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
        {{ error }}
      </div>

      <section class="mt-6 grid gap-4 xl:grid-cols-[360px_minmax(0,1fr)]">
        <aside class="rounded-[28px] border border-slate-200/70 bg-white/75 p-5 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/55">
          <div class="flex items-center justify-between gap-3">
            <div>
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Устройство</div>
              <div class="mt-1 text-xs text-slate-500 dark:text-slate-300">Выберите источник трафика, для которого создаются правила.</div>
            </div>
            <div class="rounded-2xl bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">{{ filteredRules.length }} правил</div>
          </div>

          <div class="mt-4 space-y-2">
            <button
              v-for="device in sortedDevices"
              :key="device.uuid"
              class="w-full rounded-2xl border px-4 py-3 text-left transition"
              :class="device.uuid === selectedDeviceId ? 'border-green-500 bg-green-50 text-green-900 shadow-sm dark:border-green-400 dark:bg-green-500/10 dark:text-green-100' : 'border-slate-200/70 bg-white/70 text-slate-800 hover:border-slate-300 hover:bg-white dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-100 dark:hover:bg-slate-900'"
              @click="selectedDeviceId = device.uuid"
            >
              <div class="truncate text-sm font-semibold">{{ device.user_label || device.mac }}</div>
              <div class="mt-1 truncate text-xs opacity-80">{{ device.mac }} • {{ device.ip }}</div>
            </button>
          </div>

          <div v-if="!sortedDevices.length && !loading" class="mt-4 rounded-2xl border border-dashed border-slate-300 px-4 py-5 text-sm text-slate-500 dark:border-slate-700 dark:text-slate-300">
            Устройства ещё не появились в `device_info`.
          </div>
        </aside>

        <div class="space-y-4">
          <section class="rounded-[28px] border border-slate-200/70 bg-white/80 p-5 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/60">
            <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div>
                <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Новое правило</div>
                <div class="mt-1 text-xs text-slate-500 dark:text-slate-300">
                  {{ selectedDevice ? `${selectedDevice.user_label || selectedDevice.mac} • ${selectedDevice.ip}` : 'Сначала выберите устройство' }}
                </div>
              </div>

              <div class="grid gap-2 sm:grid-cols-3">
                <div class="rounded-2xl bg-slate-100 px-3 py-2 text-xs font-semibold text-slate-700 dark:bg-slate-800 dark:text-slate-200">Всего: {{ filteredRules.length }}</div>
                <div class="rounded-2xl bg-green-100 px-3 py-2 text-xs font-semibold text-green-800 dark:bg-green-500/10 dark:text-green-200">Активных: {{ activeRulesCount }}</div>
                <div class="rounded-2xl bg-amber-100 px-3 py-2 text-xs font-semibold text-amber-800 dark:bg-amber-500/10 dark:text-amber-200">Выключено: {{ filteredRules.length - activeRulesCount }}</div>
              </div>
            </div>

            <div v-if="formError" class="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
              {{ formError }}
            </div>

            <div class="mt-4 grid gap-4 lg:grid-cols-[minmax(0,220px)_minmax(0,1fr)_auto]">
              <div class="space-y-2">
                <label class="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-300">Тип правила</label>
                <div class="grid gap-2">
                  <button
                    v-for="option in targetTypeOptions"
                    :key="option.value"
                    class="rounded-2xl border px-4 py-3 text-left transition"
                    :class="targetType === option.value ? 'border-slate-900 bg-slate-900 text-white dark:border-slate-100 dark:bg-slate-100 dark:text-slate-900' : 'border-slate-200/70 bg-white text-slate-800 hover:border-slate-300 dark:border-slate-800 dark:bg-slate-900/40 dark:text-slate-100'"
                    @click="targetType = option.value"
                  >
                    <div class="text-sm font-semibold">{{ option.label }}</div>
                    <div class="mt-1 text-xs opacity-80">{{ option.hint }}</div>
                  </button>
                </div>
              </div>

              <div>
                <label class="text-xs font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-300">Значение</label>
                <textarea
                  v-model="targetValue"
                  rows="4"
                  class="mt-2 w-full rounded-3xl border border-slate-200/70 bg-white px-4 py-3 text-sm text-slate-900 shadow-sm outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50"
                  :placeholder="targetTypeOptions.find(option => option.value === targetType)?.hint"
                />
                <label class="mt-3 inline-flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-200">
                  <input v-model="enabled" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500" />
                  Включить правило сразу после создания
                </label>
              </div>

              <div class="flex items-end">
                <button
                  class="inline-flex w-full items-center justify-center gap-2 rounded-3xl bg-green-600 px-5 py-3 text-sm font-semibold text-white shadow-sm transition hover:bg-green-700 disabled:cursor-not-allowed disabled:bg-green-400 lg:w-auto"
                  :disabled="saving || !selectedDeviceId || !targetValue.trim()"
                  @click="createRule"
                >
                  <NoSymbolIcon class="h-4 w-4" />
                  {{ saving ? 'Сохранение…' : 'Добавить блокировку' }}
                </button>
              </div>
            </div>
          </section>

          <section class="rounded-[28px] border border-slate-200/70 bg-white/80 p-5 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/60">
            <div class="flex items-center justify-between gap-3">
              <div>
                <div class="text-sm font-semibold text-slate-900 dark:text-slate-50">Текущие правила</div>
              </div>
            </div>

            <div v-if="filteredRules.length" class="mt-4 space-y-3">
              <article
                v-for="rule in filteredRules"
                :key="rule.uuid"
                class="rounded-3xl border border-slate-200/70 bg-white/80 p-4 dark:border-slate-800 dark:bg-slate-900/50"
              >
                <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700 dark:bg-slate-800 dark:text-slate-200">{{ typeLabel(rule.target_type) }}</span>
                      <span :class="rule.enabled ? 'bg-green-100 text-green-800 dark:bg-green-500/10 dark:text-green-200' : 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-300'" class="rounded-full px-3 py-1 text-xs font-semibold">
                        {{ rule.enabled ? 'Активно' : 'Выключено' }}
                      </span>
                    </div>
                    <div class="mt-3 break-all font-mono text-sm text-slate-900 dark:text-slate-50">{{ rule.target_value }}</div>
                    <div class="mt-3 text-xs text-slate-500 dark:text-slate-300">
                      {{ rule.device_label || rule.device_mac }} • {{ rule.device_mac }} • {{ rule.device_ip }}
                    </div>
                  </div>

                  <div class="flex flex-wrap items-center gap-3">
                    <label class="inline-flex items-center gap-2 text-sm font-medium text-slate-700 dark:text-slate-200">
                      <input
                        :checked="rule.enabled"
                        type="checkbox"
                        class="h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500"
                        @change="toggleRule(rule, ($event.target as HTMLInputElement).checked)"
                      />
                      Включено
                    </label>

                    <button
                      class="inline-flex items-center gap-2 rounded-2xl border border-red-200 bg-red-50 px-3 py-2 text-sm font-semibold text-red-700 transition hover:bg-red-100 disabled:cursor-not-allowed disabled:opacity-60 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200 dark:hover:bg-red-950/40"
                      :disabled="removingRuleId === rule.uuid"
                      @click="removeRule(rule)"
                    >
                      <TrashIcon class="h-4 w-4" />
                      {{ removingRuleId === rule.uuid ? 'Удаление…' : 'Удалить' }}
                    </button>
                  </div>
                </div>
              </article>
            </div>

            <div v-else class="mt-4 rounded-3xl border border-dashed border-slate-300 px-6 py-10 text-center text-sm text-slate-500 dark:border-slate-700 dark:text-slate-300">
              Для выбранного устройства пока нет правил блокировки.
            </div>
          </section>
        </div>
      </section>
    </div>
  </main>
</template>