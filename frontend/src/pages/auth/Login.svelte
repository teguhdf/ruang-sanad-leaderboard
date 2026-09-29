<script lang="ts">
	import { inertia, useForm } from "@inertiajs/svelte";
	import { ArrowRight, Eye, EyeOff, Lock, Mail } from "lucide-svelte";
	import AuthLayout from "@layouts/AuthLayout.svelte";
	import type { Flash } from "@lib/types";

	const form = useForm({ email: "", password: "" });
	let showPassword = $state(false);

	interface Props { flash?: Flash }
	let { flash }: Props = $props();

	function submitForm(e: Event) {
		e.preventDefault();
		form.post("/login");
	}
</script>

<AuthLayout
	title="Masuk"
	subtitle="Gunakan akun yang diberikan Ruang Sanad"
	headline="Ruang Sanad Ops"
	subheadline="Workspace internal untuk absensi, roster, performa, dan leaderboard tim. Akun employee dibuat melalui undangan SPV."
	stats={[
		{ value: "44j", label: "Target mingguan" },
		{ value: "8j", label: "Hari normal" },
		{ value: "4j", label: "Sabtu" },
	]}
	{flash}
>
	<div class="mb-6 rounded-2xl bg-brand-50 px-4 py-3 text-sm leading-6 text-brand-800 ring-1 ring-brand-100 dark:bg-brand-950/30 dark:text-brand-200 dark:ring-brand-900/50">
		Belum punya akun? Hubungi SPV. Registrasi publik dinonaktifkan agar akses employee tetap terkontrol.
	</div>

	<form class="space-y-5" onsubmit={submitForm}>
		<div class="space-y-2">
			<label for="email" class="block text-sm font-medium text-neutral-700 dark:text-neutral-300">Email</label>
			<div class="relative">
				<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4"><Mail class="h-5 w-5 text-neutral-500" /></div>
				<input bind:value={form.email} required type="email" name="email" id="email" class="w-full rounded-2xl border border-neutral-300 bg-white py-3 pl-12 pr-4 text-neutral-900 placeholder-neutral-400 transition-colors focus:border-brand-400 focus:outline-none focus:ring-2 focus:ring-brand-400/20 dark:border-neutral-700/80 dark:bg-neutral-900 dark:text-white" placeholder="nama@ruangsanad.com" />
			</div>
		</div>

		<div class="space-y-2">
			<label for="password" class="block text-sm font-medium text-neutral-700 dark:text-neutral-300">Password</label>
			<div class="relative">
				<div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4"><Lock class="h-5 w-5 text-neutral-500" /></div>
				<input bind:value={form.password} required type={showPassword ? "text" : "password"} name="password" id="password" placeholder="••••••••" class="w-full rounded-2xl border border-neutral-300 bg-white py-3 pl-12 pr-12 text-neutral-900 placeholder-neutral-400 transition-colors focus:border-brand-400 focus:outline-none focus:ring-2 focus:ring-brand-400/20 dark:border-neutral-700/80 dark:bg-neutral-900 dark:text-white" />
				<button type="button" onclick={() => (showPassword = !showPassword)} class="absolute inset-y-0 right-0 flex items-center pr-4 text-neutral-400 transition-colors hover:text-neutral-700 dark:hover:text-neutral-300" aria-label="Tampilkan password">{#if showPassword}<EyeOff class="h-5 w-5" />{:else}<Eye class="h-5 w-5" />{/if}</button>
			</div>
		</div>

		<div class="flex justify-end"><a href="/forgot-password" use:inertia class="text-sm font-medium text-brand-600 transition-colors hover:text-brand-700 dark:text-brand-400">Lupa password?</a></div>

		<button type="submit" disabled={form.processing} class="flex w-full items-center justify-center gap-2 rounded-full bg-brand-600 px-4 py-3 font-semibold text-white transition hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-50">
			{form.processing ? "Memproses..." : "Masuk"}<ArrowRight class="h-5 w-5" />
		</button>
	</form>
</AuthLayout>
