<script setup lang="ts">
import {
  EyeIcon,
  EyeSlashIcon,
  PencilSquareIcon,
  PlusIcon,
  ArrowPathIcon,
  TrashIcon,
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/vue/24/outline'
import { useCapturers } from '@/composables/useCapturers'

const {
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
  confirmRegenerate
} = useCapturers()
</script>

<template>
  <main class="h-full overflow-auto bg-gradient-to-b from-green-50 to-white dark:from-slate-950 dark:to-slate-900">
    <div class="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 lg:px-8">
      <header class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold tracking-tight text-slate-900 dark:text-slate-50">Захватчики трафика</h1>
          <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">Управление захватчиками трафика и их токенами.</p>
        </div>
        <button
          class="inline-flex items-center gap-2 rounded-xl bg-green-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-green-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-slate-900"
          @click="openCreate"
        >
          <PlusIcon class="h-5 w-5" />
          Создать
        </button>
      </header>

      <section v-if="error" class="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
        {{ error }}
      </section>

      <section class="mt-6 rounded-2xl border border-slate-200/70 bg-white/70 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/60">
        <div class="flex items-center justify-between border-b border-slate-200/70 px-4 py-3 text-sm font-medium text-slate-600 dark:border-slate-800 dark:text-slate-300">
          <div class="flex items-center gap-2">
            <ArrowPathIcon v-if="loading" class="h-5 w-5 animate-spin" />
            <span>{{ loading ? 'Загрузка' : 'Список захватчиков трафика' }}</span>
          </div>
          <button class="text-xs font-semibold text-slate-500 hover:text-slate-800 dark:text-slate-300 dark:hover:text-slate-100" @click="loadData" :disabled="loading">
            Обновить
          </button>
        </div>

        <div v-if="!loading && sortedCaptures.length === 0" class="px-6 py-8 text-center text-sm text-slate-500 dark:text-slate-300">
          Нет созданных захватчиков трафика. Нажмите «Создать», чтобы добавить первый.
        </div>

        <div v-else class="overflow-x-auto">
          <table class="min-w-full text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-800/60 dark:text-slate-300">
              <tr>
                <th class="px-4 py-3 text-left">Название</th>
                <th class="px-4 py-3 text-left">UUID</th>
                <th class="px-4 py-3 text-left">Ключ</th>
                <th class="px-4 py-3 text-left">Состояние</th>
                <th class="px-4 py-3 text-right">Действия</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200/70 dark:divide-slate-800">
              <tr v-for="cap in sortedCaptures" :key="cap.uuid" class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40">
                <td class="px-4 py-3 font-medium text-slate-900 dark:text-slate-50">{{ cap.name }}</td>
                <td class="px-4 py-3 text-slate-700 dark:text-slate-200">
                  <div class="font-mono text-xs">{{ cap.uuid }}</div>
                </td>
                <td class="px-4 py-3 text-slate-700 dark:text-slate-200">
                  <div class="flex items-center gap-2 font-mono text-xs">
                    <span>{{ tokenVisibility[cap.uuid] ? cap.api_key : '••••••••••••••••' }}</span>
                    <button
                      class="rounded-lg p-1 text-slate-500 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-slate-50"
                      @click="toggleToken(cap.uuid)"
                      :aria-label="tokenVisibility[cap.uuid] ? 'Hide token' : 'Show token'"
                    >
                      <EyeSlashIcon v-if="tokenVisibility[cap.uuid]" class="h-4 w-4" />
                      <EyeIcon v-else class="h-4 w-4" />
                    </button>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center gap-1 rounded-full px-2 py-1 text-xs font-semibold"
                    :class="cap.enabled ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-200' : 'bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-200'"
                  >
                    <CheckCircleIcon v-if="cap.enabled" class="h-4 w-4" />
                    <XCircleIcon v-else class="h-4 w-4" />
                    {{ cap.enabled ? 'Включен' : 'Выключен' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <div class="flex justify-end gap-2">
                    <button
                      class="inline-flex items-center gap-1 rounded-lg border border-slate-200/70 px-3 py-1.5 text-xs font-semibold text-slate-800 transition hover:bg-slate-100 dark:border-slate-800 dark:text-slate-100 dark:hover:bg-slate-800"
                      @click="openEdit(cap)"
                    >
                      <PencilSquareIcon class="h-4 w-4" />
                      Редактировать
                    </button>
                    <button
                      class="inline-flex items-center gap-1 rounded-lg border border-slate-200/70 px-3 py-1.5 text-xs font-semibold text-slate-800 transition hover:bg-slate-100 dark:border-slate-800 dark:text-slate-100 dark:hover:bg-slate-800"
                      @click="openRegenerate(cap)"
                    >
                      <ArrowPathIcon class="h-4 w-4" />
                      Перегенерировать
                    </button>
                    <button
                      class="inline-flex items-center gap-1 rounded-lg border border-red-200 bg-red-50 px-3 py-1.5 text-xs font-semibold text-red-700 transition hover:bg-red-100 dark:border-red-800 dark:bg-red-950/40 dark:text-red-200"
                      @click="openDelete(cap)"
                    >
                      <TrashIcon class="h-4 w-4" />
                      Удалить
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <div v-if="showCreate || showEdit || showDelete || showRegenerate" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 px-4">
      <div class="w-full max-w-lg rounded-2xl border border-slate-200/70 bg-white p-6 shadow-2xl dark:border-slate-800 dark:bg-slate-900">
        <template v-if="showCreate || showEdit">
          <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">{{ showCreate ? 'Создать захватчика трафика' : 'Редактировать захватчика трафика' }}</h2>
          <p class="mt-1 text-sm text-slate-600 dark:text-slate-300">Задайте имя и состояние. Токен генерируется автоматически.</p>

          <div class="mt-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-700 dark:text-slate-200">Название</label>
              <input
                v-model="formName"
                type="text"
                class="mt-1 w-full rounded-xl border border-slate-200/70 bg-white px-3 py-2 text-sm text-slate-900 shadow-sm outline-none transition focus:border-green-500 focus:ring-2 focus:ring-green-500/20 dark:border-slate-800 dark:bg-slate-900/50 dark:text-slate-50"
                placeholder="Например, office-router"
              />
            </div>

            <label class="flex items-center gap-3 text-sm font-medium text-slate-700 dark:text-slate-200">
              <input type="checkbox" v-model="formEnabled" class="h-4 w-4 rounded border-slate-300 text-green-600 focus:ring-green-500" />
              Включен
            </label>
          </div>

          <div class="mt-6 flex justify-end gap-3">
            <button class="rounded-xl px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800" @click="showCreate = false; showEdit = false">
              Отмена
            </button>
            <button
              class="inline-flex items-center justify-center gap-2 rounded-xl bg-green-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-green-700 disabled:cursor-not-allowed disabled:bg-green-400"
              :disabled="working || !formName.trim()"
              @click="showCreate ? submitCreate() : submitEdit()"
            >
              <ArrowPathIcon v-if="working" class="h-4 w-4 animate-spin" />
              <span>{{ showCreate ? 'Создать' : 'Сохранить' }}</span>
            </button>
          </div>
        </template>

        <template v-else-if="showDelete">
          <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Удалить захватчика трафика?</h2>
          <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">
            Будет удалён «{{ selected?.name }}». Это действие нельзя отменить.
          </p>
          <div class="mt-6 flex justify-end gap-3">
            <button class="rounded-xl px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800" @click="showDelete = false">
              Отмена
            </button>
            <button
              class="inline-flex items-center justify-center gap-2 rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-red-700 disabled:cursor-not-allowed disabled:bg-red-400"
              :disabled="working"
              @click="confirmDelete"
            >
              <ArrowPathIcon v-if="working" class="h-4 w-4 animate-spin" />
              <span>Удалить</span>
            </button>
          </div>
        </template>

        <template v-else-if="showRegenerate">
          <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Регенерировать ключ?</h2>
          <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">
            Для «{{ selected?.name }}» будет создан новый ключ. Старый станет недействителен.
          </p>
          <div class="mt-6 flex justify-end gap-3">
            <button class="rounded-xl px-4 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800" @click="showRegenerate = false">
              Отмена
            </button>
            <button
              class="inline-flex items-center justify-center gap-2 rounded-xl bg-amber-600 px-4 py-2 text-sm font-semibold text-white shadow-sm transition hover:bg-amber-700 disabled:cursor-not-allowed disabled:bg-amber-400"
              :disabled="working"
              @click="confirmRegenerate"
            >
              <ArrowPathIcon v-if="working" class="h-4 w-4 animate-spin" />
              <span>Регенерировать</span>
            </button>
          </div>
        </template>
      </div>
    </div>
  </main>
</template>
