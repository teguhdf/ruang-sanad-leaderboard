<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { Member, User } from "@lib/types";
	import { renderAmaliyahPosterCanvas } from "@lib/amaliyahPoster";
	import {
		Award,
		CalendarDays,
		CheckCircle2,
		CloudDownload,
		Download,
		FileDown,
		FileSpreadsheet,
		Link2,
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
		photo_url: string;
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

	type UploadRow = { file: File; member_id: number };

	let { user, members, period_month, leaderboard, flash }: Props = $props();
	let selectedPeriod = $state(period_month);
	let uploadRows = $state<UploadRow[]>([]);
	let importing = $state(false);
	let syncing = $state(false);
	let exporting = $state(false);

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
		uploadRows = Array.from(input.files ?? []).map((file) => ({ file, member_id: guessMemberID(file.name) }));
	}

	function syncGoogleSheets() {
		if (!selectedPeriod) return;
		syncing = true;
		router.post(
			"/app/amaliyah/sync",
			{ period_month: selectedPeriod },
			{ preserveScroll: true, onFinish: () => (syncing = false) },
		);
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
		return new Intl.DateTimeFormat("id-ID", { month: "long", year: "numeric" }).format(new Date(`${value}-01T00:00:00`));
	}

	function formatPercent(value: number) {
		return new Intl.NumberFormat("id-ID", { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(value ?? 0);
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

	function initials(name: string) {
		return name.trim().split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase() ?? "").join("") || "?";
	}

	function rankedIndividuals() {
		return [...leaderboard.members]
			.sort((a, b) => b.overall_score - a.overall_score || a.member_name.localeCompare(b.member_name, "id"))
			.slice(0, 10);
	}

	function downloadBlob(blob: Blob, filename: string) {
		const url = URL.createObjectURL(blob);
		const anchor = document.createElement("a");
		anchor.href = url;
		anchor.download = filename;
		document.body.appendChild(anchor);
		anchor.click();
		anchor.remove();
		setTimeout(() => URL.revokeObjectURL(url), 1000);
	}

	async function downloadPNG() {
		if (!leaderboard.members.length || exporting) return;
		exporting = true;
		try {
			const canvas = await renderAmaliyahPosterCanvas(rankedIndividuals(), formatPeriod(period_month), formatPercent);
			const blob = await new Promise<Blob | null>((resolve) => canvas.toBlob(resolve, "image/png"));
			if (!blob) throw new Error("PNG gagal dibuat");
			downloadBlob(blob, `leaderboard-amaliyah-${period_month}.png`);
		} finally {
			exporting = false;
		}
	}

	function concatBytes(chunks: Uint8Array[]) {
		const length = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
		const result = new Uint8Array(length);
		let offset = 0;
		for (const chunk of chunks) {
			result.set(chunk, offset);
			offset += chunk.length;
		}
		return result;
	}

	function base64Bytes(dataUrl: string) {
		const raw = atob(dataUrl.split(",")[1]);
		const bytes = new Uint8Array(raw.length);
		for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i);
		return bytes;
	}

	function buildImagePDF(jpeg: Uint8Array, imageWidth: number, imageHeight: number) {
		const encoder = new TextEncoder();
		const pageWidth = 595;
		const pageHeight = 842;
		const margin = 28;
		const scale = Math.min((pageWidth - margin * 2) / imageWidth, (pageHeight - margin * 2) / imageHeight);
		const drawWidth = imageWidth * scale;
		const drawHeight = imageHeight * scale;
		const x = (pageWidth - drawWidth) / 2;
		const y = (pageHeight - drawHeight) / 2;
		const content = encoder.encode(`q\n${drawWidth.toFixed(2)} 0 0 ${drawHeight.toFixed(2)} ${x.toFixed(2)} ${y.toFixed(2)} cm\n/Im0 Do\nQ\n`);
		const header = encoder.encode("%PDF-1.4\n");
		const chunks: Uint8Array[] = [header];
		const offsets = [0];
		let currentOffset = header.length;
		const addObject = (number: number, body: Uint8Array) => {
			offsets[number] = currentOffset;
			const prefix = encoder.encode(`${number} 0 obj\n`);
			const suffix = encoder.encode("\nendobj\n");
			chunks.push(prefix, body, suffix);
			currentOffset += prefix.length + body.length + suffix.length;
		};
		addObject(1, encoder.encode("<< /Type /Catalog /Pages 2 0 R >>"));
		addObject(2, encoder.encode("<< /Type /Pages /Kids [3 0 R] /Count 1 >>"));
		addObject(3, encoder.encode(`<< /Type /Page /Parent 2 0 R /MediaBox [0 0 ${pageWidth} ${pageHeight}] /Resources << /XObject << /Im0 4 0 R >> >> /Contents 5 0 R >>`));
		addObject(4, concatBytes([encoder.encode(`<< /Type /XObject /Subtype /Image /Width ${imageWidth} /Height ${imageHeight} /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length ${jpeg.length} >>\nstream\n`), jpeg, encoder.encode("\nendstream")]));
		addObject(5, concatBytes([encoder.encode(`<< /Length ${content.length} >>\nstream\n`), content, encoder.encode("endstream")]));
		const xrefOffset = currentOffset;
		let xref = "xref\n0 6\n0000000000 65535 f \n";
		for (let i = 1; i <= 5; i++) xref += `${String(offsets[i]).padStart(10, "0")} 00000 n \n`;
		chunks.push(encoder.encode(`${xref}trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF`));
		return concatBytes(chunks);
	}

	async function downloadPDF() {
		if (!leaderboard.members.length || exporting) return;
		exporting = true;
		try {
			const canvas = await renderAmaliyahPosterCanvas(rankedIndividuals(), formatPeriod(period_month), formatPercent);
			const jpeg = base64Bytes(canvas.toDataURL("image/jpeg", 0.94));
			const pdf = buildImagePDF(jpeg, canvas.width, canvas.height);
			downloadBlob(new Blob([pdf], { type: "application/pdf" }), `leaderboard-amaliyah-${period_month}.pdf`);
		} finally {
			exporting = false;
		}
	}
</script>

<svelte:head>
	<title>Leaderboard Amaliyah - Ruang Sanad Ops</title>
</svelte:head>

<AppLayout {user} group="amaliyah">
	<div class="mx-auto max-w-7xl space-y-6 px-4 py-6 sm:px-6 lg:px-8 lg:py-8">
		<section class="overflow-hidden rounded-[30px] bg-gradient-to-br from-emerald-950 via-emerald-800 to-emerald-700 p-6 text-white shadow-lg shadow-emerald-950/10 sm:p-8">
			<div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
				<div class="max-w-3xl">
					<div class="flex items-center gap-2 text-sm font-black text-amber-200"><Trophy size="19" /> Leaderboard Amaliyah</div>
					<h1 class="mt-3 text-3xl font-black tracking-tight sm:text-4xl">Konsistensi terbaik bulan ini</h1>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-emerald-50/80">Ranking divisi tetap terlihat di aplikasi, sedangkan file PNG/PDF menampilkan Top 10 individu dengan foto, nama, rank, dan overall score.</p>
				</div>
				<div class="grid min-w-[280px] gap-3 rounded-3xl bg-white/10 p-4 backdrop-blur-sm ring-1 ring-white/15">
					<div><div class="text-xs font-bold uppercase tracking-[0.12em] text-emerald-100/60">Periode aktif</div><div class="mt-1 text-2xl font-black capitalize">{formatPeriod(period_month)}</div><div class="mt-1 text-xs text-emerald-100/70">{leaderboard.members.length} snapshot anggota tersimpan</div></div>
					{#if leaderboard.members.length > 0}<div class="grid grid-cols-2 gap-2"><button type="button" onclick={downloadPNG} disabled={exporting} class="flex min-h-10 items-center justify-center gap-2 rounded-full bg-amber-200 px-3 text-xs font-black text-emerald-950 disabled:opacity-50"><Download size="15" /> PNG</button><button type="button" onclick={downloadPDF} disabled={exporting} class="flex min-h-10 items-center justify-center gap-2 rounded-full bg-white/10 px-3 text-xs font-black text-white ring-1 ring-white/20 disabled:opacity-50"><FileDown size="15" /> PDF</button></div>{/if}
				</div>
			</div>
		</section>

		{#if flash?.success}<div class="flex items-center gap-3 rounded-2xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-800 ring-1 ring-emerald-200"><CheckCircle2 size="18" />{flash.success}</div>{/if}
		{#if flash?.error}<div class="rounded-2xl bg-red-50 px-4 py-3 text-sm font-bold text-red-800 ring-1 ring-red-200">{flash.error}</div>{/if}

		<section class="grid gap-5 xl:grid-cols-[minmax(0,1.35fr)_minmax(300px,.65fr)]">
			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<div class="flex items-start gap-3"><div class="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-emerald-100 text-emerald-700"><Link2 size="21" /></div><div><div class="text-sm font-black text-emerald-700">Google Sheets Sync</div><h2 class="mt-1 text-xl font-black">Hitung semua tim tanpa upload file</h2><p class="mt-1 max-w-2xl text-sm leading-6 text-neutral-500">Aplikasi memakai link workspace yang tersimpan pada setiap anggota. Pilih bulan lalu sinkronkan.</p></div></div>
				<div class="mt-5 grid gap-4 sm:grid-cols-[220px_1fr] sm:items-end"><label class="grid gap-1.5 text-xs font-bold text-neutral-500">Bulan yang dihitung<input bind:value={selectedPeriod} type="month" class="min-h-12 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold text-neutral-900 outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/20 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-white" /></label><div class="grid gap-2"><div class="text-xs font-bold text-neutral-500">Workspace terdaftar</div><div class="flex min-h-12 items-center justify-between gap-3 rounded-2xl bg-neutral-50 px-4 ring-1 ring-neutral-200 dark:bg-white/[0.03] dark:ring-white/[0.07]"><div><div class="text-sm font-black">{members.length} anggota</div><div class="text-[11px] text-neutral-500">Semua workspace diproses untuk bulan terpilih</div></div><CloudDownload size="21" class="text-emerald-600" /></div></div></div>
				<button type="button" onclick={syncGoogleSheets} disabled={syncing || !selectedPeriod || members.length === 0} class="mt-5 inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-emerald-700 px-7 text-sm font-black text-white shadow-sm transition hover:bg-emerald-800 disabled:opacity-45">{#if syncing}<RefreshCw size="17" class="animate-spin" /> Sinkronisasi...{:else}<RefreshCw size="17" /> Sync & Hitung Semua Tim{/if}</button>
			</div>
			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6"><div class="flex items-center gap-2 text-sm font-black text-emerald-700"><CalendarDays size="18" /> Tracking bulanan</div><h2 class="mt-1 text-xl font-black">Lihat bulan lain</h2><p class="mt-1 text-sm text-neutral-500">Snapshot tersimpan per anggota dan per bulan.</p><div class="mt-5 grid gap-3"><input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold outline-none dark:border-white/[0.07] dark:bg-neutral-900" /><button type="button" onclick={changePeriod} class="min-h-11 rounded-full bg-neutral-900 px-4 text-sm font-black text-white dark:bg-white dark:text-neutral-900">Tampilkan bulan</button></div>{#if leaderboard.periods.length > 0}<div class="mt-5 flex flex-wrap gap-2">{#each leaderboard.periods as savedPeriod}<a href={`/app/amaliyah?month=${savedPeriod}`} class="rounded-full px-3 py-1.5 text-xs font-bold {savedPeriod === period_month ? 'bg-emerald-100 text-emerald-800' : 'bg-neutral-100 text-neutral-600 dark:bg-white/[0.05] dark:text-neutral-300'}">{formatPeriod(savedPeriod)}</a>{/each}</div>{/if}</div>
		</section>

		<section class="rounded-[30px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between"><div><div class="flex items-center gap-2 text-sm font-black text-emerald-700"><Medal size="18" /> Ranking divisi</div><h2 class="mt-1 text-2xl font-black capitalize">Leaderboard {formatPeriod(period_month)}</h2><p class="mt-1 text-sm text-neutral-500">Overall adalah rata-rata score individu. Empat amaliyah tetap ditampilkan terpisah.</p></div><div class="flex items-center gap-2"><div class="rounded-full bg-neutral-100 px-3 py-1.5 text-xs font-black text-neutral-600">{leaderboard.teams.length} tim</div>{#if leaderboard.members.length > 0}<button type="button" onclick={downloadPNG} disabled={exporting} class="grid h-9 w-9 place-items-center rounded-full bg-emerald-100 text-emerald-700" aria-label="Download PNG poster individu"><Download size="16" /></button><button type="button" onclick={downloadPDF} disabled={exporting} class="grid h-9 w-9 place-items-center rounded-full bg-neutral-100 text-neutral-700" aria-label="Download PDF poster individu"><FileDown size="16" /></button>{/if}</div></div>
			{#if leaderboard.teams.length === 0}<div class="mt-6 grid min-h-56 place-items-center rounded-3xl border border-dashed border-neutral-200 bg-neutral-50 p-8 text-center"><div><Trophy size="36" class="mx-auto text-neutral-300" /><div class="mt-3 text-base font-black">Belum ada data untuk bulan ini</div><div class="mt-1 text-sm text-neutral-500">Klik “Sync & Hitung Semua Tim” untuk mengambil data Google Sheets.</div></div></div>{:else}<div class="mt-6 grid gap-4">{#each leaderboard.teams as team}<article class="grid gap-5 rounded-3xl border p-5 {team.rank === 1 ? 'border-amber-300 bg-amber-50/55' : 'border-neutral-200 bg-neutral-50/70'} lg:grid-cols-[180px_minmax(0,1fr)] lg:items-center"><div class="flex items-center gap-4 lg:block"><div class="grid h-14 w-14 place-items-center rounded-2xl bg-white text-2xl shadow-sm ring-1 ring-neutral-200">{rankLabel(team.rank)}</div><div class="lg:mt-3"><div class="text-lg font-black">{team.division}</div><div class="mt-1 flex items-center gap-1 text-xs text-neutral-500"><UsersRound size="14" />{team.member_count} anggota</div>{#if team.rank === 1}<div class="mt-2 inline-flex items-center gap-1 rounded-full bg-amber-100 px-2.5 py-1 text-[10px] font-black uppercase tracking-wide text-amber-800"><Award size="13" /> Kandidat reward</div>{/if}</div></div><div><div class="flex items-end justify-between gap-3"><div class="text-xs font-bold uppercase tracking-[0.12em] text-neutral-400">Overall score</div><div class="text-3xl font-black tracking-tight">{formatPercent(team.overall_score)}%</div></div><div class="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">{#each team.metrics as metric}<div class="rounded-2xl bg-white p-3 ring-1 ring-neutral-200"><div class="flex items-center justify-between gap-2"><span class="truncate text-xs font-bold text-neutral-600">{metric.label}</span><span class="text-xs font-black">{formatPercent(metric.score)}%</span></div><div class="mt-2 h-2 overflow-hidden rounded-full bg-neutral-100"><div class="h-full rounded-full bg-emerald-600" style={`width: ${scoreWidth(metric.score)}`}></div></div></div>{/each}</div></div></article>{/each}</div>{/if}
		</section>

		<section class="overflow-hidden rounded-[30px] bg-white shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="flex items-center justify-between gap-4 border-b border-neutral-200 px-5 py-5 sm:px-6"><div><div class="text-sm font-black text-emerald-700">Detail individu</div><h2 class="mt-1 text-xl font-black">Kontribusi anggota</h2></div><div class="text-xs font-bold text-neutral-500">{leaderboard.members.length} data</div></div>{#if leaderboard.members.length > 0}<div class="overflow-x-auto"><table class="min-w-[920px] w-full text-left"><thead class="bg-neutral-50 text-[11px] font-black uppercase tracking-wide text-neutral-500"><tr><th class="px-6 py-3">Anggota</th><th class="px-4 py-3">Divisi</th><th class="px-4 py-3 text-right">Overall</th><th class="px-4 py-3 text-right">Tahajud</th><th class="px-4 py-3 text-right">Tilawah + Tadabur</th><th class="px-4 py-3 text-right">Dzikir Pagi</th><th class="px-4 py-3 text-right">Dzikir Petang</th></tr></thead><tbody class="divide-y divide-neutral-100">{#each leaderboard.members as member}<tr class="text-sm"><td class="px-6 py-4"><div class="flex items-center gap-3"><div class="grid h-9 w-9 shrink-0 place-items-center overflow-hidden rounded-full bg-emerald-50 text-xs font-black text-emerald-800">{#if member.photo_url}<img src={member.photo_url} alt={`Foto ${member.member_name}`} class="h-full w-full object-cover" />{:else}{initials(member.member_name)}{/if}</div><div><div class="font-black">{member.member_name}</div><div class="mt-0.5 max-w-64 truncate text-[10px] text-neutral-400" title={member.source_file_name}>{member.source_file_name}</div></div></div></td><td class="px-4 py-4 text-neutral-500">{member.division}</td><td class="px-4 py-4 text-right text-base font-black">{formatPercent(member.overall_score)}%</td>{#each metricKeys as key}<td class="px-4 py-4 text-right"><div class="font-bold">{formatPercent(memberMetric(member, key))}%</div><div class="mt-0.5 text-[10px] text-neutral-400">{memberMetricAvailable(member, key)} checkbox</div></td>{/each}</tr>{/each}</tbody></table></div>{:else}<div class="px-6 py-10 text-center text-sm text-neutral-500">Belum ada snapshot anggota pada periode ini.</div>{/if}</section>

		<details class="group rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6"><summary class="cursor-pointer list-none"><div class="flex items-center justify-between gap-3"><div><div class="flex items-center gap-2 text-sm font-black text-neutral-600"><UploadCloud size="18" /> Fallback manual</div><div class="mt-1 text-base font-black">Upload XLSX hanya jika Google Sheets gagal</div></div><FileSpreadsheet size="20" class="text-neutral-400" /></div></summary><div class="mt-5 grid gap-4 border-t border-neutral-100 pt-5"><div class="grid gap-4 sm:grid-cols-[220px_1fr]"><label class="grid gap-1.5 text-xs font-bold text-neutral-500">Bulan<input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold outline-none" /></label><label class="grid cursor-pointer gap-1.5 text-xs font-bold text-neutral-500">Workspace anggota<div class="flex min-h-11 items-center gap-3 rounded-2xl border border-dashed border-neutral-300 px-4 text-sm font-bold"><FileSpreadsheet size="18" /><span>{uploadRows.length ? `${uploadRows.length} file dipilih` : "Pilih file .xlsx"}</span></div><input type="file" accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" multiple class="hidden" onchange={handleFiles} /></label></div>{#if uploadRows.length > 0}<div class="grid gap-2">{#each uploadRows as row, index}<div class="grid gap-3 rounded-2xl bg-neutral-50 p-3 sm:grid-cols-[minmax(0,1fr)_260px] sm:items-center"><div class="min-w-0"><div class="truncate text-sm font-black">{row.file.name}</div><div class="mt-0.5 text-[11px] text-neutral-500">{(row.file.size / 1024).toFixed(0)} KB</div></div><select bind:value={uploadRows[index].member_id} class="min-h-10 rounded-xl border border-neutral-200 bg-white px-3 text-sm font-bold outline-none"><option value={0}>Pilih anggota...</option>{#each members as member}<option value={member.id}>{member.name} · {member.division}</option>{/each}</select></div>{/each}</div><button type="button" onclick={importWorkspaces} disabled={importing || uploadRows.some((row) => !row.member_id)} class="inline-flex min-h-11 w-fit items-center justify-center gap-2 rounded-full bg-neutral-900 px-5 text-sm font-black text-white disabled:opacity-45">{#if importing}<RefreshCw size="16" class="animate-spin" /> Menghitung...{:else}<UploadCloud size="16" /> Hitung upload manual{/if}</button>{/if}</div></details>
	</div>
</AppLayout>
