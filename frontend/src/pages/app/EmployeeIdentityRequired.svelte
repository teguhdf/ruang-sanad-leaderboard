<script lang="ts">
	import { inertia } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { User } from "@lib/types";
	import { AlertCircle, ArrowLeft, UsersRound } from "lucide-svelte";

	interface Props {
		user?: User;
		group: string;
		title: string;
		description: string;
	}

	let { user, group, title, description }: Props = $props();
	const canManage = $derived(user?.role === "super_admin" || user?.role === "admin" || user?.role === "spv");
</script>

<AppLayout {user} {group}>
	<div class="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8 lg:py-12">
		<section class="rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-8">
			<div class="flex items-start gap-4">
				<div class="grid h-12 w-12 shrink-0 place-items-center rounded-2xl bg-amber-100 text-amber-700 dark:bg-amber-950/50 dark:text-amber-300">
					<AlertCircle size="22" />
				</div>
				<div class="min-w-0">
					<div class="text-xs font-black uppercase tracking-[0.14em] text-neutral-400">Employee access</div>
					<h1 class="mt-1 text-2xl font-black tracking-tight">{title} belum aktif</h1>
					<p class="mt-3 text-sm leading-6 text-neutral-500">{description}</p>
					<p class="mt-2 text-sm leading-6 text-neutral-500">
						{#if canManage}
							Hubungkan akun ini ke anggota yang tepat dari menu Attendance. Tidak ada data yang dihapus atau dipindahkan.
						{:else}
							Minta SPV menghubungkan akun Anda ke profil anggota yang benar.
						{/if}
					</p>

					<div class="mt-6 flex flex-wrap gap-3">
						{#if canManage}
							<a href="/app/attendance" use:inertia class="inline-flex min-h-11 items-center gap-2 rounded-full bg-brand-700 px-5 text-sm font-black text-white transition hover:bg-brand-800">
								<UsersRound size="17" /> Buka Attendance
							</a>
						{/if}
						<a href="/app" use:inertia class="inline-flex min-h-11 items-center gap-2 rounded-full border border-neutral-200 bg-white px-5 text-sm font-black text-neutral-700 transition hover:bg-neutral-50 dark:border-white/[0.08] dark:bg-neutral-900 dark:text-neutral-200">
							<ArrowLeft size="17" /> Kembali
						</a>
					</div>
				</div>
			</div>
		</section>
	</div>
</AppLayout>
