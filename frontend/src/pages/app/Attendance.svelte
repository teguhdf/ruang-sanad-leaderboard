<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { AttendanceFoundation, User } from "@lib/types";
	import {
		BadgeCheck,
		Building2,
		CalendarClock,
		ChevronRight,
		Clock3,
		Coffee,
		Laptop,
		MailPlus,
		Save,
		ShieldCheck,
		UsersRound,
	} from "lucide-svelte";

	interface Props {
		user?: User;
		foundation: AttendanceFoundation;
		success?: string;
		error?: string;
	}

	let { user, foundation, success, error }: Props = $props();
	let drafts = $state<Record<number, { work_mode: string; schedule_type: string }>>(
		Object.fromEntries(foundation.profiles.map((p) => [p.member_id, { work_mode: p.work_mode, schedule_type: p.schedule_type }])),
	);
	let inviteEmails = $state<Record<number, string>>({});
	let saving = $state<string | null>(null);

	function formatMinutes(total: number) {
		const h = Math.floor(total / 60);
		const m = total % 60;
		return m ? `${h}j ${m}m` : `${h} jam`;
	}

	function saveProfile(memberID: number) {
		saving = `profile-${memberID}`;
		router.put(`/app/attendance/profiles/${memberID}`, drafts[memberID], {
			preserveScroll: true,
			onFinish: () => (saving = null),
		});
	}

	function invite(memberID: number) {
		const email = (inviteEmails[memberID] ?? "").trim();
		if (!email) return;
		saving = `invite-${memberID}`;
		router.post(`/app/attendance/invite/${memberID}`, { email }, {
			preserveScroll: true,
			onFinish: () => (saving = null),
		});
	}
</script>

<AppLayout {user} group="attendance">
	<div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="flex flex-col gap-5 rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] lg:flex-row lg:items-end lg:justify-between">
			<div>
				<div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><BadgeCheck size="18" /> Attendance Foundation</div>
				<h1 class="mt-2 text-3xl font-black tracking-tight">Absensi yang mengikuti pola kerja tim</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-neutral-500">Shift hybrid, remote fleksibel, Sabtu setengah hari, carry-over mingguan, akun employee, koreksi karena lupa absensi, dan lembur dengan approval SPV.</p>
			</div>
			<a href="/app/attendance/reviews" class="inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-brand-600 px-5 text-sm font-black text-white transition hover:bg-brand-700">Buka Approval Center <ChevronRight size="17" /></a>
		</section>

		{#if success}<div class="rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 dark:bg-emerald-950/30 dark:text-emerald-300">{success}</div>{/if}
		{#if error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 dark:bg-red-950/30 dark:text-red-300">{error}</div>{/if}

		<section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><Clock3 size="18" />Senin - Jumat</div><div class="mt-3 text-2xl font-black">{formatMinutes(foundation.policy.weekday_target_minutes)}</div><div class="mt-1 text-xs text-neutral-500">target harian normal</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><Coffee size="18" />Sabtu</div><div class="mt-3 text-2xl font-black">{formatMinutes(foundation.policy.saturday_target_minutes)}</div><div class="mt-1 text-xs text-neutral-500">setengah hari</div></article>
			<article class="rounded-3xl bg-brand-50 p-5 ring-1 ring-brand-200 dark:bg-brand-950/35 dark:ring-brand-900/60"><div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><CalendarClock size="18" />Target mingguan</div><div class="mt-3 text-2xl font-black text-brand-900 dark:text-brand-100">{formatMinutes(foundation.policy.weekly_target_minutes)}</div><div class="mt-1 text-xs text-brand-700/70 dark:text-brand-300/70">carry-over dalam minggu yang sama</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><ShieldCheck size="18" />Lembur</div><div class="mt-3 text-xl font-black">Approval SPV</div><div class="mt-1 text-xs text-neutral-500">terpisah dari jam pengganti</div></article>
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<h2 class="text-lg font-black">Template shift</h2>
			<p class="mt-1 text-sm text-neutral-500">Pola awal yang sudah dikunci untuk operasional.</p>
			<div class="mt-5 grid gap-4 lg:grid-cols-2">
				{#each foundation.shift_templates as shift}
					<article class="rounded-3xl bg-neutral-50 p-5 dark:bg-white/[0.03]">
						<div class="flex items-start justify-between gap-4"><div><div class="text-base font-black">{shift.name}</div><div class="mt-1 text-xs text-neutral-500">{shift.description}</div></div><span class="rounded-full bg-emerald-100 px-3 py-1 text-[10px] font-black uppercase text-emerald-700">{formatMinutes(shift.required_minutes)}</span></div>
						<div class="mt-4 grid gap-2">
							{#each shift.segments as segment}
								<div class="flex items-center gap-3 rounded-2xl bg-white px-4 py-3 ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.06]"><div class="grid h-9 w-9 place-items-center rounded-xl bg-brand-50 text-brand-700 dark:bg-brand-950/50 dark:text-brand-300">{#if segment.location_mode === "office"}<Building2 size="17" />{:else}<Laptop size="17" />{/if}</div><div class="min-w-0 flex-1"><div class="text-xs font-black">{segment.location_mode === "office" ? "Kantor" : "Remote"}</div><div class="mt-0.5 text-[11px] text-neutral-500">{segment.is_flexible ? `Fleksibel · ${formatMinutes(segment.target_minutes)}` : `${segment.start_time} - ${segment.end_time} · ${formatMinutes(segment.target_minutes)}`}</div></div></div>
							{/each}
						</div>
					</article>
				{/each}
			</div>
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex items-end justify-between gap-4"><div><div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><UsersRound size="18" /> Employee access</div><h2 class="mt-1 text-xl font-black">Profil absensi & akun employee</h2><p class="mt-1 text-sm text-neutral-500">Akun employee hanya dibuat dari sini. Tidak ada public register.</p></div><div class="rounded-full bg-neutral-100 px-3 py-1.5 text-xs font-black dark:bg-white/[0.05]">{foundation.profiles.length} anggota</div></div>

			<div class="mt-5 grid gap-4">
				{#each foundation.profiles as profile}
					<article class="grid gap-4 rounded-3xl border border-neutral-200 p-4 dark:border-white/[0.07] xl:grid-cols-[minmax(180px,1fr)_170px_180px_minmax(240px,1.2fr)] xl:items-end">
						<div class="flex items-center gap-3"><div class="grid h-11 w-11 place-items-center rounded-2xl bg-brand-100 text-sm font-black text-brand-800 dark:bg-brand-950/60 dark:text-brand-200">{profile.member_name.charAt(0).toUpperCase()}</div><div class="min-w-0"><div class="truncate text-sm font-black">{profile.member_name}</div><div class="mt-0.5 truncate text-xs text-neutral-500">{profile.division}</div></div></div>
						<label class="grid gap-1.5 text-xs font-bold text-neutral-500">Mode kerja<select bind:value={drafts[profile.member_id].work_mode} class="rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2.5 text-sm font-bold text-neutral-800 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100"><option value="onsite">Onsite</option><option value="remote">Remote</option><option value="hybrid">Hybrid</option></select></label>
						<div class="grid gap-2"><label class="grid gap-1.5 text-xs font-bold text-neutral-500">Tipe jadwal<select bind:value={drafts[profile.member_id].schedule_type} class="rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2.5 text-sm font-bold text-neutral-800 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100"><option value="shift">Shift / roster</option><option value="flexible">Fleksibel</option></select></label><button type="button" onclick={() => saveProfile(profile.member_id)} disabled={saving !== null} class="inline-flex min-h-10 items-center justify-center gap-2 rounded-full bg-neutral-900 px-4 text-xs font-black text-white dark:bg-white dark:text-neutral-900"><Save size="15" />Simpan profil</button></div>
						{#if profile.linked_user_id}
							<div class="rounded-2xl bg-emerald-50 px-4 py-3 dark:bg-emerald-950/25"><div class="text-xs font-black text-emerald-700 dark:text-emerald-300">Akun employee terhubung</div><div class="mt-1 text-[11px] text-emerald-700/70 dark:text-emerald-300/70">ID user #{profile.linked_user_id}. Invite dapat dikirim ulang dengan email yang sama melalui endpoint jika diperlukan.</div></div>
						{:else}
							<div class="grid gap-2"><label class="grid gap-1.5 text-xs font-bold text-neutral-500">Email employee<input bind:value={inviteEmails[profile.member_id]} type="email" placeholder="nama@ruangsanad.com" class="rounded-2xl border border-neutral-200 bg-neutral-50 px-3 py-2.5 text-sm font-medium dark:border-white/[0.07] dark:bg-neutral-900" /></label><button type="button" onclick={() => invite(profile.member_id)} disabled={saving !== null || !(inviteEmails[profile.member_id] ?? '').trim()} class="inline-flex min-h-10 items-center justify-center gap-2 rounded-full bg-brand-600 px-4 text-xs font-black text-white disabled:opacity-40"><MailPlus size="15" />Undang employee</button></div>
						{/if}
					</article>
				{/each}
			</div>
		</section>
	</div>
</AppLayout>
