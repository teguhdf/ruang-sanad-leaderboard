<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { Member, User } from "@lib/types";
	import {
		Award,
		CalendarDays,
		CheckCircle2,
		FileSpreadsheet,
		Medal,
		RefreshCw,
		Trophy,
		UploadCloud,
		UsersRound,
	} from "lucide-svelte";

	interface AmaliyahMetricScore {
		key: "tahajud" | "tilawah_tadabur" | "dzikir_pagi" | "dzikir_petang";
		label: string;
		completed: number;
		available: number;
		score: number;
	}

	interface AmaliyahMemberSnapshot {
		id: number;
		member_id: number;
		member_name: string;
		division: string;
		period_month: string;
		source_file_name: string;
		overall_score: number;
		metrics: AmaliyahMetricScore[];
		updated_at: string;
	}

	interface AmaliyahTeamLeaderboard {
		rank: number;
		division: string;
		member_count: number;
		overall_score: number;
		metrics: AmaliyahMetricScore[];
	}

	interface AmaliyahLeaderboard {
		period_month: string;
		teams: AmaliyahTeamLeaderboard[];
		members: AmaliyahMemberSnapshot[];
		periods: string[];
	}

	interface Props {
		user?: User;
		members: Member[];
		period_month: string;
		leaderboard: AmaliyahLeaderboard;
		flash?: { success?: string; error?: string };
	}

	type UploadRow = {
		file: File;
		member_id: number;
	};

	let { user, members, period_month, leaderboard, flash }: Props = $props();
	let selectedPeriod = $state(period_month);
	let uploadRows = $state<UploadRow[]>([]);
	let importing = $state(false);

	const metricKeys = ["tahajud", "tilawah_tadabur", "dzikir_pagi", "dzikir_petang"];

	function normalizeName(value: string) {
		return value
			.toLowerCase()
			.replace(/\.xlsx$/i, "")
			.replace(/workspace|202[0-9]|spv|leader|supervisor/gi, " ")
			.replace(/\([^)]*\)/g, " ")
			.replace(/[^a-z0-9]+/g, " ")
			.trim()
			.replace(/\s+/g, " ");
	}

	function guessMemberID(fileName: string) {
		const fileKey = normalizeName(fileName);
		let bestID = 0;
		let bestScore = 0;
		for (const member of members) {
			const memberKey = normalizeName(member.name);
			if (!memberKey) continue;
			if (fileKey.includes(memberKey)) return member.id;
			const tokens = memberKey.split(" ").filter((token) => token.length > 1);
			const score = tokens.filter((token) => fileKey.includes(token)).length;
			if (score > bestScore && score === tokens.length) {
				bestScore = score;
				bestID = member.id;
			}
		}
		return bestID;
	}

	function handleFiles(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const files = Array.from(input.files ?? []);
		uploadRows = files.map((file) => ({ file, member_id: guessMemberID(file.name) }));
	}

	function importWorkspaces() {
		if (!selectedPeriod || uploadRows.length === 0) return;
		if (uploadRows.some((row) => !row.member_id)) {
			alert("Pasangkan semua workspace ke anggota terlebih dahulu.");
			return;
		}

		const data = new FormData();
		data.append("period_month", selectedPeriod);
		for (const row of uploadRows) {
			data.append("files", row.file);
			data.append("member_ids", String(row.member_id));
		}

		importing = true;
		router.post("/app/amaliyah/import", data, {
			forceFormData: true,
			preserveScroll: true,
			onFinish: () => (importing = false),
		});
	}

	function changePeriod() {
		router.get("/app/amaliyah", { month: selectedPeriod }, { preserveState: false });
	}

	function formatPeriod(value: string) {
		if (!value) return "-";
		return new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(
			new Date(`${value}-01T00:00:00`),
		);
	}

	function formatPercent(value: number) {
		return new Intl.NumberFormat("id-ID", {
			minimumFractionDigits: 1,
			maximumFractionDigits: 1,
		}).format(value ?? 0);
	}

	function scoreWidth(value: number) {
		return `${Math.max(0, Math.min(100, value ?? 0))}%`;
	}

	function rankLabel(rank: number) {
		if (rank === 1) return "🥇";
		if (rank === 2) return "🥈";
		if (rank === 3) return "🥉";
		return `#${rank}`;
	}

	function memberMetric(member: AmaliyahMemberSnapshot, key: string) {
		return member.metrics.find((metric) => metric.key === key)?.score ?? 0;
	}

	function memberMetricAvailable(member: AmaliyahMemberSnapshot, key: string) {
		return member.metrics.find((metric) => metric.key === key)?.available ?? 0;
	}
</script>

<svelte:head>
	<title>Leaderboard Amaliyah - Ruang Sanad Ops</title>
</svelte:head>

<AppLayout {user} group="amaliyah">
	<div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="overflow-hidden rounded-[30px] bg-gradient-to-br from-brand-700 via-brand-600 to-brand-500 p-6 text-white shadow-lg shadow-brand-900/10 sm:p-8">
			<div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
				<div class="max-w-3xl">
					<div class="flex items-center gap-2 text-sm font-black text-white/80"><Trophy size="19" /> Leaderboard Amaliyah Tim</div>
					<h1 class="mt-3 text-3xl font-black tracking-tight sm:text-4xl">Siapa tim paling konsisten bulan ini?</h1>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-white/80">Empat amaliyah yang dinilai: Tahajud, Tilawah + Tadabur, Dzikir Pagi, dan Dzikir Petang. Score tim adalah rata-rata persentase seluruh anggotanya.</p>
				</div>
				<div class="grid min-w-[260px] gap-2 rounded-3xl bg-white/10 p-4 backdrop-blur-sm ring-1 ring-white/15">
					<div class="text-xs font-bold uppercase tracking-[0.12em] text-white/60">Periode aktif</div>
					<div class="text-2xl font-black capitalize">{formatPeriod(period_month)}</div>
					<div class="text-xs text-white/70">{leaderboard.members.length} workspace anggota tersimpan</div>
				</div>
			</div>
		</section>

		{#if flash?.success}
			<div class="flex items-center gap-3 rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 ring-1 ring-emerald-200 dark:bg-emerald-950/30 dark:text-emerald-300 dark:ring-emerald-900/40"><CheckCircle2 size="18" />{flash.success}</div>
		{/if}
		{#if flash?.error}
			<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 ring-1 ring-red-200 dark:bg-red-950/30 dark:text-red-300 dark:ring-red-900/40">{flash.error}</div>
		{/if}

		<section class="grid gap-5 xl:grid-cols-[minmax(0,1.35fr)_minmax(300px,.65fr)]">
			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<div class="flex items-center gap-2 text-sm font-black text-brand-700 dark:text-brand-300"><UploadCloud size="18" /> Upload workspace tim</div>
						<h2 class="mt-1 text-xl font-black">Hitung leaderboard dalam sekali upload</h2>
						<p class="mt-1 text-sm text-neutral-500">Bisa pilih banyak file .xlsx sekaligus. File yang sama untuk anggota + bulan yang sama akan memperbarui snapshot, bukan membuat duplikat.</p>
					</div>
				</div>

				<div class="mt-5 grid gap-4 sm:grid-cols-[220px_1fr]">
					<label class="grid gap-1.5 text-xs font-bold text-neutral-500">
						Bulan yang dihitung
						<input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold text-neutral-900 outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/20 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-white" />
					</label>
					<label class="grid cursor-pointer gap-1.5 text-xs font-bold text-neutral-500">
						Workspace anggota
						<div class="flex min-h-11 items-center gap-3 rounded-2xl border border-dashed border-brand-300 bg-brand-50/60 px-4 text-sm font-bold text-brand-800 transition hover:bg-brand-50 dark:border-brand-800 dark:bg-brand-950/25 dark:text-brand-200">
							<FileSpreadsheet size="18" />
							<span>{uploadRows.length ? `${uploadRows.length} file dipilih` : "Pilih file .xlsx"}</span>
						</div>
						<input type="file" accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" multiple class="hidden" onchange={handleFiles} />
					</label>
				</div>

				{#if uploadRows.length > 0}
					<div class="mt-5 grid gap-2">
						{#each uploadRows as row, index}
							<div class="grid gap-3 rounded-2xl bg-neutral-50 p-3 dark:bg-white/[0.03] sm:grid-cols-[minmax(0,1fr)_260px] sm:items-center">
								<div class="min-w-0"><div class="truncate text-sm font-black">{row.file.name}</div><div class="mt-0.5 text-[11px] text-neutral-500">{(row.file.size / 1024).toFixed(0)} KB</div></div>
								<select bind:value={uploadRows[index].member_id} class="min-h-10 rounded-xl border border-neutral-200 bg-white px-3 text-sm font-bold outline-none focus:border-brand-400 dark:border-white/[0.08] dark:bg-neutral-900">
									<option value={0}>Pilih anggota...</option>
									{#each members as member}
										<option value={member.id}>{member.name} · {member.division}</option>
									{/each}
								</select>
							</div>
						{/each}
					</div>

					<button type="button" onclick={importWorkspaces} disabled={importing || uploadRows.some((row) => !row.member_id)} class="mt-5 inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-brand-600 px-6 text-sm font-black text-white shadow-sm transition hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-45">
						{#if importing}<RefreshCw size="17" class="animate-spin" /> Menghitung...{:else}<UploadCloud size="17" /> Hitung {uploadRows.length} workspace{/if}
					</button>
				{/if}
			</div>

			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<div class="flex items-center gap-2 text-sm font-black text-brand-700 dark:text-brand-300"><CalendarDays size="18" /> Tracking bulanan</div>
				<h2 class="mt-1 text-xl font-black">Lihat bulan lain</h2>
				<p class="mt-1 text-sm text-neutral-500">Snapshot tersimpan per anggota dan per bulan.</p>
				<div class="mt-5 grid gap-3">
					<input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold outline-none focus:border-brand-400 dark:border-white/[0.07] dark:bg-neutral-900" />
					<button type="button" onclick={changePeriod} class="min-h-11 rounded-full bg-neutral-900 px-4 text-sm font-black text-white dark:bg-white dark:text-neutral-900">Tampilkan bulan</button>
				</div>
				{#if leaderboard.periods.length > 0}
					<div class="mt-5 flex flex-wrap gap-2">
						{#each leaderboard.periods as savedPeriod}
							<a href={`/app/amaliyah?month=${savedPeriod}`} class="rounded-full px-3 py-1.5 text-xs font-bold {savedPeriod === period_month ? 'bg-brand-100 text-brand-800 dark:bg-brand-950/60 dark:text-brand-200' : 'bg-neutral-100 text-neutral-600 dark:bg-white/[0.05] dark:text-neutral-300'}">{formatPeriod(savedPeriod)}</a>
						{/each}
					</div>
				{/if}
			</div>
		</section>

		<section class="rounded-[30px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
				<div><div class="flex items-center gap-2 text-sm font-black text-brand-700 dark:text-brand-300"><Medal size="18" /> Ranking divisi</div><h2 class="mt-1 text-2xl font-black capitalize">Leaderboard {formatPeriod(period_month)}</h2><p class="mt-1 text-sm text-neutral-500">Overall adalah rata-rata score individu. Setiap amaliyah juga dirata-ratakan per anggota agar ukuran tim tidak memengaruhi ranking.</p></div>
				<div class="rounded-full bg-neutral-100 px-3 py-1.5 text-xs font-black text-neutral-600 dark:bg-white/[0.05] dark:text-neutral-300">{leaderboard.teams.length} tim</div>
			</div>

			{#if leaderboard.teams.length === 0}
				<div class="mt-6 grid min-h-56 place-items-center rounded-3xl border border-dashed border-neutral-200 bg-neutral-50 p-8 text-center dark:border-white/[0.08] dark:bg-white/[0.02]">
					<div><Trophy size="36" class="mx-auto text-neutral-300" /><div class="mt-3 text-base font-black">Belum ada data untuk bulan ini</div><div class="mt-1 text-sm text-neutral-500">Upload workspace anggota untuk mulai membuat leaderboard.</div></div>
				</div>
			{:else}
				<div class="mt-6 grid gap-4">
					{#each leaderboard.teams as team}
						<article class="grid gap-5 rounded-3xl border p-5 {team.rank === 1 ? 'border-amber-300 bg-amber-50/55 dark:border-amber-800 dark:bg-amber-950/15' : 'border-neutral-200 bg-neutral-50/70 dark:border-white/[0.07] dark:bg-white/[0.025]'} lg:grid-cols-[180px_minmax(0,1fr)] lg:items-center">
							<div class="flex items-center gap-4 lg:block">
								<div class="grid h-14 w-14 place-items-center rounded-2xl bg-white text-2xl shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.07]">{rankLabel(team.rank)}</div>
								<div class="lg:mt-3"><div class="text-lg font-black">{team.division}</div><div class="mt-1 flex items-center gap-1 text-xs text-neutral-500"><UsersRound size="14" />{team.member_count} anggota</div>{#if team.rank === 1}<div class="mt-2 inline-flex items-center gap-1 rounded-full bg-amber-100 px-2.5 py-1 text-[10px] font-black uppercase tracking-wide text-amber-800"><Award size="13" /> Kandidat reward</div>{/if}</div>
							</div>
							<div>
								<div class="flex items-end justify-between gap-3"><div class="text-xs font-bold uppercase tracking-[0.12em] text-neutral-400">Overall score</div><div class="text-3xl font-black tracking-tight">{formatPercent(team.overall_score)}%</div></div>
								<div class="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
									{#each team.metrics as metric}
										<div class="rounded-2xl bg-white p-3 ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.06]"><div class="flex items-center justify-between gap-2"><span class="truncate text-xs font-bold text-neutral-600 dark:text-neutral-300">{metric.label}</span><span class="text-xs font-black">{formatPercent(metric.score)}%</span></div><div class="mt-2 h-2 overflow-hidden rounded-full bg-neutral-100 dark:bg-white/[0.06]"><div class="h-full rounded-full bg-brand-500" style={`width: ${scoreWidth(metric.score)}`}></div></div></div>
									{/each}
								</div>
							</div>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<section class="overflow-hidden rounded-[30px] bg-white shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]">
			<div class="flex items-center justify-between gap-4 border-b border-neutral-200 px-5 py-5 dark:border-white/[0.06] sm:px-6"><div><div class="text-sm font-black text-brand-700 dark:text-brand-300">Detail individu</div><h2 class="mt-1 text-xl font-black">Kontribusi anggota</h2></div><div class="text-xs font-bold text-neutral-500">{leaderboard.members.length} data</div></div>
			{#if leaderboard.members.length > 0}
				<div class="overflow-x-auto">
					<table class="min-w-[920px] w-full text-left">
						<thead class="bg-neutral-50 text-[11px] font-black uppercase tracking-wide text-neutral-500 dark:bg-white/[0.025]"><tr><th class="px-6 py-3">Anggota</th><th class="px-4 py-3">Divisi</th><th class="px-4 py-3 text-right">Overall</th><th class="px-4 py-3 text-right">Tahajud</th><th class="px-4 py-3 text-right">Tilawah + Tadabur</th><th class="px-4 py-3 text-right">Dzikir Pagi</th><th class="px-4 py-3 text-right">Dzikir Petang</th></tr></thead>
						<tbody class="divide-y divide-neutral-100 dark:divide-white/[0.05]">
							{#each leaderboard.members as member}
								<tr class="text-sm"><td class="px-6 py-4"><div class="font-black">{member.member_name}</div><div class="mt-0.5 max-w-64 truncate text-[10px] text-neutral-400" title={member.source_file_name}>{member.source_file_name}</div></td><td class="px-4 py-4 text-neutral-500">{member.division}</td><td class="px-4 py-4 text-right text-base font-black">{formatPercent(member.overall_score)}%</td>{#each metricKeys as key}<td class="px-4 py-4 text-right"><div class="font-bold">{formatPercent(memberMetric(member, key))}%</div><div class="mt-0.5 text-[10px] text-neutral-400">{memberMetricAvailable(member, key)} checkbox</div></td>{/each}</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{:else}
				<div class="px-6 py-10 text-center text-sm text-neutral-500">Belum ada snapshot anggota pada periode ini.</div>
			{/if}
		</section>
	</div>
</AppLayout>
