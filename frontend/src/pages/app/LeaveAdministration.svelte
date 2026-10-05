<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { User } from "@lib/types";
	import { CalendarDays, Check, Save, X } from "lucide-svelte";

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
		member_id: number;
		member_name: string;
		division: string;
		start_date: string;
		end_date: string;
		chargeable_days: number;
		reason: string;
		status: string;
	}

	interface Props {
		user?: User;
		balances?: LeaveBalance[];
		pending?: LeaveRequest[];
		year: number;
		success?: string;
		error?: string;
	}

	let { user, balances = [], pending = [], year, success, error }: Props = $props();
	let sending = $state<string | null>(null);
	let drafts = $state<Record<number, number>>({});

	function entitlementValue(item: LeaveBalance) {
		return drafts[item.member_id] ?? item.entitled_days;
	}

	function saveEntitlement(item: LeaveBalance) {
		const key = `entitlement-${item.member_id}`;
		sending = key;
		router.post(`/app/leave/entitlements/${item.member_id}`, {
			year,
			entitled_days: entitlementValue(item),
			note: "Jatah cuti tahunan",
		}, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}

	function review(id: number, action: "approve" | "reject") {
		const key = `${action}-${id}`;
		sending = key;
		router.post(`/app/leave/reviews/${id}`, { action, review_note: "" }, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}
</script>

<AppLayout {user} group="leave">
	<div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="flex flex-col gap-4 rounded-[28px] bg-gradient-to-br from-brand-800 to-brand-600 p-6 text-white shadow-sm md:flex-row md:items-end md:justify-between sm:p-8">
			<div>
				<div class="flex items-center gap-2 text-sm font-bold text-secondary-200"><CalendarDays size="18" /> Cuti Tahunan</div>
				<h1 class="mt-2 text-3xl font-black tracking-tight">Administrasi cuti {year}</h1>
				<p class="mt-2 max-w-2xl text-sm leading-6 text-white/75">Atur jatah tiap anggota dan review pengajuan. Saldo hanya berkurang setelah approval.</p>
			</div>
			<div class="rounded-3xl bg-white/10 px-5 py-4 ring-1 ring-white/15"><div class="text-xs font-bold uppercase tracking-wider text-white/60">Menunggu approval</div><div class="mt-1 text-3xl font-black">{pending.length}</div></div>
		</section>

		{#if success}<div class="rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 ring-1 ring-emerald-200">{success}</div>{/if}
		{#if error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 ring-1 ring-red-200">{error}</div>{/if}

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 sm:p-6">
			<div><h2 class="text-lg font-black">Saldo cuti anggota</h2><p class="mt-1 text-sm text-neutral-500">Jatah bisa berbeda per anggota. Pending tidak mengurangi sisa tersedia.</p></div>
			<div class="mt-5 overflow-x-auto">
				<table class="w-full min-w-[820px] text-left text-sm">
					<thead><tr class="border-b border-neutral-200 text-xs uppercase tracking-wider text-neutral-500"><th class="px-3 py-3">Anggota</th><th class="px-3 py-3">Jatah</th><th class="px-3 py-3">Terpakai</th><th class="px-3 py-3">Pending</th><th class="px-3 py-3">Sisa</th><th class="px-3 py-3"></th></tr></thead>
					<tbody>
						{#each balances as item}
							<tr class="border-b border-neutral-100 last:border-0">
								<td class="px-3 py-4"><div class="font-black">{item.member_name}</div><div class="text-xs text-neutral-500">{item.division}</div></td>
								<td class="px-3 py-4"><input value={entitlementValue(item)} oninput={(e) => (drafts[item.member_id] = Number(e.currentTarget.value))} type="number" min="0" max="365" class="w-24 rounded-xl border border-neutral-200 px-3 py-2 text-sm font-bold outline-none focus:border-brand-500" /></td>
								<td class="px-3 py-4 font-bold">{item.approved_days}</td>
								<td class="px-3 py-4 font-bold text-amber-600">{item.pending_days}</td>
								<td class="px-3 py-4 text-lg font-black text-brand-700">{item.available_days}</td>
								<td class="px-3 py-4 text-right"><button type="button" onclick={() => saveEntitlement(item)} disabled={sending !== null} class="inline-flex items-center gap-2 rounded-full bg-brand-700 px-4 py-2 text-xs font-black text-white disabled:opacity-50"><Save size="15" />Simpan</button></td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 sm:p-6">
			<div><h2 class="text-lg font-black">Permintaan cuti</h2><p class="mt-1 text-sm text-neutral-500">Approval melakukan pengecekan saldo ulang agar tidak mungkin minus.</p></div>
			{#if pending.length === 0}
				<div class="mt-5 rounded-2xl bg-neutral-50 p-6 text-center text-sm text-neutral-500">Tidak ada pengajuan cuti yang menunggu.</div>
			{:else}
				<div class="mt-5 grid gap-3">
					{#each pending as item}
						<article class="rounded-2xl border border-neutral-200 p-4">
							<div class="flex flex-col gap-4 lg:flex-row lg:items-center">
								<div class="min-w-0 flex-1"><div class="text-sm font-black">{item.member_name} · {item.division}</div><div class="mt-1 text-xs text-neutral-500">{item.start_date} s/d {item.end_date} · {item.chargeable_days} hari kerja</div><div class="mt-2 text-sm text-neutral-700">{item.reason}</div></div>
								<div class="flex gap-2"><button type="button" onclick={() => review(item.id, "reject")} disabled={sending !== null} class="inline-flex items-center gap-2 rounded-full bg-red-50 px-4 py-2.5 text-xs font-black text-red-700"><X size="15" />Tolak</button><button type="button" onclick={() => review(item.id, "approve")} disabled={sending !== null} class="inline-flex items-center gap-2 rounded-full bg-brand-700 px-4 py-2.5 text-xs font-black text-white"><Check size="15" />Approve</button></div>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	</div>
</AppLayout>
