<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { Member, User } from "@lib/types";
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

	type UploadRow = {
		file: File;
		member_id: number;
	};

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
		const files = Array.from(input.files ?? []);
		uploadRows = files.map((file) => ({ file, member_id: guessMemberID(file.name) }));
	}

	function syncGoogleSheets() {
		if (!selectedPeriod) return;
		syncing = true;
		router.post(
			"/app/amaliyah/sync",
			{ period_month: selectedPeriod },
			{
				preserveScroll: true,
				onFinish: () => (syncing = false),
			},
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

	function rankedIndividuals() {
		return [...leaderboard.members]
			.sort((a, b) => {
				if (b.overall_score === a.overall_score) {
					return a.member_name.localeCompare(b.member_name, "id");
				}
				return b.overall_score - a.overall_score;
			})
			.slice(0, 10);
	}

	function initials(name: string) {
		return name
			.trim()
			.split(/\s+/)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() ?? "")
			.join("") || "?";
	}

	function roundedRectPath(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number) {
		const r = Math.min(radius, width / 2, height / 2);
		ctx.beginPath();
		ctx.moveTo(x + r, y);
		ctx.lineTo(x + width - r, y);
		ctx.quadraticCurveTo(x + width, y, x + width, y + r);
		ctx.lineTo(x + width, y + height - r);
		ctx.quadraticCurveTo(x + width, y + height, x + width - r, y + height);
		ctx.lineTo(x + r, y + height);
		ctx.quadraticCurveTo(x, y + height, x, y + height - r);
		ctx.lineTo(x, y + r);
		ctx.quadraticCurveTo(x, y, x + r, y);
		ctx.closePath();
	}

	function fillRoundedRect(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number, fill: string) {
		ctx.save();
		ctx.fillStyle = fill;
		roundedRectPath(ctx, x, y, width, height, radius);
		ctx.fill();
		ctx.restore();
	}

	function fitText(ctx: CanvasRenderingContext2D, value: string, maxWidth: number) {
		if (ctx.measureText(value).width <= maxWidth) return value;
		let trimmed = value;
		while (trimmed.length > 1 && ctx.measureText(`${trimmed}…`).width > maxWidth) {
			trimmed = trimmed.slice(0, -1);
		}
		return `${trimmed}…`;
	}

	function loadImage(src: string) {
		return new Promise<HTMLImageElement | null>((resolve) => {
			if (!src) {
				resolve(null);
				return;
			}
			const image = new Image();
			image.onload = () => resolve(image);
			image.onerror = () => resolve(null);
			image.src = src;
		});
	}

	function drawCoverImage(ctx: CanvasRenderingContext2D, image: HTMLImageElement, x: number, y: number, size: number) {
		const scale = Math.max(size / image.naturalWidth, size / image.naturalHeight);
		const sourceWidth = size / scale;
		const sourceHeight = size / scale;
		const sourceX = (image.naturalWidth - sourceWidth) / 2;
		const sourceY = (image.naturalHeight - sourceHeight) / 2;
		ctx.drawImage(image, sourceX, sourceY, sourceWidth, sourceHeight, x, y, size, size);
	}

	function drawAvatar(
		ctx: CanvasRenderingContext2D,
		member: AmaliyahMemberSnapshot,
		image: HTMLImageElement | null,
		cx: number,
		cy: number,
		radius: number,
	) {
		ctx.save();
		ctx.beginPath();
		ctx.arc(cx, cy, radius, 0, Math.PI * 2);
		ctx.clip();
		if (image) {
			drawCoverImage(ctx, image, cx - radius, cy - radius, radius * 2);
		} else {
			ctx.fillStyle = "#DDECE6";
			ctx.fillRect(cx - radius, cy - radius, radius * 2, radius * 2);
			ctx.fillStyle = "#155B49";
			ctx.font = `800 ${Math.round(radius * 0.58)}px Arial, sans-serif`;
			ctx.textAlign = "center";
			ctx.textBaseline = "middle";
			ctx.fillText(initials(member.member_name), cx, cy + 2);
		}
		ctx.restore();

		ctx.save();
		ctx.strokeStyle = "#FFFFFF";
		ctx.lineWidth = 7;
		ctx.beginPath();
		ctx.arc(cx, cy, radius + 2, 0, Math.PI * 2);
		ctx.stroke();
		ctx.strokeStyle = "#D4B46A";
		ctx.lineWidth = 3;
		ctx.beginPath();
		ctx.arc(cx, cy, radius + 7, 0, Math.PI * 2);
		ctx.stroke();
		ctx.restore();
	}

	function drawRankBadge(ctx: CanvasRenderingContext2D, rank: number, x: number, y: number, size: number) {
		const fill = rank === 1 ? "#C99A3D" : rank === 2 ? "#8D99A5" : rank === 3 ? "#B7774A" : "#176B55";
		ctx.save();
		ctx.fillStyle = fill;
		ctx.beginPath();
		ctx.arc(x, y, size, 0, Math.PI * 2);
		ctx.fill();
		ctx.strokeStyle = "#FFFFFF";
		ctx.lineWidth = 4;
		ctx.stroke();
		ctx.fillStyle = "#FFFFFF";
		ctx.font = `900 ${Math.round(size * 0.95)}px Arial, sans-serif`;
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(String(rank), x, y + 1);
		ctx.restore();
	}

	function drawMemberLabel(
		ctx: CanvasRenderingContext2D,
		member: AmaliyahMemberSnapshot,
		x: number,
		y: number,
		maxWidth: number,
		featured: boolean,
	) {
		ctx.save();
		ctx.fillStyle = "#19352D";
		ctx.font = `${featured ? 900 : 800} ${featured ? 30 : 22}px Arial, sans-serif`;
		ctx.textAlign = "center";
		ctx.textBaseline = "alphabetic";
		ctx.fillText(fitText(ctx, member.member_name, maxWidth), x, y);

		const score = `${formatPercent(member.overall_score)}%`;
		ctx.font = `800 ${featured ? 18 : 15}px Arial, sans-serif`;
		const scoreWidth = ctx.measureText(score).width + 30;
		fillRoundedRect(ctx, x - scoreWidth / 2, y + 13, scoreWidth, featured ? 34 : 30, 17, featured ? "#F3E3B7" : "#E8F1ED");
		ctx.fillStyle = featured ? "#8C651C" : "#155B49";
		ctx.textBaseline = "middle";
		ctx.fillText(score, x, y + (featured ? 30 : 28));
		ctx.restore();
	}

	async function renderPosterCanvas() {
		const ranked = rankedIndividuals();
		if (!ranked.length) throw new Error("Belum ada data individu untuk poster");

		const width = 1080;
		const height = 1350;
		const canvas = document.createElement("canvas");
		canvas.width = width;
		canvas.height = height;
		const ctx = canvas.getContext("2d");
		if (!ctx) throw new Error("Canvas tidak tersedia");

		const images = await Promise.all(ranked.map((member) => loadImage(member.photo_url)));

		ctx.fillStyle = "#F7F5EF";
		ctx.fillRect(0, 0, width, height);

		ctx.fillStyle = "#155B49";
		ctx.fillRect(0, 0, width, 225);
		ctx.strokeStyle = "rgba(255,255,255,0.09)";
		ctx.lineWidth = 2;
		ctx.beginPath();
		ctx.arc(930, -40, 215, 0, Math.PI * 2);
		ctx.stroke();
		ctx.beginPath();
		ctx.arc(100, 230, 165, 0, Math.PI * 2);
		ctx.stroke();

		ctx.fillStyle = "#D4B46A";
		fillRoundedRect(ctx, 70, 48, 132, 32, 16, "#D4B46A");
		ctx.fillStyle = "#17362E";
		ctx.font = "900 15px Arial, sans-serif";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText("RUANG SANAD", 136, 64);

		ctx.fillStyle = "#FFFFFF";
		ctx.font = "900 48px Arial, sans-serif";
		ctx.textAlign = "left";
		ctx.textBaseline = "alphabetic";
		ctx.fillText("Leaderboard Amaliyah", 70, 135);
		ctx.fillStyle = "#DCECE6";
		ctx.font = "700 23px Arial, sans-serif";
		ctx.fillText(formatPeriod(period_month), 70, 176);
		ctx.fillStyle = "rgba(255,255,255,0.72)";
		ctx.font = "500 16px Arial, sans-serif";
		ctx.fillText("Top individu seluruh tim · maksimal 10 anggota", 70, 205);

		ctx.fillStyle = "#D4B46A";
		ctx.fillRect(70, 242, 940, 3);

		const top = ranked.slice(0, 3);
		const topPositions = [
			{ x: 540, y: 410, radius: 112, labelY: 565, featured: true },
			{ x: 260, y: 440, radius: 86, labelY: 560, featured: false },
			{ x: 820, y: 440, radius: 86, labelY: 560, featured: false },
		];
		for (let i = 0; i < top.length; i++) {
			const member = top[i];
			const pos = topPositions[i];
			drawAvatar(ctx, member, images[i], pos.x, pos.y, pos.radius);
			drawRankBadge(ctx, i + 1, pos.x - pos.radius * 0.82, pos.y - pos.radius * 0.78, pos.featured ? 35 : 29);
			drawMemberLabel(ctx, member, pos.x, pos.labelY, pos.featured ? 280 : 220, pos.featured);
		}

		const lower = ranked.slice(3, 10);
		if (lower.length) {
			fillRoundedRect(ctx, 55, 655, 970, 575, 38, "#FFFFFF");
			ctx.fillStyle = "#6B7E77";
			ctx.font = "800 15px Arial, sans-serif";
			ctx.textAlign = "left";
			ctx.textBaseline = "alphabetic";
			ctx.fillText("PERINGKAT BERIKUTNYA", 90, 705);

			const rowOne = lower.slice(0, 4);
			const rowTwo = lower.slice(4, 7);
			const drawGridRow = (items: AmaliyahMemberSnapshot[], sourceOffset: number, y: number) => {
				const gap = 220;
				const totalWidth = (items.length - 1) * gap;
				const startX = width / 2 - totalWidth / 2;
				items.forEach((member, index) => {
					const x = startX + index * gap;
					const rank = sourceOffset + index + 4;
					const imageIndex = rank - 1;
					drawAvatar(ctx, member, images[imageIndex], x, y, 67);
					drawRankBadge(ctx, rank, x - 55, y - 53, 24);
					drawMemberLabel(ctx, member, x, y + 99, 180, false);
				});
			};

			drawGridRow(rowOne, 0, 815);
			if (rowTwo.length) drawGridRow(rowTwo, 4, 1050);
		}

		ctx.fillStyle = "#155B49";
		ctx.fillRect(0, 1270, width, 80);
		ctx.fillStyle = "#FFFFFF";
		ctx.font = "800 18px Arial, sans-serif";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText("Ruang Sanad Ops", width / 2, 1298);
		ctx.fillStyle = "rgba(255,255,255,0.7)";
		ctx.font = "500 14px Arial, sans-serif";
		ctx.fillText("Overall = rata-rata Tahajud · Tilawah + Tadabur · Dzikir Pagi · Dzikir Petang", width / 2, 1325);

		return canvas;
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
			const canvas = await renderPosterCanvas();
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
		const trailer = `trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${xrefOffset}\n%%EOF`;
		chunks.push(encoder.encode(xref + trailer));
		return concatBytes(chunks);
	}

	async function downloadPDF() {
		if (!leaderboard.members.length || exporting) return;
		exporting = true;
		try {
			const canvas = await renderPosterCanvas();
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
		<section class="overflow-hidden rounded-[30px] bg-gradient-to-br from-brand-700 via-brand-600 to-brand-500 p-6 text-white shadow-lg shadow-brand-900/10 sm:p-8">
			<div class="flex flex-col gap-6 xl:flex-row xl:items-end xl:justify-between">
				<div class="max-w-3xl">
					<div class="flex items-center gap-2 text-sm font-black text-white/80"><Trophy size="19" /> Leaderboard Amaliyah Tim</div>
					<h1 class="mt-3 text-3xl font-black tracking-tight sm:text-4xl">Siapa tim paling konsisten bulan ini?</h1>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-white/80">Empat amaliyah yang dinilai: Tahajud, Tilawah + Tadabur, Dzikir Pagi, dan Dzikir Petang. Data diambil langsung dari workspace Google Sheets tiap anggota.</p>
				</div>
				<div class="grid min-w-[280px] gap-3 rounded-3xl bg-white/10 p-4 backdrop-blur-sm ring-1 ring-white/15">
					<div><div class="text-xs font-bold uppercase tracking-[0.12em] text-white/60">Periode aktif</div><div class="mt-1 text-2xl font-black capitalize">{formatPeriod(period_month)}</div><div class="mt-1 text-xs text-white/70">{leaderboard.members.length} snapshot anggota tersimpan</div></div>
					{#if leaderboard.members.length > 0}
						<div class="grid grid-cols-2 gap-2">
							<button type="button" onclick={downloadPNG} disabled={exporting} class="flex min-h-10 items-center justify-center gap-2 rounded-full bg-white px-3 text-xs font-black text-brand-800 disabled:opacity-50"><Download size="15" /> PNG</button>
							<button type="button" onclick={downloadPDF} disabled={exporting} class="flex min-h-10 items-center justify-center gap-2 rounded-full bg-white/15 px-3 text-xs font-black text-white ring-1 ring-white/20 disabled:opacity-50"><FileDown size="15" /> PDF</button>
						</div>
					{/if}
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
				<div class="flex items-start gap-3"><div class="grid h-11 w-11 shrink-0 place-items-center rounded-2xl bg-brand-100 text-brand-700 dark:bg-brand-950/60 dark:text-brand-300"><Link2 size="21" /></div><div><div class="text-sm font-black text-brand-700 dark:text-brand-300">Google Sheets Sync</div><h2 class="mt-1 text-xl font-black">Hitung semua tim tanpa upload file</h2><p class="mt-1 max-w-2xl text-sm leading-6 text-neutral-500">Aplikasi memakai link workspace yang sudah tersimpan pada setiap anggota. Cukup pilih bulan lalu sinkronkan. Link workspace tidak perlu diinput ulang.</p></div></div>
				<div class="mt-5 grid gap-4 sm:grid-cols-[220px_1fr] sm:items-end">
					<label class="grid gap-1.5 text-xs font-bold text-neutral-500">Bulan yang dihitung<input bind:value={selectedPeriod} type="month" class="min-h-12 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold text-neutral-900 outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/20 dark:border-white/[0.07] dark:bg-neutral-900 dark:text-white" /></label>
					<div class="grid gap-2"><div class="text-xs font-bold text-neutral-500">Workspace terdaftar</div><div class="flex min-h-12 items-center justify-between gap-3 rounded-2xl bg-neutral-50 px-4 ring-1 ring-neutral-200 dark:bg-white/[0.03] dark:ring-white/[0.07]"><div><div class="text-sm font-black">{members.length} anggota</div><div class="text-[11px] text-neutral-500">Semua workspace akan diproses untuk bulan yang dipilih</div></div><CloudDownload size="21" class="text-brand-600" /></div></div>
				</div>
				<button type="button" onclick={syncGoogleSheets} disabled={syncing || !selectedPeriod || members.length === 0} class="mt-5 inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-brand-600 px-7 text-sm font-black text-white shadow-sm transition hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-45">{#if syncing}<RefreshCw size="17" class="animate-spin" /> Sinkronisasi...{:else}<RefreshCw size="17" /> Sync & Hitung Semua Tim{/if}</button>
			</div>

			<div class="rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
				<div class="flex items-center gap-2 text-sm font-black text-brand-700 dark:text-brand-300"><CalendarDays size="18" /> Tracking bulanan</div>
				<h2 class="mt-1 text-xl font-black">Lihat bulan lain</h2>
				<p class="mt-1 text-sm text-neutral-500">Snapshot tersimpan per anggota dan per bulan.</p>
				<div class="mt-5 grid gap-3"><input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold outline-none focus:border-brand-400 dark:border-white/[0.07] dark:bg-neutral-900" /><button type="button" onclick={changePeriod} class="min-h-11 rounded-full bg-neutral-900 px-4 text-sm font-black text-white dark:bg-white dark:text-neutral-900">Tampilkan bulan</button></div>
				{#if leaderboard.periods.length > 0}<div class="mt-5 flex flex-wrap gap-2">{#each leaderboard.periods as savedPeriod}<a href={`/app/amaliyah?month=${savedPeriod}`} class="rounded-full px-3 py-1.5 text-xs font-bold {savedPeriod === period_month ? 'bg-brand-100 text-brand-800 dark:bg-brand-950/60 dark:text-brand-200' : 'bg-neutral-100 text-neutral-600 dark:bg-white/[0.05] dark:text-neutral-300'}">{formatPeriod(savedPeriod)}</a>{/each}</div>{/if}
			</div>
		</section>

		<section class="rounded-[30px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
				<div><div class="flex items-center gap-2 text-sm font-black text-brand-700 dark:text-brand-300"><Medal size="18" /> Ranking divisi</div><h2 class="mt-1 text-2xl font-black capitalize">Leaderboard {formatPeriod(period_month)}</h2><p class="mt-1 text-sm text-neutral-500">Overall adalah rata-rata score individu. Persentase empat amaliyah ditampilkan terpisah agar mudah dibandingkan.</p></div>
				<div class="flex items-center gap-2"><div class="rounded-full bg-neutral-100 px-3 py-1.5 text-xs font-black text-neutral-600 dark:bg-white/[0.05] dark:text-neutral-300">{leaderboard.teams.length} tim</div>{#if leaderboard.members.length > 0}<button type="button" onclick={downloadPNG} disabled={exporting} class="grid h-9 w-9 place-items-center rounded-full bg-brand-100 text-brand-700 hover:bg-brand-200 disabled:opacity-50 dark:bg-brand-950/50 dark:text-brand-300" aria-label="Download PNG poster individu"><Download size="16" /></button><button type="button" onclick={downloadPDF} disabled={exporting} class="grid h-9 w-9 place-items-center rounded-full bg-neutral-100 text-neutral-700 hover:bg-neutral-200 disabled:opacity-50 dark:bg-white/[0.05] dark:text-neutral-300" aria-label="Download PDF poster individu"><FileDown size="16" /></button>{/if}</div>
			</div>

			{#if leaderboard.teams.length === 0}
				<div class="mt-6 grid min-h-56 place-items-center rounded-3xl border border-dashed border-neutral-200 bg-neutral-50 p-8 text-center dark:border-white/[0.08] dark:bg-white/[0.02]"><div><Trophy size="36" class="mx-auto text-neutral-300" /><div class="mt-3 text-base font-black">Belum ada data untuk bulan ini</div><div class="mt-1 text-sm text-neutral-500">Klik “Sync & Hitung Semua Tim” untuk mengambil data langsung dari Google Sheets.</div></div></div>
			{:else}
				<div class="mt-6 grid gap-4">
					{#each leaderboard.teams as team}
						<article class="grid gap-5 rounded-3xl border p-5 {team.rank === 1 ? 'border-amber-300 bg-amber-50/55 dark:border-amber-800 dark:bg-amber-950/15' : 'border-neutral-200 bg-neutral-50/70 dark:border-white/[0.07] dark:bg-white/[0.025]'} lg:grid-cols-[180px_minmax(0,1fr)] lg:items-center">
							<div class="flex items-center gap-4 lg:block"><div class="grid h-14 w-14 place-items-center rounded-2xl bg-white text-2xl shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.07]">{rankLabel(team.rank)}</div><div class="lg:mt-3"><div class="text-lg font-black">{team.division}</div><div class="mt-1 flex items-center gap-1 text-xs text-neutral-500"><UsersRound size="14" />{team.member_count} anggota</div>{#if team.rank === 1}<div class="mt-2 inline-flex items-center gap-1 rounded-full bg-amber-100 px-2.5 py-1 text-[10px] font-black uppercase tracking-wide text-amber-800"><Award size="13" /> Kandidat reward</div>{/if}</div></div>
							<div><div class="flex items-end justify-between gap-3"><div class="text-xs font-bold uppercase tracking-[0.12em] text-neutral-400">Overall score</div><div class="text-3xl font-black tracking-tight">{formatPercent(team.overall_score)}%</div></div><div class="mt-3 grid gap-3 sm:grid-cols-2 xl:grid-cols-4">{#each team.metrics as metric}<div class="rounded-2xl bg-white p-3 ring-1 ring-neutral-200 dark:bg-neutral-900 dark:ring-white/[0.06]"><div class="flex items-center justify-between gap-2"><span class="truncate text-xs font-bold text-neutral-600 dark:text-neutral-300">{metric.label}</span><span class="text-xs font-black">{formatPercent(metric.score)}%</span></div><div class="mt-2 h-2 overflow-hidden rounded-full bg-neutral-100 dark:bg-white/[0.06]"><div class="h-full rounded-full bg-brand-500" style={`width: ${scoreWidth(metric.score)}`}></div></div></div>{/each}</div></div>
						</article>
					{/each}
				</div>
			{/if}
		</section>

		<section class="overflow-hidden rounded-[30px] bg-white shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]">
			<div class="flex items-center justify-between gap-4 border-b border-neutral-200 px-5 py-5 dark:border-white/[0.06] sm:px-6"><div><div class="text-sm font-black text-brand-700 dark:text-brand-300">Detail individu</div><h2 class="mt-1 text-xl font-black">Kontribusi anggota</h2></div><div class="text-xs font-bold text-neutral-500">{leaderboard.members.length} data</div></div>
			{#if leaderboard.members.length > 0}<div class="overflow-x-auto"><table class="min-w-[920px] w-full text-left"><thead class="bg-neutral-50 text-[11px] font-black uppercase tracking-wide text-neutral-500 dark:bg-white/[0.025]"><tr><th class="px-6 py-3">Anggota</th><th class="px-4 py-3">Divisi</th><th class="px-4 py-3 text-right">Overall</th><th class="px-4 py-3 text-right">Tahajud</th><th class="px-4 py-3 text-right">Tilawah + Tadabur</th><th class="px-4 py-3 text-right">Dzikir Pagi</th><th class="px-4 py-3 text-right">Dzikir Petang</th></tr></thead><tbody class="divide-y divide-neutral-100 dark:divide-white/[0.05]">{#each leaderboard.members as member}<tr class="text-sm"><td class="px-6 py-4"><div class="flex items-center gap-3"><div class="grid h-9 w-9 shrink-0 place-items-center overflow-hidden rounded-full bg-emerald-50 text-xs font-black text-emerald-800">{#if member.photo_url}<img src={member.photo_url} alt={`Foto ${member.member_name}`} class="h-full w-full object-cover" />{:else}{initials(member.member_name)}{/if}</div><div><div class="font-black">{member.member_name}</div><div class="mt-0.5 max-w-64 truncate text-[10px] text-neutral-400" title={member.source_file_name}>{member.source_file_name}</div></div></div></td><td class="px-4 py-4 text-neutral-500">{member.division}</td><td class="px-4 py-4 text-right text-base font-black">{formatPercent(member.overall_score)}%</td>{#each metricKeys as key}<td class="px-4 py-4 text-right"><div class="font-bold">{formatPercent(memberMetric(member, key))}%</div><div class="mt-0.5 text-[10px] text-neutral-400">{memberMetricAvailable(member, key)} checkbox</div></td>{/each}</tr>{/each}</tbody></table></div>{:else}<div class="px-6 py-10 text-center text-sm text-neutral-500">Belum ada snapshot anggota pada periode ini.</div>{/if}
		</section>

		<details class="group rounded-[28px] bg-white p-5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06] sm:p-6">
			<summary class="cursor-pointer list-none"><div class="flex items-center justify-between gap-3"><div><div class="flex items-center gap-2 text-sm font-black text-neutral-600 dark:text-neutral-300"><UploadCloud size="18" /> Fallback manual</div><div class="mt-1 text-base font-black">Upload XLSX hanya jika Google Sheets gagal</div></div><FileSpreadsheet size="20" class="text-neutral-400" /></div></summary>
			<div class="mt-5 grid gap-4 border-t border-neutral-100 pt-5 dark:border-white/[0.05]"><div class="grid gap-4 sm:grid-cols-[220px_1fr]"><label class="grid gap-1.5 text-xs font-bold text-neutral-500">Bulan<input bind:value={selectedPeriod} type="month" class="min-h-11 rounded-2xl border border-neutral-200 bg-neutral-50 px-3 text-sm font-bold outline-none dark:border-white/[0.07] dark:bg-neutral-900" /></label><label class="grid cursor-pointer gap-1.5 text-xs font-bold text-neutral-500">Workspace anggota<div class="flex min-h-11 items-center gap-3 rounded-2xl border border-dashed border-neutral-300 px-4 text-sm font-bold"><FileSpreadsheet size="18" /><span>{uploadRows.length ? `${uploadRows.length} file dipilih` : "Pilih file .xlsx"}</span></div><input type="file" accept=".xlsx,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" multiple class="hidden" onchange={handleFiles} /></label></div>{#if uploadRows.length > 0}<div class="grid gap-2">{#each uploadRows as row, index}<div class="grid gap-3 rounded-2xl bg-neutral-50 p-3 dark:bg-white/[0.03] sm:grid-cols-[minmax(0,1fr)_260px] sm:items-center"><div class="min-w-0"><div class="truncate text-sm font-black">{row.file.name}</div><div class="mt-0.5 text-[11px] text-neutral-500">{(row.file.size / 1024).toFixed(0)} KB</div></div><select bind:value={uploadRows[index].member_id} class="min-h-10 rounded-xl border border-neutral-200 bg-white px-3 text-sm font-bold outline-none dark:border-white/[0.08] dark:bg-neutral-900"><option value={0}>Pilih anggota...</option>{#each members as member}<option value={member.id}>{member.name} · {member.division}</option>{/each}</select></div>{/each}</div><button type="button" onclick={importWorkspaces} disabled={importing || uploadRows.some((row) => !row.member_id)} class="inline-flex min-h-11 w-fit items-center justify-center gap-2 rounded-full bg-neutral-900 px-5 text-sm font-black text-white disabled:opacity-45 dark:bg-white dark:text-neutral-900">{#if importing}<RefreshCw size="16" class="animate-spin" /> Menghitung...{:else}<UploadCloud size="16" /> Hitung upload manual{/if}</button>{/if}</div>
		</details>
	</div>
</AppLayout>
