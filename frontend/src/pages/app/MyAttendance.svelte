<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { MyAttendanceSummary, User } from "@lib/types";
	import {
		AlertCircle,
		BriefcaseBusiness,
		Building2,
		Clock3,
		History,
		Laptop,
		Play,
		PlusCircle,
		Send,
		Square,
		TimerReset,
	} from "lucide-svelte";

	interface Props {
		user?: User;
		attendance: MyAttendanceSummary;
		success?: string;
		error?: string;
	}

	let { user, attendance, success, error }: Props = $props();
	let locationMode = $state<"office" | "remote">(
		attendance.today_plan.work_mode === "remote" ? "remote" : "office",
	);
	const canStartToday = $derived(attendance.today_plan.day_type === "work" && attendance.today_plan.required_minutes > 0);
	const allowsOffice = $derived(attendance.today_plan.work_mode !== "remote");
	const allowsRemote = $derived(attendance.today_plan.work_mode !== "onsite");
	let nowMs = $state(Date.now());
	let forgotCheckout = $state({ end_time: "", reason: "" });
	let correction = $state({
		work_date: attendance.today,
		location_mode: attendance.identity.work_mode === "remote" ? "remote" : "office",
		start_time: "",
		end_time: "",
		reason: "",
	});
	let overtime = $state({ work_date: attendance.today, requested_minutes: 60, reason: "" });
	let sending = $state<string | null>(null);

	$effect(() => {
		const timer = setInterval(() => (nowMs = Date.now()), 1000);
		return () => clearInterval(timer);
	});

	function formatMinutes(total: number) {
		const sign = total < 0 ? "-" : total > 0 ? "+" : "";
		const value = Math.abs(total);
		const h = Math.floor(value / 60);
		const m = value % 60;
		return `${sign}${h}j ${m.toString().padStart(2, "0")}m`;
	}

	function formatTime(value?: string) {
		if (!value) return "-";
		return new Intl.DateTimeFormat("id-ID", {
			hour: "2-digit",
			minute: "2-digit",
			timeZone: "Asia/Jakarta",
		}).format(new Date(value));
	}

	function dayTypeLabel(value: string) {
		switch (value) {
		case "work": return "Hari kerja";
		case "replacement_off": return "Libur pengganti";
		case "holiday": return "Libur";
		case "sick": return "Sakit";
		case "leave": return "Cuti";
		default: return "Off";
		}
	}

	function activeElapsed() {
		if (!attendance.active_session) return "0j 00m";
		const start = new Date(attendance.active_session.check_in_at).getTime();
		return formatMinutes(Math.max(0, Math.floor((nowMs - start) / 60000))).replace("+", "");
	}

	function post(path: string, data: Record<string, unknown>, key: string) {
		sending = key;
		router.post(path, data, {
			preserveScroll: true,
			onFinish: () => (sending = null),
		});
	}

	function startWork() {
		post("/app/my-attendance/start", { location_mode: locationMode }, "start");
	}

	function stopWork() {
		post("/app/my-attendance/stop", {}, "stop");
	}

	function submitForgotCheckout(event: SubmitEvent) {
		event.preventDefault();
		post("/app/my-attendance/forgot-checkout", forgotCheckout, "forgot-checkout");
	}

	function submitCorrection(event: SubmitEvent) {
		event.preventDefault();
		post("/app/my-attendance/adjustments", correction, "correction");
	}

	function submitOvertime(event: SubmitEvent) {
		event.preventDefault();
		post("/app/my-attendance/overtime", overtime, "overtime");
	}
</script>

<AppLayout {user} group="my-attendance">
	<div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="overflow-hidden rounded-[28px] bg-gradient-to-br from-brand-700 to-brand-500 p-6 text-white shadow-sm sm:p-8">
			<div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
				<div>
					<div class="mb-2 text-sm font-semibold text-white/80">Absensi Saya · {attendance.identity.division}</div>
					<h1 class="text-3xl font-black tracking-tight sm:text-4xl">Halo, {attendance.identity.member_name}</h1>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-white/80">
						Mulai dan selesaikan setiap sesi kerja dari sini. Waktu absensi selalu memakai jam server dan tidak bisa diedit langsung.
					</p>
				</div>
				<div class="rounded-3xl bg-white/12 px-5 py-4 backdrop-blur-sm ring-1 ring-white/20">
					<div class="text-xs font-semibold uppercase tracking-wider text-white/70">Minggu ini</div>
					<div class="mt-1 text-2xl font-black">{formatMinutes(attendance.weekly_minutes).replace("+", "")}</div>
					<div class="text-xs text-white/75">target {formatMinutes(attendance.weekly_target_minutes).replace("+", "")}</div>
				</div>
			</div>
		</section>

		{#if success}
			<div class="rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm font-semibold text-emerald-800 dark:border-emerald-900/50 dark:bg-emerald-950/30 dark:text-emerald-300">{success}</div>
		{/if}
		{#if error}
			<div class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-semibold text-red-800 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300">{error}</div>
		{/if}

		<section class="grid gap-4 md:grid-cols-3">
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]">
				<div class="flex items-center gap-3 text-neutral-500"><Clock3 size="20" /><span class="text-sm font-semibold">Hari ini</span></div>
				<div class="mt-4 text-3xl font-black">{formatMinutes(attendance.today_minutes).replace("+", "")}</div>
				<div class="mt-1 text-xs text-neutral-500">{dayTypeLabel(attendance.today_plan.day_type)} · target {formatMinutes(attendance.today_plan.required_minutes).replace("+", "")}</div>
			</article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]">
				<div class="flex items-center gap-3 text-neutral-500"><TimerReset size="20" /><span class="text-sm font-semibold">Balance minggu</span></div>
				<div class="mt-4 text-3xl font-black {attendance.weekly_balance_minutes < 0 ? 'text-amber-600' : 'text-emerald-600'}">{formatMinutes(attendance.weekly_balance_minutes)}</div>
				<div class="mt-1 text-xs text-neutral-500">carry-over hanya dalam minggu yang sama</div>
			</article>
			<article class="rounded-3xl bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]">
				<div class="flex items-center gap-3 text-neutral-500"><BriefcaseBusiness size="20" /><span class="text-sm font-semibold">Pola kerja</span></div>
				<div class="mt-4 text-lg font-black capitalize">{attendance.identity.work_mode}</div>
				<div class="mt-1 text-xs capitalize text-neutral-500">{attendance.identity.schedule_type}</div>
			</article>
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			{#if attendance.active_session}
				<div class="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<div class="flex items-center gap-2 text-sm font-bold text-emerald-700 dark:text-emerald-400"><span class="h-2.5 w-2.5 animate-pulse rounded-full bg-emerald-500"></span>Sesi sedang berjalan</div>
						<div class="mt-3 text-4xl font-black tabular-nums">{activeElapsed()}</div>
						<div class="mt-2 text-sm text-neutral-500">Mulai {formatTime(attendance.active_session.check_in_at)} · {attendance.active_session.location_mode === "office" ? "Kantor" : "Remote"}</div>
					</div>
					<button type="button" onclick={stopWork} disabled={sending === "stop"} class="inline-flex min-h-14 items-center justify-center gap-2 rounded-full bg-red-600 px-7 text-sm font-bold text-white shadow-sm transition hover:bg-red-700 disabled:opacity-50">
						<Square size="18" fill="currentColor" /> {sending === "stop" ? "Menyimpan..." : "Selesai Kerja"}
					</button>
				</div>

				<details class="mt-5 rounded-2xl bg-amber-50 p-4 ring-1 ring-amber-200 dark:bg-amber-950/20 dark:ring-amber-900/50">
					<summary class="text-sm font-black text-amber-800 dark:text-amber-300">Ternyata tadi lupa checkout?</summary>
					<p class="mt-2 text-xs leading-5 text-amber-800/75 dark:text-amber-300/70">Sesi aktif akan ditandai incomplete, bukan dihitung sampai waktu yang salah. Masukkan jam selesai yang sebenarnya untuk approval SPV. Setelah dikirim Anda bisa langsung mulai sesi baru.</p>
					<form onsubmit={submitForgotCheckout} class="mt-4 grid gap-3 sm:grid-cols-[160px_1fr_auto] sm:items-end">
						<label class="grid gap-1.5 text-xs font-bold text-amber-900 dark:text-amber-200">Selesai sebenarnya<input bind:value={forgotCheckout.end_time} type="time" required class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900" /></label>
						<label class="grid gap-1.5 text-xs font-bold text-amber-900 dark:text-amber-200">Alasan<input bind:value={forgotCheckout.reason} type="text" required placeholder="Contoh: lupa menekan selesai kerja" class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900" /></label>
						<button type="submit" disabled={sending === "forgot-checkout"} class="inline-flex min-h-11 items-center justify-center gap-2 rounded-full bg-amber-700 px-5 text-sm font-bold text-white disabled:opacity-50"><Send size="16" />{sending === "forgot-checkout" ? "Mengirim..." : "Ajukan & Tutup"}</button>
					</form>
				</details>
			{:else}
				{#if canStartToday}
					<div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
						<div>
							<h2 class="text-xl font-black">Mulai sesi kerja</h2>
							<p class="mt-1 text-sm text-neutral-500">Roster hari ini: {attendance.today_plan.work_mode} · target {formatMinutes(attendance.today_plan.required_minutes).replace("+", "")}. Anda boleh membuat beberapa sesi dalam satu hari.</p>
						</div>
						<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
							<div class="inline-flex rounded-full bg-neutral-100 p-1 dark:bg-white/[0.06]">
								{#if allowsOffice}<button type="button" onclick={() => (locationMode = "office")} class="flex items-center gap-2 rounded-full px-4 py-2 text-sm font-bold {locationMode === 'office' ? 'bg-white text-brand-700 shadow-sm dark:bg-neutral-800 dark:text-brand-300' : 'text-neutral-500'}"><Building2 size="16" />Kantor</button>{/if}
								{#if allowsRemote}<button type="button" onclick={() => (locationMode = "remote")} class="flex items-center gap-2 rounded-full px-4 py-2 text-sm font-bold {locationMode === 'remote' ? 'bg-white text-brand-700 shadow-sm dark:bg-neutral-800 dark:text-brand-300' : 'text-neutral-500'}"><Laptop size="16" />Remote</button>{/if}
							</div>
							<button type="button" onclick={startWork} disabled={sending === "start"} class="inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-brand-600 px-6 text-sm font-bold text-white transition hover:bg-brand-700 disabled:opacity-50"><Play size="18" fill="currentColor" />{sending === "start" ? "Memulai..." : "Mulai Kerja"}</button>
						</div>
					</div>
				{:else}
					<div class="rounded-3xl bg-neutral-50 p-5 ring-1 ring-neutral-200 dark:bg-white/[0.03] dark:ring-white/[0.06]">
						<div class="flex items-start gap-3">
							<div class="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-neutral-200 text-neutral-600 dark:bg-white/[0.07] dark:text-neutral-300"><Clock3 size="20" /></div>
							<div>
								<h2 class="text-lg font-black">Tidak ada kewajiban kerja hari ini</h2>
								<p class="mt-1 text-sm leading-6 text-neutral-500">Status roster: {dayTypeLabel(attendance.today_plan.day_type)}. Tombol mulai kerja dinonaktifkan agar sesi tidak tercatat pada hari off, cuti, sakit, libur, atau libur pengganti.</p>
							</div>
						</div>
					</div>
				{/if}
			{/if}
		</section>

		<section class="grid gap-6 lg:grid-cols-[1.25fr_.75fr]">
			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<div class="flex items-center gap-2"><History size="20" /><h2 class="text-lg font-black">Sesi hari ini</h2></div>
				{#if attendance.today_sessions.length === 0}
					<div class="mt-5 rounded-2xl bg-neutral-50 p-6 text-center text-sm text-neutral-500 dark:bg-white/[0.03]">Belum ada sesi kerja hari ini.</div>
				{:else}
					<div class="mt-4 divide-y divide-neutral-100 dark:divide-white/[0.05]">
						{#each attendance.today_sessions as session}
							<div class="flex items-center gap-4 py-4">
								<div class="grid h-11 w-11 place-items-center rounded-2xl bg-brand-50 text-brand-700 dark:bg-brand-950/60 dark:text-brand-300">{#if session.location_mode === "office"}<Building2 size="18" />{:else}<Laptop size="18" />{/if}</div>
								<div class="min-w-0 flex-1"><div class="text-sm font-bold">{formatTime(session.check_in_at)} - {formatTime(session.check_out_at)}</div><div class="mt-0.5 text-xs capitalize text-neutral-500">{session.location_mode} · {session.status}</div></div>
								<div class="text-sm font-black">{session.duration_minutes ? formatMinutes(session.duration_minutes).replace("+", "") : "-"}</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>

			<div class="rounded-[28px] bg-amber-50 p-5 ring-1 ring-amber-200 dark:bg-amber-950/20 dark:ring-amber-900/50 sm:p-6">
				<div class="flex items-center gap-2 text-amber-800 dark:text-amber-300"><AlertCircle size="20" /><h2 class="text-lg font-black">Lupa satu sesi?</h2></div>
				<p class="mt-2 text-sm leading-6 text-amber-800/75 dark:text-amber-300/70">Jika satu sesi sama sekali tidak tercatat, ajukan sesi yang benar beserta lokasi. Jam baru dihitung setelah SPV approve.</p>
				<form onsubmit={submitCorrection} class="mt-5 grid gap-3">
					<input bind:value={correction.work_date} type="date" required class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900" />
					<select bind:value={correction.location_mode} required class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900"><option value="office">Kantor</option><option value="remote">Remote</option></select>
					<div class="grid grid-cols-2 gap-3"><input bind:value={correction.start_time} type="time" required class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900" /><input bind:value={correction.end_time} type="time" required class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900" /></div>
					<textarea bind:value={correction.reason} required rows="3" placeholder="Contoh: lupa check-in sesi remote 19:00-20:00" class="rounded-2xl border border-amber-200 bg-white px-4 py-3 text-sm dark:border-amber-900/60 dark:bg-neutral-900"></textarea>
					<button type="submit" disabled={sending === "correction"} class="inline-flex min-h-11 items-center justify-center gap-2 rounded-full bg-amber-700 px-5 text-sm font-bold text-white disabled:opacity-50"><Send size="16" />{sending === "correction" ? "Mengirim..." : "Ajukan Koreksi"}</button>
				</form>
			</div>
		</section>

		<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex items-start gap-3"><div class="grid h-11 w-11 place-items-center rounded-2xl bg-violet-100 text-violet-700 dark:bg-violet-950/50 dark:text-violet-300"><PlusCircle size="20" /></div><div><h2 class="text-lg font-black">Ajukan lembur</h2><p class="mt-1 text-sm text-neutral-500">Lembur terpisah dari pengganti kekurangan jam dan baru sah setelah approval SPV.</p></div></div>
			<form onsubmit={submitOvertime} class="mt-5 grid gap-3 md:grid-cols-[1fr_180px_2fr_auto] md:items-end">
				<label class="grid gap-1.5 text-xs font-bold text-neutral-600 dark:text-neutral-300">Tanggal<input bind:value={overtime.work_date} type="date" required class="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm font-medium dark:border-white/[0.07] dark:bg-neutral-900" /></label>
				<label class="grid gap-1.5 text-xs font-bold text-neutral-600 dark:text-neutral-300">Durasi (menit)<input bind:value={overtime.requested_minutes} type="number" min="30" max="720" step="30" required class="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm font-medium dark:border-white/[0.07] dark:bg-neutral-900" /></label>
				<label class="grid gap-1.5 text-xs font-bold text-neutral-600 dark:text-neutral-300">Alasan<input bind:value={overtime.reason} type="text" required placeholder="Pekerjaan tambahan yang membutuhkan lembur" class="rounded-2xl border border-neutral-200 bg-neutral-50 px-4 py-3 text-sm font-medium dark:border-white/[0.07] dark:bg-neutral-900" /></label>
				<button type="submit" disabled={sending === "overtime"} class="inline-flex min-h-11 items-center justify-center gap-2 rounded-full bg-violet-700 px-5 text-sm font-bold text-white disabled:opacity-50"><Send size="16" />Kirim</button>
			</form>
		</section>

		{#if attendance.adjustments.length > 0 || attendance.overtime_requests.length > 0}
			<section class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<h2 class="text-lg font-black">Status pengajuan</h2>
				<div class="mt-4 grid gap-3 md:grid-cols-2">
					{#each attendance.adjustments.slice(0, 4) as item}
						<div class="rounded-2xl bg-neutral-50 p-4 dark:bg-white/[0.03]"><div class="flex items-center justify-between gap-3"><strong class="text-sm">Koreksi {item.work_date}</strong><span class="rounded-full px-2.5 py-1 text-[10px] font-black uppercase {item.status === 'approved' ? 'bg-emerald-100 text-emerald-700' : item.status === 'rejected' ? 'bg-red-100 text-red-700' : 'bg-amber-100 text-amber-700'}">{item.status}</span></div><div class="mt-2 text-xs text-neutral-500">{item.location_mode === "office" ? "Kantor" : "Remote"} · {formatTime(item.requested_check_in)} - {formatTime(item.requested_check_out)}</div></div>
					{/each}
					{#each attendance.overtime_requests.slice(0, 4) as item}
						<div class="rounded-2xl bg-neutral-50 p-4 dark:bg-white/[0.03]"><div class="flex items-center justify-between gap-3"><strong class="text-sm">Lembur {item.work_date}</strong><span class="rounded-full px-2.5 py-1 text-[10px] font-black uppercase {item.status === 'approved' ? 'bg-emerald-100 text-emerald-700' : item.status === 'rejected' ? 'bg-red-100 text-red-700' : 'bg-violet-100 text-violet-700'}">{item.status}</span></div><div class="mt-2 text-xs text-neutral-500">{formatMinutes(item.requested_minutes).replace("+", "")} · {item.reason}</div></div>
					{/each}
				</div>
			</section>
		{/if}
	</div>
</AppLayout>
