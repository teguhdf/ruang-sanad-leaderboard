<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { AttendanceAdjustmentRequest, OvertimeRequest, User } from "@lib/types";
	import { Check, Clock3, FileCheck2, TimerReset, X } from "lucide-svelte";

	interface Props {
		user?: User;
		adjustments?: AttendanceAdjustmentRequest[];
		overtime?: OvertimeRequest[];
		success?: string;
		error?: string;
	}

	let { user, adjustments = [], overtime = [], success, error }: Props = $props();
	let sending = $state<string | null>(null);

	function formatTime(value: string) {
		return new Intl.DateTimeFormat("id-ID", { hour: "2-digit", minute: "2-digit", timeZone: "Asia/Jakarta" }).format(new Date(value));
	}

	function formatMinutes(total: number) {
		const h = Math.floor(total / 60);
		const m = total % 60;
		return `${h}j ${m.toString().padStart(2, "0")}m`;
	}

	function review(kind: "adjustments" | "overtime", id: number, action: "approve" | "reject") {
		const key = `${kind}-${id}-${action}`;
		sending = key;
		router.post(`/app/attendance/reviews/${kind}/${id}`, { action, review_note: "" }, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}
</script>

<AppLayout {user} group="attendance-reviews">
	<div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="flex flex-col gap-4 rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] md:flex-row md:items-end md:justify-between">
			<div>
				<div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><FileCheck2 size="18" /> Approval Center</div>
				<h1 class="mt-2 text-3xl font-black tracking-tight">Review absensi & lembur</h1>
				<p class="mt-2 max-w-2xl text-sm leading-6 text-neutral-500">Employee tidak dapat mengubah jam sendiri. Koreksi dan lembur masuk ke inbox ini untuk keputusan SPV.</p>
			</div>
			<div class="rounded-2xl bg-brand-50 px-4 py-3 text-brand-800 dark:bg-brand-950/50 dark:text-brand-200"><div class="text-xs font-bold uppercase tracking-wider">Pending</div><div class="mt-1 text-2xl font-black">{adjustments.length + overtime.length}</div></div>
		</section>

		{#if success}<div class="rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 dark:bg-emerald-950/30 dark:text-emerald-300">{success}</div>{/if}
		{#if error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 dark:bg-red-950/30 dark:text-red-300">{error}</div>{/if}

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex items-center gap-2"><Clock3 size="20" /><h2 class="text-lg font-black">Koreksi absensi</h2></div>
			{#if adjustments.length === 0}
				<div class="mt-5 rounded-2xl bg-neutral-50 p-6 text-center text-sm text-neutral-500 dark:bg-white/[0.03]">Tidak ada koreksi yang menunggu approval.</div>
			{:else}
				<div class="mt-4 grid gap-3">
					{#each adjustments as item}
						<article class="rounded-2xl border border-neutral-200 p-4 dark:border-white/[0.07]">
							<div class="flex flex-col gap-4 lg:flex-row lg:items-center">
								<div class="min-w-0 flex-1"><div class="text-sm font-black">{item.member_name}</div><div class="mt-1 text-xs text-neutral-500">{item.work_date} · {formatTime(item.requested_check_in)} - {formatTime(item.requested_check_out)}</div><div class="mt-2 text-sm text-neutral-700 dark:text-neutral-300">{item.reason}</div></div>
								<div class="flex gap-2"><button type="button" onclick={() => review("adjustments", item.id, "reject")} disabled={sending !== null} class="inline-flex min-h-10 items-center gap-2 rounded-full bg-red-50 px-4 text-xs font-black text-red-700 dark:bg-red-950/30 dark:text-red-300"><X size="15" />Tolak</button><button type="button" onclick={() => review("adjustments", item.id, "approve")} disabled={sending !== null} class="inline-flex min-h-10 items-center gap-2 rounded-full bg-emerald-600 px-4 text-xs font-black text-white"><Check size="15" />Approve</button></div>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex items-center gap-2"><TimerReset size="20" /><h2 class="text-lg font-black">Permintaan lembur</h2></div>
			{#if overtime.length === 0}
				<div class="mt-5 rounded-2xl bg-neutral-50 p-6 text-center text-sm text-neutral-500 dark:bg-white/[0.03]">Tidak ada lembur yang menunggu approval.</div>
			{:else}
				<div class="mt-4 grid gap-3">
					{#each overtime as item}
						<article class="rounded-2xl border border-neutral-200 p-4 dark:border-white/[0.07]">
							<div class="flex flex-col gap-4 lg:flex-row lg:items-center">
								<div class="min-w-0 flex-1"><div class="text-sm font-black">Member #{item.member_id} · {item.work_date}</div><div class="mt-1 text-xs font-bold text-violet-600">{formatMinutes(item.requested_minutes)}</div><div class="mt-2 text-sm text-neutral-700 dark:text-neutral-300">{item.reason}</div></div>
								<div class="flex gap-2"><button type="button" onclick={() => review("overtime", item.id, "reject")} disabled={sending !== null} class="inline-flex min-h-10 items-center gap-2 rounded-full bg-red-50 px-4 text-xs font-black text-red-700 dark:bg-red-950/30 dark:text-red-300"><X size="15" />Tolak</button><button type="button" onclick={() => review("overtime", item.id, "approve")} disabled={sending !== null} class="inline-flex min-h-10 items-center gap-2 rounded-full bg-violet-700 px-4 text-xs font-black text-white"><Check size="15" />Approve</button></div>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	</div>
</AppLayout>
