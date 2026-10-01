export interface PosterMember {
	member_name: string;
	division: string;
	photo_url: string;
	overall_score: number;
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

function fillRound(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number, fill: string) {
	ctx.save();
	ctx.fillStyle = fill;
	roundedRectPath(ctx, x, y, width, height, radius);
	ctx.fill();
	ctx.restore();
}

function strokeRound(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number, stroke: string, lineWidth = 2) {
	ctx.save();
	ctx.strokeStyle = stroke;
	ctx.lineWidth = lineWidth;
	roundedRectPath(ctx, x, y, width, height, radius);
	ctx.stroke();
	ctx.restore();
}

function initials(name: string) {
	return name.trim().split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase() ?? "").join("") || "?";
}

function fitText(ctx: CanvasRenderingContext2D, value: string, maxWidth: number) {
	if (ctx.measureText(value).width <= maxWidth) return value;
	let result = value;
	while (result.length > 1 && ctx.measureText(`${result}…`).width > maxWidth) result = result.slice(0, -1);
	return `${result}…`;
}

function loadImage(src: string) {
	return new Promise<HTMLImageElement | null>((resolve) => {
		if (!src) return resolve(null);
		const image = new Image();
		image.onload = () => resolve(image);
		image.onerror = () => resolve(null);
		image.src = src;
	});
}

function drawCover(ctx: CanvasRenderingContext2D, image: HTMLImageElement, cx: number, cy: number, radius: number) {
	const size = radius * 2;
	const scale = Math.max(size / image.naturalWidth, size / image.naturalHeight);
	const sw = size / scale;
	const sh = size / scale;
	const sx = (image.naturalWidth - sw) / 2;
	const sy = (image.naturalHeight - sh) / 2;
	ctx.drawImage(image, sx, sy, sw, sh, cx - radius, cy - radius, size, size);
}

function drawAvatar(ctx: CanvasRenderingContext2D, member: PosterMember, image: HTMLImageElement | null, cx: number, cy: number, radius: number, ring = "#D6B767") {
	ctx.save();
	ctx.shadowColor = "rgba(8, 47, 36, .18)";
	ctx.shadowBlur = 22;
	ctx.shadowOffsetY = 8;
	ctx.fillStyle = "#FFFFFF";
	ctx.beginPath();
	ctx.arc(cx, cy, radius + 10, 0, Math.PI * 2);
	ctx.fill();
	ctx.restore();

	ctx.save();
	ctx.beginPath();
	ctx.arc(cx, cy, radius, 0, Math.PI * 2);
	ctx.clip();
	if (image) {
		drawCover(ctx, image, cx, cy, radius);
	} else {
		const gradient = ctx.createLinearGradient(cx - radius, cy - radius, cx + radius, cy + radius);
		gradient.addColorStop(0, "#E5F1EC");
		gradient.addColorStop(1, "#C7E0D6");
		ctx.fillStyle = gradient;
		ctx.fillRect(cx - radius, cy - radius, radius * 2, radius * 2);
		ctx.fillStyle = "#0D5B48";
		ctx.font = `900 ${Math.round(radius * .62)}px Arial, sans-serif`;
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(initials(member.member_name), cx, cy + 2);
	}
	ctx.restore();

	ctx.save();
	ctx.strokeStyle = ring;
	ctx.lineWidth = 5;
	ctx.beginPath();
	ctx.arc(cx, cy, radius + 5, 0, Math.PI * 2);
	ctx.stroke();
	ctx.restore();
}

function drawRankMedal(ctx: CanvasRenderingContext2D, rank: number, x: number, y: number, radius: number) {
	const colors = rank === 1 ? ["#F7D882", "#B98220"] : rank === 2 ? ["#E8EDF1", "#8D9AA5"] : rank === 3 ? ["#E9B58F", "#A76038"] : ["#5BA58B", "#176B55"];
	const gradient = ctx.createLinearGradient(x - radius, y - radius, x + radius, y + radius);
	gradient.addColorStop(0, colors[0]);
	gradient.addColorStop(1, colors[1]);
	ctx.save();
	ctx.shadowColor = "rgba(0,0,0,.16)";
	ctx.shadowBlur = 10;
	ctx.fillStyle = gradient;
	ctx.beginPath();
	ctx.arc(x, y, radius, 0, Math.PI * 2);
	ctx.fill();
	ctx.strokeStyle = "#FFFFFF";
	ctx.lineWidth = 4;
	ctx.stroke();
	ctx.fillStyle = "#FFFFFF";
	ctx.font = `900 ${Math.round(radius * 1.05)}px Arial, sans-serif`;
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(String(rank), x, y + 1);
	ctx.restore();
}

function drawStar(ctx: CanvasRenderingContext2D, x: number, y: number, radius: number, color: string, alpha = 1) {
	ctx.save();
	ctx.translate(x, y);
	ctx.rotate(Math.PI / 4);
	ctx.globalAlpha = alpha;
	ctx.strokeStyle = color;
	ctx.lineWidth = 2;
	ctx.strokeRect(-radius, -radius, radius * 2, radius * 2);
	ctx.rotate(Math.PI / 4);
	ctx.strokeRect(-radius, -radius, radius * 2, radius * 2);
	ctx.restore();
}

function drawCrown(ctx: CanvasRenderingContext2D, x: number, y: number) {
	ctx.save();
	ctx.fillStyle = "#D7B45F";
	ctx.beginPath();
	ctx.moveTo(x - 35, y + 14);
	ctx.lineTo(x - 26, y - 18);
	ctx.lineTo(x - 8, y + 2);
	ctx.lineTo(x, y - 30);
	ctx.lineTo(x + 10, y + 2);
	ctx.lineTo(x + 30, y - 18);
	ctx.lineTo(x + 35, y + 14);
	ctx.closePath();
	ctx.fill();
	fillRound(ctx, x - 36, y + 11, 72, 11, 5, "#C69631");
	ctx.restore();
}

function drawPodium(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, fill: string, rank: number) {
	ctx.save();
	ctx.shadowColor = "rgba(30, 73, 59, .13)";
	ctx.shadowBlur = 22;
	ctx.shadowOffsetY = 10;
	fillRound(ctx, x, y, width, height, 28, fill);
	ctx.restore();
	strokeRound(ctx, x, y, width, height, 28, "rgba(212,180,106,.75)", 2);
	ctx.fillStyle = rank === 1 ? "#9C6B16" : "#176B55";
	ctx.font = "900 20px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(`#${rank}`, x + width / 2, y + height - 25);
}

function drawNameAndScore(ctx: CanvasRenderingContext2D, member: PosterMember, x: number, y: number, width: number, scoreFormatter: (value: number) => string, featured = false) {
	ctx.textAlign = "center";
	ctx.textBaseline = "alphabetic";
	ctx.fillStyle = "#16382E";
	ctx.font = `${featured ? 900 : 800} ${featured ? 31 : 23}px Arial, sans-serif`;
	ctx.fillText(fitText(ctx, member.member_name, width), x, y);
	ctx.fillStyle = "#6F827A";
	ctx.font = `${featured ? 600 : 500} ${featured ? 15 : 13}px Arial, sans-serif`;
	ctx.fillText(fitText(ctx, member.division, width), x, y + 24);
	const scoreText = `${scoreFormatter(member.overall_score)}%`;
	ctx.font = `900 ${featured ? 20 : 16}px Arial, sans-serif`;
	const pillWidth = ctx.measureText(scoreText).width + (featured ? 38 : 30);
	fillRound(ctx, x - pillWidth / 2, y + 38, pillWidth, featured ? 38 : 32, 18, featured ? "#F5DE9C" : "#E4F0EB");
	ctx.fillStyle = featured ? "#815B15" : "#145B49";
	ctx.textBaseline = "middle";
	ctx.fillText(scoreText, x, y + (featured ? 57 : 54));
}

function drawMetricChips(ctx: CanvasRenderingContext2D, y: number) {
	const labels = ["Tahajud", "Tilawah + Tadabur", "Dzikir Pagi", "Dzikir Petang"];
	ctx.font = "700 14px Arial, sans-serif";
	const widths = labels.map((label) => ctx.measureText(label).width + 30);
	const gap = 12;
	const total = widths.reduce((sum, value) => sum + value, 0) + gap * (labels.length - 1);
	let x = 540 - total / 2;
	for (let i = 0; i < labels.length; i++) {
		fillRound(ctx, x, y, widths[i], 34, 17, i % 2 === 0 ? "rgba(255,255,255,.11)" : "rgba(215,180,95,.16)");
		ctx.fillStyle = i % 2 === 0 ? "#E6F1EC" : "#F1DFAE";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(labels[i], x + widths[i] / 2, y + 17);
		x += widths[i] + gap;
	}
}

export async function renderAmaliyahPosterCanvas(
	members: PosterMember[],
	periodLabel: string,
	scoreFormatter: (value: number) => string,
) {
	const ranked = [...members]
		.sort((a, b) => b.overall_score - a.overall_score || a.member_name.localeCompare(b.member_name, "id"))
		.slice(0, 10);
	if (!ranked.length) throw new Error("Belum ada data individu untuk poster");

	const width = 1080;
	const height = 1350;
	const canvas = document.createElement("canvas");
	canvas.width = width;
	canvas.height = height;
	const ctx = canvas.getContext("2d");
	if (!ctx) throw new Error("Canvas tidak tersedia");
	const photos = await Promise.all(ranked.map((member) => loadImage(member.photo_url)));

	const bg = ctx.createLinearGradient(0, 0, 1080, 1350);
	bg.addColorStop(0, "#073D31");
	bg.addColorStop(.52, "#0D5A47");
	bg.addColorStop(1, "#124B3D");
	ctx.fillStyle = bg;
	ctx.fillRect(0, 0, width, height);

	for (let x = 40; x <= width; x += 92) {
		for (let y = 36; y < height; y += 92) drawStar(ctx, x, y, 18, "#D7B45F", .045);
	}

	const glow = ctx.createRadialGradient(540, 510, 40, 540, 510, 500);
	glow.addColorStop(0, "rgba(239,220,163,.22)");
	glow.addColorStop(1, "rgba(239,220,163,0)");
	ctx.fillStyle = glow;
	ctx.fillRect(40, 220, 1000, 780);

	fillRound(ctx, 54, 42, 972, 202, 38, "rgba(3, 46, 36, .70)");
	strokeRound(ctx, 54, 42, 972, 202, 38, "rgba(215,180,95,.42)", 2);
	fillRound(ctx, 78, 67, 160, 38, 19, "#E1C371");
	ctx.fillStyle = "#0A4134";
	ctx.font = "900 16px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText("RUANG SANAD", 158, 86);
	ctx.fillStyle = "#FFFFFF";
	ctx.font = "900 55px Arial, sans-serif";
	ctx.textAlign = "left";
	ctx.textBaseline = "alphabetic";
	ctx.fillText("Leaderboard Amaliyah", 78, 162);
	fillRound(ctx, 78, 180, 235, 39, 19, "rgba(255,255,255,.10)");
	ctx.fillStyle = "#E8F2EE";
	ctx.font = "800 19px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(periodLabel, 195, 199);
	drawMetricChips(ctx, 184);

	fillRound(ctx, 54, 268, 972, 950, 44, "#F6F3EA");
	strokeRound(ctx, 54, 268, 972, 950, 44, "rgba(215,180,95,.65)", 2);
	ctx.fillStyle = "#D7B45F";
	ctx.fillRect(108, 294, 864, 3);

	const top = ranked.slice(0, 3);
	const positions = [
		{ x: 540, y: 448, radius: 118, nameY: 606, featured: true, podiumX: 410, podiumY: 690, podiumW: 260, podiumH: 150, fill: "#FFF8E5" },
		{ x: 270, y: 500, radius: 89, nameY: 626, featured: false, podiumX: 160, podiumY: 735, podiumW: 220, podiumH: 105, fill: "#EEF2F1" },
		{ x: 810, y: 500, radius: 89, nameY: 626, featured: false, podiumX: 700, podiumY: 735, podiumW: 220, podiumH: 105, fill: "#F3E8E1" },
	];
	for (let index = 0; index < top.length; index++) {
		const member = top[index];
		const pos = positions[index];
		if (index === 0) drawCrown(ctx, pos.x, pos.y - pos.radius - 48);
		drawAvatar(ctx, member, photos[index], pos.x, pos.y, pos.radius, index === 0 ? "#D6B767" : index === 1 ? "#9AA7B0" : "#B97950");
		drawRankMedal(ctx, index + 1, pos.x - pos.radius * .82, pos.y - pos.radius * .74, index === 0 ? 37 : 30);
		drawNameAndScore(ctx, member, pos.x, pos.nameY, pos.featured ? 300 : 220, scoreFormatter, pos.featured);
		drawPodium(ctx, pos.podiumX, pos.podiumY, pos.podiumW, pos.podiumH, pos.fill, index + 1);
	}

	const remaining = ranked.slice(3);
	if (remaining.length) {
		ctx.fillStyle = "#61756D";
		ctx.font = "900 14px Arial, sans-serif";
		ctx.textAlign = "left";
		ctx.textBaseline = "alphabetic";
		ctx.fillText("TOP 4 - 10", 105, 900);
		ctx.fillStyle = "#D7B45F";
		ctx.fillRect(105, 913, 92, 3);
		const colWidth = 420;
		const cardH = 86;
		const gapY = 14;
		remaining.forEach((member, index) => {
			const col = index % 2;
			const row = Math.floor(index / 2);
			const x = 105 + col * 455;
			const y = 935 + row * (cardH + gapY);
			fillRound(ctx, x, y, colWidth, cardH, 22, "#FFFFFF");
			strokeRound(ctx, x, y, colWidth, cardH, 22, "#E4DED0", 1.5);
			const rank = index + 4;
			drawAvatar(ctx, member, photos[index + 3], x + 51, y + 43, 27, "#D7B45F");
			drawRankMedal(ctx, rank, x + 26, y + 24, 16);
			ctx.fillStyle = "#173A30";
			ctx.font = "800 17px Arial, sans-serif";
			ctx.textAlign = "left";
			ctx.textBaseline = "alphabetic";
			ctx.fillText(fitText(ctx, member.member_name, 210), x + 92, y + 36);
			ctx.fillStyle = "#7B8B84";
			ctx.font = "500 12px Arial, sans-serif";
			ctx.fillText(fitText(ctx, member.division, 210), x + 92, y + 58);
			fillRound(ctx, x + 330, y + 24, 70, 38, 19, rank <= 5 ? "#F3E2AD" : "#E5F0EB");
			ctx.fillStyle = rank <= 5 ? "#7F5A13" : "#155B49";
			ctx.font = "900 15px Arial, sans-serif";
			ctx.textAlign = "center";
			ctx.textBaseline = "middle";
			ctx.fillText(`${scoreFormatter(member.overall_score)}%`, x + 365, y + 43);
		});
	} else {
		fillRound(ctx, 135, 900, 810, 180, 32, "#EAF2EE");
		ctx.fillStyle = "#155B49";
		ctx.font = "900 23px Arial, sans-serif";
		ctx.textAlign = "center";
		ctx.fillText("Konsistensi kecil yang dijaga setiap hari,", 540, 967);
		ctx.fillText("akan menjadi kebiasaan besar.", 540, 1001);
		ctx.fillStyle = "#71827B";
		ctx.font = "600 14px Arial, sans-serif";
		ctx.fillText("Leaderboard akan otomatis terisi hingga Top 10 saat anggota bertambah.", 540, 1040);
	}

	fillRound(ctx, 54, 1240, 972, 70, 30, "rgba(3,46,36,.72)");
	ctx.fillStyle = "#FFFFFF";
	ctx.font = "900 18px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText("Ruang Sanad Ops", 540, 1265);
	ctx.fillStyle = "rgba(255,255,255,.72)";
	ctx.font = "500 13px Arial, sans-serif";
	ctx.fillText("Overall = rata-rata Tahajud · Tilawah + Tadabur · Dzikir Pagi · Dzikir Petang", 540, 1293);

	return canvas;
}
