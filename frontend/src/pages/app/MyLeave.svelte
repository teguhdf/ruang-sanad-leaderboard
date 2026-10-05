<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { User } from "@lib/types";
	import { CalendarDays, CheckCircle2, Clock3, Send, XCircle } from "lucide-svelte";

	interface LeaveBalance {
		member_id: number;
		member_name: string;
		division: string;
		year: number;
		entitled_days: number;
		carryover_days: number;
		adjustment_days: number;
		approved_days: number;
		pending_days: number;
		available_days: number;
	}

	interface LeaveRequest {
		id: number;
		start_date: string;
		end_date: string;
		chargeable_days: number;
		reason: string;
		status: "pending" | "approved" | "rejected" | "cancelled";
		review_note: string;
		created_at: string;
	}

	interface Props {
		user?: User;
		balance: LeaveBalance;
		requests?: LeaveRequest[];
		year: number;
		success?: string;
		error?: string;
	}

	let { user, balance, requests = [], year, success, error }: Props = $props();
	let form = $state({ start_date: "", end_date: "", reason: "" });
	let sending = $state<string | null>(null);

	function submitLeave(event: SubmitEvent) {
		event.preventDefault();
		sending = "request";
		router.post("/app/my-leave/requests", form, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}

	function cancelRequest(id: number) {
		sending = `cancel-${id}`;
		router.post(`/app/my-leave/requests/${id}/cancel`, {}, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}

	function statusLabel(status: string) {
		if (status === "approved") return "Disetujui";
		if (status === "rejected") return "Ditolak";
		if (status === "cancelled") return "Dibatalkan";
		return "Menunggu";
	}
</script>

<AppLayout {user} group="my-leave">
	<div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-brand-800 to-brand-600 p-6 text-white shadow-sm sm:p-8">
			<div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
				<div>
					<div class="flex items-center gap-2 text-sm font-bold text-secondary-200"><CalendarDays size="18" /> Cuti Tahunan {year}</div>
					<h1 class="mt-2 text-3xl font-black tracking-tight sm:text-4xl">Kelola cuti tanpa kehilangan jejak saldo</h1>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-white/75">Saldo hanya berkurang setelah SPV menyetujui pengajuan. Hari yang bukan jadwal kerja tidak ikut dipotong.</p>
				</div>
				<div class="rounded-3xl bg-white/10 px-6 py-4 ring-1 ring-white/15">
					<div class="text-xs font-bold uppercase tracking-wider text-white/60">Sisa tersedia</div>
					<div class="mt-1 text-4xl font-black">{balance.available_days}</div>
					<div class="text-xs text-white/70">hari cuti</div>
				</div>
			</div>
		</section>

		{#if success}<div class="rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 ring-1 ring-emerald-200">{success}</div>{/if}
		{#if error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 ring-1 ring-red-200">{error}</div>{/if}

		<section class="grid gap-4 sm:grid-cols-3">
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200"><div class="text-xs font-bold uppercase tracking-wider text-neutral-500">Jatah tahun ini</div><div class="mt-3 text-3xl font-black">{balance.entitled_days + balance.carryover_days + balance.adjustment_days}</div><div class="mt-1 text-xs text-neutral-500">hari</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200"><div class="text-xs font-bold uppercase tracking-wider text-neutral-500">Sudah dipakai</div><div class="mt-3 text-3xl font-black text-brand-700">{balance.approved_days}</div><div class="mt-1 text-xs text-neutral-500">approved</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200"><div class="text-xs font-bold uppercase tracking-wider text-neutral-500">Menunggu</div><div class="mt-3 text-3xl font-black text-amber-600">{balance.pending_days}</div><div class="mt-1 text-xs text-neutral-500">belum memotong saldo</div></article>
		</section>

		<section class="grid gap-6 lg:grid-cols-[.9fr_1.1fr]">
			<form onsubmit={submitLeave} class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 sm:p-6">
				<div class="flex items-center gap-2"><Send size="20" class="text-brand-700" /><h2 class="text-lg font-black">Ajukan cuti</h2></div>
				<p class="mt-2 text-sm text-neutral-500">Cuti v1 menggunakan full-day. Saldo dicek ulang saat approval.</p>
				<div class="mt-5 grid gap-4 sm:grid-cols-2">
					<label class="grid gap-1.5 text-xs font-bold text-neutral-600">Tanggal mulai<input bind:value={form.start_date} type="date" required class="rounded-2xl border border-neutral-200 px-4 py-3 text-sm outline-none focus:border-brand-500" /></label>
					<label class="grid gap-1.5 text-xs font-bold text-neutral-600">Tanggal selesai<input bind:value={form.end_date} type="date" required class="rounded-2xl border border-neutral-200 px-4 py-3 text-sm outline-none focus:border-brand-500" /></label>
				</div>
				<label class="mt-4 grid gap-1.5 text-xs font-bold text-neutral-600">Alasan<textarea bind:value={form.reason} required minlength="5" rows="4" placeholder="Contoh: keperluan keluarga" class="rounded-2xl border border-neutral-200 px-4 py-3 text-sm outline-none focus:border-brand-500"></textarea></label>
				<button type="submit" disabled={sending !== null} class="mt-5 inline-flex min-h-12 w-full items-center justify-center gap-2 rounded-full bg-brand-700 px-5 text-sm font-black text-white transition hover:bg-brand-800 disabled:opacity-50"><Send size="17" />{sending === "request" ? "Mengirim..." : "Kirim pengajuan"}</button>
			</form>

			<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 sm:p-6">
				<div class="flex items-center gap-2"><Clock3 size="20" /><h2 class="text-lg font-black">Riwayat pengajuan</h2></div>
				{#if requests.length === 0}
					<div class="mt-5 rounded-2xl bg-neutral-50 p-6 text-center text-sm text-neutral-500">Belum ada pengajuan cuti.</div>
				{:else}
					<div class="mt-4 grid gap-3">
						{#each requests as item}
							<article class="rounded-2xl border border-neutral-200 p-4">
								<div class="flex items-start gap-3">
									<div class="grid h-10 w-10 place-items-center rounded-2xl {item.status === 'approved' ? 'bg-emerald-50 text-emerald-700' : item.status === 'rejected' ? 'bg-red-50 text-red-700' : 'bg-amber-50 text-amber-700'}">{#if item.status === "approved"}<CheckCircle2 size="18" />{:else if item.status === "rejected"}<XCircle size="18" />{:else}<Clock3 size="18" />{/if}</div>
									<div class="min-w-0 flex-1"><div class="text-sm font-black">{item.start_date} s/d {item.end_date}</div><div class="mt-1 text-xs text-neutral-500">{item.chargeable_days} hari kerja · {statusLabel(item.status)}</div><div class="mt-2 text-sm text-neutral-700">{item.reason}</div>{#if item.review_note}<div class="mt-2 text-xs text-neutral-500">Catatan SPV: {item.review_note}</div>{/if}</div>
									{#if item.status === "pending"}<button type="button" onclick={() => cancelRequest(item.id)} disabled={sending !== null} class="rounded-full bg-red-50 px-3 py-2 text-xs font-black text-red-700">Batalkan</button>{/if}
								</div>
							</article>
						{/each}
					</div>
				{/if}
			</section>
		</section>
	</div>
</AppLayout>
