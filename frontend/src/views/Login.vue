<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLogin } from '@/composables/login'
import { useTheme } from '@/composables/useTheme'
import { useRouter } from 'vue-router'

const router = useRouter()

const nickname = ref('')
const password = ref('')

// validation / touched state
const nicknameTouched = ref(false)
const passwordTouched = ref(false)
const submissionAttempted = ref(false)

const { loading, error, login } = useLogin()
const { theme, toggleTheme } = useTheme()

const themeButtonLabel = computed(() => (theme.value === 'dark' ? 'Light mode' : 'Dark mode'))

const canSubmit = computed(() => {
	return nickname.value.trim().length > 0 && password.value.length > 0 && !loading.value
})

function onNicknameBlur() {
	nicknameTouched.value = true
}

function onPasswordBlur() {
	passwordTouched.value = true
}

const nicknameInvalid = computed(
	() => nickname.value.trim().length === 0 && (nicknameTouched.value || submissionAttempted.value)
)
const passwordInvalid = computed(
	() => password.value.length === 0 && (passwordTouched.value || submissionAttempted.value)
)

async function onSubmit() {
	// mark that user attempted submit so empty fields can be highlighted
	submissionAttempted.value = true
	if (!canSubmit.value) return
	await login(nickname.value.trim(), password.value)
	if (!error.value) password.value = ''
	// reset touched/submission state on successful login
	if (!error.value) {
		submissionAttempted.value = false
		nicknameTouched.value = false
    passwordTouched.value = false
    router.push({ name: 'Home' })
	}
}
</script>

<template>
	<main
		class="h-full w-full px-4 py-10 sm:px-6 lg:px-8 flex items-center justify-center bg-gradient-to-b from-green-50 to-white dark:from-slate-950 dark:to-slate-900"
	>
		<div class="mx-auto w-full max-w-md">
			<section
				class="rounded-2xl border border-slate-200/70 bg-white/80 p-6 shadow-sm backdrop-blur dark:border-slate-800 dark:bg-slate-900/70 sm:p-8"
			>
				<header>
					<h1 class="text-2xl font-semibold tracking-tight text-slate-900 dark:text-slate-50">Sign in</h1>
					<p class="mt-1 text-sm text-slate-600 dark:text-slate-300">
						Use your nickname and password.
					</p>
				</header>

				<form class="mt-6 space-y-4" @submit.prevent="onSubmit">
					<div>
						<label :class="['block text-sm font-medium text-slate-700 dark:text-slate-200', { 'text-red-700 dark:text-red-300': nicknameInvalid }]" for="nickname">
							Nickname <span class="text-red-500 ml-1" aria-hidden="true">*</span>
						</label>
						<input
							id="nickname"
							v-model="nickname"
							type="text"
							autocomplete="username"
							inputmode="text"
							:class="[
								'mt-1 block w-full rounded-xl border bg-white px-3 py-2 text-slate-900 outline-none transition placeholder:text-slate-400 dark:bg-slate-950/40 dark:text-slate-100 dark:placeholder:text-slate-500',
								{ 'border-slate-300 focus:border-green-500 focus:ring-2 focus:ring-green-500/30 dark:border-slate-700': !nicknameInvalid, 'border-red-500 focus:border-red-500 focus:ring-red-500/30 dark:border-red-700': nicknameInvalid }
							]"
							placeholder="your_nickname"
							@blur="onNicknameBlur"
							:aria-invalid="nicknameInvalid ? 'true' : 'false'"
						/>
					</div>

					<div>
						<label :class="['block text-sm font-medium text-slate-700 dark:text-slate-200', { 'text-red-700 dark:text-red-300': passwordInvalid }]" for="password">
							Password <span class="text-red-500 ml-1" aria-hidden="true">*</span>
						</label>
						<input
							id="password"
							v-model="password"
							type="password"
							autocomplete="current-password"
							:class="[
								'mt-1 block w-full rounded-xl border bg-white px-3 py-2 text-slate-900 outline-none transition placeholder:text-slate-400 dark:bg-slate-950/40 dark:text-slate-100 dark:placeholder:text-slate-500',
								{ 'border-slate-300 focus:border-green-500 focus:ring-2 focus:ring-green-500/30 dark:border-slate-700': !passwordInvalid, 'border-red-500 focus:border-red-500 focus:ring-red-500/30 dark:border-red-700': passwordInvalid }
							]"
							placeholder="••••••••"
							@blur="onPasswordBlur"
							:aria-invalid="passwordInvalid ? 'true' : 'false'"
						/>
					</div>

					<p v-if="error" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-200">
						{{ error }}
					</p>

					<button
						type="submit"
						:disabled="!canSubmit"
						class="group relative inline-flex w-full items-center justify-center gap-2 rounded-xl bg-green-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition hover:bg-green-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-green-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:cursor-not-allowed disabled:bg-green-400 dark:focus-visible:ring-offset-slate-900"
					>
						<span
							v-if="loading"
							class="h-4 w-4 animate-spin rounded-full border-2 border-white/40 border-t-white"
							aria-hidden="true"
						/>
						<span>{{ loading ? 'Signing in…' : 'Sign in' }}</span>
					</button>
				</form>

				<p class="mt-6 text-center text-xs text-slate-500 dark:text-slate-400">
					By signing in, you’ll get access to your dashboard.
				</p>
			</section>
		</div>
	</main>
</template>
