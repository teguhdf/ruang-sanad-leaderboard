<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { AttendanceFoundation, User } from "@lib/types";
	import {
		BadgeCheck,
		Building2,
		CalendarClock,
		CalendarDays,
		ChevronLeft,
		ChevronRight,
		Clock3,
		Coffee,
		Laptop,
		MailPlus,
		RotateCcw,
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

	type RosterDraft = {
		day_type: string;
		work_mode: string;
		shift_template_id: string;
		required_minutes: number;
	};

	let { user, foundation, success, error }: Props = $props();
	let drafts = $state<Record<number, { work_mode: string; schedule_type: string }>>(
		Object.fromEntries(foundation.profiles.map((p) => [p.member_id, { work_mode: p.work_mode, schedule_type: p.schedule_type }])),
	);
	let inviteEmails = $state<Record<number, string>>({});
	let saving = $state<string | null>(null);
	let selectedMemberID = $state<number>(foundation.roster_weeks[0]?.member_id ?? 0);
	let weekInput = $state(foundation.roster_week_start);
	let rosterDrafts = $state<Record<number, Record<string, RosterDraft>>>(
		Object.fromEntries(
			foundation.roster_weeks.map((week) => [
				week.member_id,
				Object.fromEntries(
					week.plan.days.map((day) => [
						day.work_date,
						{
							day_type: day.day_type,
							work_mode: day.work_mode,
							shift_template_id: day.shift_template_id ? String(day.shift_template_id) : "",
							required_minutes: day.required_minutes,
						},
					]),
				),
			]),
		),
	);

	const selectedWeek = $derived(foundation.roster_weeks.find((week) => week.member_id === selectedMemberID) ?? foundation.roster_weeks[0]);

	function formatMinutes(total: number) {
		const h = Math.floor(total / 60);
		const m = total % 60;
		return m ? `${h}j ${m}m` : `${h} jam`;
	}

	function formatDate(date: string) {
		return new Intl.DateTimeFormat("id-ID", { weekday: "short", day: "numeric", month: "short" }).format(new Date(`${date}T00:00:00`));
	}

	function shiftName(id?: number) {
		return foundation.shift_templates.find((shift) => shift.id === id)?.name ?? "Tanpa shift";
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

	function showWeek(value = weekInput) {
		if (!value) return;
		router.get("/app/attendance", { week_start: value }, { preserveState: false, preserveScroll: true });
	}

	function moveWeek(days: number) {
		const date = new Date(`${foundation.roster_week_start}T00:00:00`);
		date.setDate(date.getDate() + days);
		showWeek(date.toISOString().slice(0, 10));
	}

	function saveRoster(memberID: number, workDate: string) {
		const draft = rosterDrafts[memberID]?.[workDate];
		if (!draft) return;
		saving = `roster-${memberID}-${workDate}`;
		const isWork = draft.day_type === "work";
		router.post(
			`/app/attendance/roster/${memberID}/${workDate}?week_start=${foundation.roster_week_start}`,
			{
				day_type: draft.day_type,
				work_mode: draft.work_mode,
				shift_template_id: isWork && draft.shift_template_id ? Number(draft.shift_template_id) : null,
				required_minutes: isWork ? Number(draft.required_minutes) : 0,
			},
			{ preserveScroll: true, onFinish: () => (saving = null) },
		);
	}

	function resetRoster(memberID: number, workDate: string) {
		saving = `reset-${memberID}-${workDate}`;
		router.delete(`/app/attendance/roster/${memberID}/${workDate}?week_start=${foundation.roster_week_start}`, {
			preserveScroll: true,
			onFinish: () => (saving = null),
		});
	}
</script>

<AppLayout {user} group="attendance">
	<div class="mx-auto min-w-0 max-w-7xl space-y-6 overflow-x-hidden px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="flex flex-col gap-5 rounded-[28px] bg-white p-6 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] lg:flex-row lg:items-end lg:justify-between">
			<div>
				<div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><BadgeCheck size="18" /> Attendance Foundation</div>
				<h1 class="mt-2 text-3xl font-black tracking-tight">Absensi yang mengikuti pola kerja tim</h1>
				<p class="mt-2 max-w-3xl text-sm leading-6 text-neutral-500">Roster menjadi sumber kewajiban kerja. Cuti approved, hari pengganti, mode kerja dan target jam dihitung sebelum evidence absensi dinilai.</p>
			</div>
			<a href="/app/attendance/reviews" class="inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-brand-600 px-5 text-sm font-black text-white transition hover:bg-brand-700">Buka Approval Center <ChevronRight size="17" /></a>
		</section>

		{#if success}<div class="rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 dark:bg-emerald-950/30 dark:text-emerald-300">{success}</div>{/if}
		{#if error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 dark:bg-red-950/30 dark:text-red-300">{error}</div>{/if}

		<section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><Clock3 size="18" />Senin - Jumat</div><div class="mt-3 text-2xl font-black">{formatMinutes(foundation.policy.weekday_target_minutes)}</div><div class="mt-1 text-xs text-neutral-500">target harian normal</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><Coffee size="18" />Sabtu</div><div class="mt-3 text-2xl font-black">{formatMinutes(foundation.policy.saturday_target_minutes)}</div><div class="mt-1 text-xs text-neutral-500">setengah hari</div></article>
			<article class="rounded-3xl bg-brand-50 p-5 ring-1 ring-brand-200 dark:bg-brand-950/35 dark:ring-brand-900/60"><div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><CalendarClock size="18" />Target normal</div><div class="mt-3 text-2xl font-black text-brand-900 dark:text-brand-100">{formatMinutes(foundation.policy.weekly_target_minutes)}</div><div class="mt-1 text-xs text-brand-700/70 dark:text-brand-300/70">target efektif mengikuti roster & cuti</div></article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center gap-2 text-sm font-bold text-neutral-500"><ShieldCheck size="18" />Lembur</div><div class="mt-3 text-xl font-black">Approval SPV</div><div class="mt-1 text-xs text-neutral-500">terpisah dari jam pengganti</div></article>
		</section>

		<section class="min-w-0 overflow-hidden rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex flex-col gap-5 border-b border-neutral-100 pb-5 dark:border-white/[0.06] 2xl:flex-row 2xl:items-center 2xl:justify-between">
				<div class="min-w-0">
					<div class="flex items-center gap-2 text-sm font-bold text-brand-700 dark:text-brand-300"><CalendarDays size="18" /> Roster mingguan</div>
					<h2 class="mt-1 text-xl font-black">Atur kewajiban kerja per anggota</h2>
					<p class="mt-1 max-w-3xl text-sm leading-6 text-neutral-500">Jadwal default mengikuti hari kerja anggota. Override hanya hari yang berubah. Cuti approved otomatis dikunci.</p>
				</div>
				<div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
					<div class="flex items-center gap-1.5 rounded-2xl bg-neutral-50 p-1.5 ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.07]">
						<button type="button" onclick={() => moveWeek(-7)} aria-label="Minggu sebelumnya" class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-neutral-600 shadow-sm ring-1 ring-neutral-200 transition hover:text-neutral-950 dark:bg-neutral-800 dark:text-neutral-300 dark:ring-white/[0.07]"><ChevronLeft size="17" /></button>
						<label class="flex min-w-0 items-center gap-2 px-1 text-xs font-bold text-neutral-500"><span class="hidden sm:inline">Minggu</span><input bind:value={weekInput} type="date" class="h-9 w-[148px] border-0 bg-transparent p-0 text-sm font-black text-neutral-800 outline-none focus:ring-0 dark:text-neutral-100" /></label>
						<button type="button" onclick={() => moveWeek(7)} aria-label="Minggu berikutnya" class="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-white text-neutral-600 shadow-sm ring-1 ring-neutral-200 transition hover:text-neutral-950 dark:bg-neutral-800 dark:text-neutral-300 dark:ring-white/[0.07]"><ChevronRight size="17" /></button>
					</div>
					<button type="button" onclick={() => showWeek()} class="min-h-12 rounded-2xl bg-neutral-900 px-5 text-xs font-black text-white transition hover:bg-neutral-800 dark:bg-white dark:text-neutral-900">Tampilkan</button>
				</div>
			</div>

			<div class="mt-5 grid min-w-0 gap-5 2xl:grid-cols-[220px_minmax(0,1fr)]">
				<aside class="min-w-0 2xl:sticky 2xl:top-6 2xl:self-start">
					<div class="rounded-3xl bg-neutral-50 p-2.5 ring-1 ring-neutral-100 dark:bg-white/[0.03] dark:ring-white/[0.05]">
						<div class="px-2 pb-2 pt-1 text-[10px] font-black uppercase tracking-[0.14em] text-neutral-400">Pilih anggota</div>
						<div class="grid gap-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-1">
							{#each foundation.roster_weeks as week}
								<button type="button" onclick={() => (selectedMemberID = week.member_id)} class={`group flex items-center justify-between gap-3 rounded-2xl px-3 py-2.5 text-left transition ${selectedMemberID === week.member_id ? 'bg-brand-600 text-white shadow-sm' : 'text-neutral-800 hover:bg-white hover:shadow-sm dark:text-neutral-100 dark:hover:bg-white/[0.05]'}`}>
									<div class="min-w-0"><div class="truncate text-sm font-black">{week.member_name}</div><div class={`mt-0.5 truncate text-[11px] ${selectedMemberID === week.member_id ? 'text-white/70' : 'text-neutral-500'}`}>{week.division}</div></div>
									<div class={`shrink-0 text-[11px] font-black ${selectedMemberID === week.member_id ? 'text-white' : 'text-neutral-500'}`}>{formatMinutes(week.plan.required_minutes)}</div>
								</button>
							{/each}
						</div>
					</div>
				</aside>

				{#if selectedWeek}
					<div class="min-w-0">
						<div class="flex flex-col gap-3 rounded-2xl bg-neutral-50 px-4 py-3 dark:bg-white/[0.03] sm:flex-row sm:items-center sm:justify-between">
							<div class="min-w-0"><div class="truncate text-base font-black">{selectedWeek.member_name}</div><div class="mt-0.5 text-xs text-neutral-500">{foundation.roster_week_start} s/d {foundation.roster_week_end}</div></div>
							<div class="flex items-center gap-2"><span class="text-xs font-bold text-neutral-500">Target efektif</span><span class="rounded-xl bg-brand-100 px-3 py-1.5 text-sm font-black text-brand-800 dark:bg-brand-950/50 dark:text-brand-200">{formatMinutes(selectedWeek.plan.required_minutes)}</span></div>
						</div>

						<div class="mt-3 overflow-x-auto rounded-3xl border border-neutral-200 dark:border-white/[0.07]">
							<div class="hidden bg-neutral-50 px-4 py-2.5 text-[10px] font-black uppercase tracking-[0.12em] text-neutral-400 dark:bg-white/[0.03] lg:grid lg:min-w-[900px] lg:grid-cols-[110px_135px_115px_minmax(180px,1fr)_90px_210px] lg:gap-2">
								<div>Hari</div><div>Status</div><div>Mode</div><div>Shift</div><div>Target</div><div class="text-right">Aksi</div>
							</div>

							{#each selectedWeek.plan.days as day}
								{@const draft = rosterDrafts[selectedWeek.member_id]?.[day.work_date]}
								<div class={`border-t border-neutral-100 px-4 py-3 first:border-t-0 dark:border-white/[0.06] ${day.source === 'approved_leave' ? 'bg-amber-50/65 dark:bg-amber-950/10' : 'bg-white dark:bg-neutral-925'}`}>
									<div class="grid gap-3 sm:grid-cols-2 lg:min-w-[900px] lg:grid-cols-[110px_135px_115px_minmax(180px,1fr)_90px_210px] lg:items-center lg:gap-2">
										<div class="sm:col-span-2 lg:col-span-1">
											<div class="text-sm font-black">{formatDate(day.work_date)}</div>
											<div class={`mt-1 inline-flex rounded-lg px-2 py-1 text-[9px] font-black uppercase tracking-wide ${day.source === 'approved_leave' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950/50 dark:text-amber-200' : day.source === 'roster' ? 'bg-brand-50 text-brand-700 dark:bg-brand-950/50 dark:text-brand-300' : 'bg-neutral-100 text-neutral-500 dark:bg-white/[0.05] dark:text-neutral-400'}`}>{day.source === 'approved_leave' ? 'Cuti approved' : day.source === 'roster' ? 'Override' : 'Default'}</div>
										</div>

										<label class="grid gap-1 text-[10px] font-bold text-neutral-500"><span class="lg:hidden">Status</span><select bind:value={draft.day_type} disabled={day.source === 'approved_leave'} class="h-10 rounded-xl border border-neutral-200 bg-neutral-50 px-2.5 text-xs font-bold text-neutral-800 disabled:cursor-not-allowed disabled:opacity-60 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100"><option value="work">Kerja</option><option value="off">Off</option><option value="replacement_off">Libur pengganti</option><option value="holiday">Libur</option><option value="sick">Sakit</option>{#if day.source === 'approved_leave'}<option value="leave">Cuti</option>{/if}</select></label>
										<label class="grid gap-1 text-[10px] font-bold text-neutral-500"><span class="lg:hidden">Mode</span><select bind:value={draft.work_mode} disabled={day.source === 'approved_leave' || draft.day_type !== 'work'} class="h-10 rounded-xl border border-neutral-200 bg-neutral-50 px-2.5 text-xs font-bold text-neutral-800 disabled:cursor-not-allowed disabled:opacity-45 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100"><option value="onsite">Onsite</option><option value="remote">Remote</option><option value="hybrid">Hybrid</option></select></label>
										<label class="grid gap-1 text-[10px] font-bold text-neutral-500"><span class="lg:hidden">Shift</span><select bind:value={draft.shift_template_id} disabled={day.source === 'approved_leave' || draft.day_type !== 'work'} class="h-10 min-w-0 rounded-xl border border-neutral-200 bg-neutral-50 px-2.5 text-xs font-bold text-neutral-800 disabled:cursor-not-allowed disabled:opacity-45 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100"><option value="">Tanpa shift</option>{#each foundation.shift_templates as shift}<option value={String(shift.id)}>{shift.name}</option>{/each}</select></label>
										<label class="grid gap-1 text-[10px] font-bold text-neutral-500"><span class="lg:hidden">Target menit</span><input bind:value={draft.required_minutes} disabled={day.source === 'approved_leave' || draft.day_type !== 'work'} type="number" min="1" max="960" class="h-10 w-full min-w-0 rounded-xl border border-neutral-200 bg-neutral-50 px-2.5 text-xs font-black text-neutral-800 disabled:cursor-not-allowed disabled:opacity-45 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-neutral-100" /></label>

										<div class="flex min-w-0 items-center justify-end gap-2 sm:col-span-2 lg:col-span-1">
											{#if day.source !== 'approved_leave'}
												<button type="button" onclick={() => saveRoster(selectedWeek.member_id, day.work_date)} disabled={saving !== null} class="inline-flex h-10 shrink-0 items-center justify-center gap-1.5 rounded-xl bg-brand-600 px-3.5 text-[11px] font-black text-white transition hover:bg-brand-700 disabled:opacity-50"><Save size="14" />Simpan</button>
												{#if day.source === 'roster'}<button type="button" onclick={() => resetRoster(selectedWeek.member_id, day.work_date)} disabled={saving !== null} class="inline-flex h-10 shrink-0 items-center justify-center gap-1.5 rounded-xl border border-neutral-200 bg-white px-3 text-[11px] font-black text-neutral-600 transition hover:bg-neutral-50 disabled:opacity-50 dark:border-white/[0.08] dark:bg-neutral-900 dark:text-neutral-300"><RotateCcw size="14" />Default</button>{/if}
											{:else}
												<span class="rounded-xl bg-amber-100 px-3 py-2 text-[10px] font-black text-amber-800 dark:bg-amber-950/40 dark:text-amber-200">Dikelola dari Cuti</span>
											{/if}
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
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
					<article class="grid gap-4 rounded-3xl border border-neutral-200 p-4 dark:border-white/[0.07] lg:grid-cols-2 2xl:grid-cols-[minmax(180px,1fr)_170px_180px_minmax(240px,1.2fr)] 2xl:items-end">
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
