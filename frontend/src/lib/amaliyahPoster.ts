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

function fillRound(ctx: CanvasRenderingContext2D, x: number, y: number, width: number, height: number, radius: number, fill: string | CanvasGradient) {
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
	ctx.shadowColor = "rgba(8, 47, 36, .22)";
	ctx.shadowBlur = 24;
	ctx.shadowOffsetY = 8;
	ctx.fillStyle = "#FFFFFF";
	ctx.beginPath();
	ctx.arc(cx, cy, radius + 11, 0, Math.PI * 2);
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
		gradient.addColorStop(0, "#EAF4F0");
		gradient.addColorStop(1, "#C8E2D8");
		ctx.fillStyle = gradient;
		ctx.fillRect(cx - radius, cy - radius, radius * 2, radius * 2);
		ctx.fillStyle = "#0B5946";
		ctx.font = `900 ${Math.round(radius * .62)}px Arial, sans-serif`;
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(initials(member.member_name), cx, cy + 2);
	}
	ctx.restore();

	ctx.save();
	ctx.strokeStyle = ring;
	ctx.lineWidth = 6;
	ctx.beginPath();
	ctx.arc(cx, cy, radius + 5, 0, Math.PI * 2);
	ctx.stroke();
	ctx.strokeStyle = "rgba(255,255,255,.95)";
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.arc(cx, cy, radius - 2, 0, Math.PI * 2);
	ctx.stroke();
	ctx.restore();
}

function drawRankMedal(ctx: CanvasRenderingContext2D, rank: number, x: number, y: number, radius: number) {
	const colors = rank === 1 ? ["#FFE69A", "#B97D13"] : rank === 2 ? ["#EDF2F5", "#8B99A5"] : rank === 3 ? ["#F0BB94", "#A65B34"] : ["#6BB297", "#176B55"];
	const gradient = ctx.createLinearGradient(x - radius, y - radius, x + radius, y + radius);
	gradient.addColorStop(0, colors[0]);
	gradient.addColorStop(1, colors[1]);
	ctx.save();
	ctx.shadowColor = "rgba(0,0,0,.16)";
	ctx.shadowBlur = 12;
	ctx.fillStyle = gradient;
	ctx.beginPath();
	ctx.arc(x, y, radius, 0, Math.PI * 2);
	ctx.fill();
	ctx.strokeStyle = "#FFFFFF";
	ctx.lineWidth = 4;
	ctx.stroke();
	ctx.fillStyle = "#FFFFFF";
	ctx.font = `900 ${Math.round(radius * 1.02)}px Arial, sans-serif`;
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

function drawDiamond(ctx: CanvasRenderingContext2D, x: number, y: number, size: number, color: string, alpha = 1) {
	ctx.save();
	ctx.translate(x, y);
	ctx.rotate(Math.PI / 4);
	ctx.globalAlpha = alpha;
	ctx.fillStyle = color;
	ctx.fillRect(-size / 2, -size / 2, size, size);
	ctx.restore();
}

function drawCornerOrnament(ctx: CanvasRenderingContext2D, x: number, y: number, flipX: boolean, flipY: boolean) {
	ctx.save();
	ctx.translate(x, y);
	ctx.scale(flipX ? -1 : 1, flipY ? -1 : 1);
	ctx.strokeStyle = "rgba(225,195,113,.44)";
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.moveTo(0, 74);
	ctx.quadraticCurveTo(0, 0, 74, 0);
	ctx.stroke();
	ctx.beginPath();
	ctx.moveTo(10, 68);
	ctx.quadraticCurveTo(10, 10, 68, 10);
	ctx.stroke();
	drawDiamond(ctx, 36, 10, 8, "#E1C371", .7);
	drawDiamond(ctx, 10, 36, 8, "#E1C371", .7);
	ctx.restore();
}

function drawLantern(ctx: CanvasRenderingContext2D, x: number, y: number, scale = 1) {
	ctx.save();
	ctx.translate(x, y);
	ctx.scale(scale, scale);
	ctx.strokeStyle = "rgba(239,211,132,.82)";
	ctx.fillStyle = "rgba(239,211,132,.14)";
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.moveTo(0, -34);
	ctx.lineTo(0, -15);
	ctx.stroke();
	ctx.beginPath();
	ctx.moveTo(-14, -15);
	ctx.lineTo(14, -15);
	ctx.lineTo(10, 10);
	ctx.lineTo(-10, 10);
	ctx.closePath();
	ctx.fill();
	ctx.stroke();
	ctx.beginPath();
	ctx.moveTo(-8, 10);
	ctx.lineTo(0, 22);
	ctx.lineTo(8, 10);
	ctx.stroke();
	ctx.beginPath();
	ctx.arc(0, -19, 4, 0, Math.PI * 2);
	ctx.fillStyle = "#F2D483";
	ctx.fill();
	ctx.restore();
}

function drawCrown(ctx: CanvasRenderingContext2D, x: number, y: number) {
	ctx.save();
	const gradient = ctx.createLinearGradient(x - 36, y - 30, x + 36, y + 20);
	gradient.addColorStop(0, "#F3D77F");
	gradient.addColorStop(1, "#B77A17");
	ctx.fillStyle = gradient;
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

function drawLaurel(ctx: CanvasRenderingContext2D, cx: number, cy: number, radius: number, color: string) {
	ctx.save();
	ctx.strokeStyle = color;
	ctx.fillStyle = color;
	ctx.lineWidth = 2;
	for (const side of [-1, 1]) {
		ctx.beginPath();
		ctx.arc(cx, cy + 5, radius + 30, side < 0 ? Math.PI * .62 : Math.PI * 1.38, side < 0 ? Math.PI * 1.02 : Math.PI * 1.98, side < 0);
		ctx.stroke();
		for (let i = 0; i < 6; i++) {
			const angle = side < 0 ? Math.PI * (.66 + i * .055) : Math.PI * (1.34 - i * .055);
			const x = cx + Math.cos(angle) * (radius + 30);
			const y = cy + 5 + Math.sin(angle) * (radius + 30);
			ctx.save();
			ctx.translate(x, y);
			ctx.rotate(angle + Math.PI / 2);
			ctx.beginPath();
			ctx.ellipse(0, 0, 4, 10, 0, 0, Math.PI * 2);
			ctx.fill();
			ctx.restore();
		}
	}
	ctx.restore();
}

function drawDivider(ctx: CanvasRenderingContext2D, y: number, label: string) {
	ctx.save();
	ctx.strokeStyle = "rgba(202,157,61,.62)";
	ctx.lineWidth = 2;
	ctx.beginPath();
	ctx.moveTo(120, y);
	ctx.lineTo(415, y);
	ctx.moveTo(665, y);
	ctx.lineTo(960, y);
	ctx.stroke();
	drawDiamond(ctx, 437, y, 8, "#C99C3D", .85);
	drawDiamond(ctx, 643, y, 8, "#C99C3D", .85);
	fillRound(ctx, 447, y - 21, 186, 42, 21, "#0D5B48");
	strokeRound(ctx, 447, y - 21, 186, 42, 21, "rgba(215,180,95,.8)", 1.5);
	ctx.fillStyle = "#F7E9BE";
	ctx.font = "900 15px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(label, 540, y + 1);
	ctx.restore();
}

function drawMetricChips(ctx: CanvasRenderingContext2D, y: number) {
	const labels = ["Tahajud", "Tilawah + Tadabur", "Dzikir Pagi", "Dzikir Petang"];
	ctx.font = "700 13px Arial, sans-serif";
	const widths = labels.map((label) => ctx.measureText(label).width + 34);
	const gap = 10;
	const total = widths.reduce((sum, value) => sum + value, 0) + gap * (labels.length - 1);
	let x = 540 - total / 2;
	for (let i = 0; i < labels.length; i++) {
		const fill = i % 2 === 0 ? "rgba(255,255,255,.10)" : "rgba(215,180,95,.16)";
		fillRound(ctx, x, y, widths[i], 34, 17, fill);
		strokeRound(ctx, x, y, widths[i], 34, 17, "rgba(255,255,255,.08)", 1);
		drawDiamond(ctx, x + 14, y + 17, 6, i % 2 === 0 ? "#B9D7CC" : "#DDBF72", .85);
		ctx.fillStyle = i % 2 === 0 ? "#E6F1EC" : "#F1DFAE";
		ctx.textAlign = "center";
		ctx.textBaseline = "middle";
		ctx.fillText(labels[i], x + widths[i] / 2 + 5, y + 17);
		x += widths[i] + gap;
	}
}

function drawPodiumCard(
	ctx: CanvasRenderingContext2D,
	member: PosterMember,
	x: number,
	y: number,
	width: number,
	height: number,
	rank: number,
	scoreFormatter: (value: number) => string,
) {
	const gradients = [
		["#FFF9E7", "#F5E6B7"],
		["#F4F7F6", "#E7EFEC"],
		["#F8EEE7", "#EFDACE"],
	];
	const pair = gradients[Math.min(rank - 1, 2)];
	const gradient = ctx.createLinearGradient(x, y, x, y + height);
	gradient.addColorStop(0, pair[0]);
	gradient.addColorStop(1, pair[1]);
	ctx.save();
	ctx.shadowColor = "rgba(22,58,46,.12)";
	ctx.shadowBlur = 24;
	ctx.shadowOffsetY = 10;
	fillRound(ctx, x, y, width, height, 28, gradient);
	ctx.restore();
	strokeRound(ctx, x, y, width, height, 28, rank === 1 ? "rgba(198,150,49,.78)" : "rgba(197,177,129,.68)", 2);

	ctx.fillStyle = rank === 1 ? "#8A6114" : "#173A30";
	ctx.font = `${rank === 1 ? 900 : 800} ${rank === 1 ? 28 : 24}px Georgia, serif`;
	ctx.textAlign = "center";
	ctx.textBaseline = "alphabetic";
	ctx.fillText(fitText(ctx, member.member_name, width - 42), x + width / 2, y + (rank === 1 ? 96 : 86));
	ctx.fillStyle = "#6F827A";
	ctx.font = `${rank === 1 ? 600 : 500} ${rank === 1 ? 15 : 13}px Arial, sans-serif`;
	ctx.fillText(fitText(ctx, member.division, width - 50), x + width / 2, y + (rank === 1 ? 122 : 110));

	const scoreText = `${scoreFormatter(member.overall_score)}%`;
	ctx.font = `900 ${rank === 1 ? 21 : 17}px Arial, sans-serif`;
	const pillWidth = ctx.measureText(scoreText).width + (rank === 1 ? 42 : 34);
	fillRound(ctx, x + width / 2 - pillWidth / 2, y + height - (rank === 1 ? 62 : 54), pillWidth, rank === 1 ? 40 : 34, 20, rank === 1 ? "#EBCB72" : "#DDEDE7");
	ctx.fillStyle = rank === 1 ? "#714B0B" : "#145B49";
	ctx.textBaseline = "middle";
	ctx.fillText(scoreText, x + width / 2, y + height - (rank === 1 ? 42 : 37));

	ctx.fillStyle = rank === 1 ? "rgba(156,107,22,.08)" : "rgba(23,107,85,.07)";
	for (let i = 0; i < 4; i++) drawDiamond(ctx, x + 44 + i * 58, y + height - 18, 5, rank === 1 ? "#B98220" : "#176B55", .28);
}

function drawListCard(
	ctx: CanvasRenderingContext2D,
	member: PosterMember,
	image: HTMLImageElement | null,
	rank: number,
	x: number,
	y: number,
	width: number,
	height: number,
	scoreFormatter: (value: number) => string,
) {
	const cardFill = rank <= 5 ? "#FFFDF8" : "#FFFFFF";
	ctx.save();
	ctx.shadowColor = "rgba(19,58,47,.08)";
	ctx.shadowBlur = 14;
	ctx.shadowOffsetY = 5;
	fillRound(ctx, x, y, width, height, 20, cardFill);
	ctx.restore();
	strokeRound(ctx, x, y, width, height, 20, rank <= 5 ? "#E5CF8E" : "#E5E0D5", 1.4);

	fillRound(ctx, x + 10, y + 10, 6, height - 20, 3, rank <= 5 ? "#D6B767" : "#8EBCAA");
	drawAvatar(ctx, member, image, x + 51, y + height / 2, 24, rank <= 5 ? "#D6B767" : "#B9D7CC");
	drawRankMedal(ctx, rank, x + 25, y + 22, 15);

	ctx.fillStyle = "#173A30";
	ctx.font = "800 16px Arial, sans-serif";
	ctx.textAlign = "left";
	ctx.textBaseline = "alphabetic";
	ctx.fillText(fitText(ctx, member.member_name, 190), x + 92, y + 31);
	ctx.fillStyle = "#7B8B84";
	ctx.font = "500 11.5px Arial, sans-serif";
	ctx.fillText(fitText(ctx, member.division, 190), x + 92, y + 51);

	const scoreText = `${scoreFormatter(member.overall_score)}%`;
	ctx.font = "900 14px Arial, sans-serif";
	const pillWidth = Math.max(70, ctx.measureText(scoreText).width + 26);
	fillRound(ctx, x + width - pillWidth - 18, y + (height - 34) / 2, pillWidth, 34, 17, rank <= 5 ? "#F3E2AD" : "#E5F0EB");
	ctx.fillStyle = rank <= 5 ? "#7F5A13" : "#155B49";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText(scoreText, x + width - pillWidth / 2 - 18, y + height / 2);
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

	const bg = ctx.createLinearGradient(0, 0, width, height);
	bg.addColorStop(0, "#063D31");
	bg.addColorStop(.48, "#0D5A47");
	bg.addColorStop(1, "#084334");
	ctx.fillStyle = bg;
	ctx.fillRect(0, 0, width, height);

	for (let x = 38; x <= width; x += 90) {
		for (let y = 32; y < height; y += 90) drawStar(ctx, x, y, 17, "#D7B45F", .05);
	}
	for (let i = 0; i < 16; i++) {
		const x = 34 + i * 70;
		drawDiamond(ctx, x, 18, 7, "#E1C371", .12);
		drawDiamond(ctx, x, height - 18, 7, "#E1C371", .12);
	}

	const glow = ctx.createRadialGradient(540, 500, 70, 540, 500, 510);
	glow.addColorStop(0, "rgba(239,220,163,.23)");
	glow.addColorStop(1, "rgba(239,220,163,0)");
	ctx.fillStyle = glow;
	ctx.fillRect(20, 160, 1040, 850);

	fillRound(ctx, 54, 42, 972, 205, 38, "rgba(3,46,36,.76)");
	strokeRound(ctx, 54, 42, 972, 205, 38, "rgba(215,180,95,.48)", 2);
	drawCornerOrnament(ctx, 70, 58, false, false);
	drawCornerOrnament(ctx, 1010, 58, true, false);
	drawLantern(ctx, 120, 118, .95);
	drawLantern(ctx, 960, 118, .95);
	fillRound(ctx, 450, 63, 180, 36, 18, "#E1C371");
	ctx.fillStyle = "#0A4134";
	ctx.font = "900 15px Arial, sans-serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText("RUANG SANAD", 540, 81);
	ctx.fillStyle = "#FFFDF7";
	ctx.font = "900 55px Georgia, serif";
	ctx.textBaseline = "alphabetic";
	ctx.fillText("Leaderboard Amaliyah", 540, 153);
	ctx.fillStyle = "#E8F2EE";
	ctx.font = "800 19px Arial, sans-serif";
	ctx.fillText(periodLabel, 540, 187);
	drawMetricChips(ctx, 201);

	fillRound(ctx, 54, 270, 972, 948, 44, "#F7F3E9");
	strokeRound(ctx, 54, 270, 972, 948, 44, "rgba(215,180,95,.72)", 2);
	drawCornerOrnament(ctx, 78, 292, false, false);
	drawCornerOrnament(ctx, 1002, 292, true, false);
	ctx.fillStyle = "#D7B45F";
	ctx.fillRect(120, 299, 840, 2);

	const archGradient = ctx.createRadialGradient(540, 470, 60, 540, 470, 330);
	archGradient.addColorStop(0, "rgba(239,211,132,.22)");
	archGradient.addColorStop(1, "rgba(239,211,132,0)");
	ctx.fillStyle = archGradient;
	ctx.beginPath();
	ctx.ellipse(540, 485, 330, 205, 0, 0, Math.PI * 2);
	ctx.fill();

	const top = ranked.slice(0, 3);
	const positions = [
		{ x: 540, y: 445, radius: 112, cardX: 395, cardY: 552, cardW: 290, cardH: 230, ring: "#D6B767" },
		{ x: 270, y: 472, radius: 86, cardX: 125, cardY: 590, cardW: 290, cardH: 192, ring: "#99A8B2" },
		{ x: 810, y: 472, radius: 86, cardX: 665, cardY: 590, cardW: 290, cardH: 192, ring: "#B97950" },
	];

	for (let index = 0; index < top.length; index++) {
		const member = top[index];
		const pos = positions[index];
		drawPodiumCard(ctx, member, pos.cardX, pos.cardY, pos.cardW, pos.cardH, index + 1, scoreFormatter);
		if (index === 0) {
			drawLaurel(ctx, pos.x, pos.y, pos.radius, "rgba(199,154,48,.75)");
			drawCrown(ctx, pos.x, pos.y - pos.radius - 46);
		}
		drawAvatar(ctx, member, photos[index], pos.x, pos.y, pos.radius, pos.ring);
		drawRankMedal(ctx, index + 1, pos.x - pos.radius * .80, pos.y - pos.radius * .74, index === 0 ? 36 : 29);
	}

	drawDivider(ctx, 830, ranked.length > 3 ? "TOP 4 - 10" : "TOP 3");

	const remaining = ranked.slice(3);
	if (remaining.length) {
		const colWidth = 420;
		const cardH = 72;
		const gapY = 12;
		remaining.forEach((member, index) => {
			const col = index % 2;
			const row = Math.floor(index / 2);
			const x = 105 + col * 455;
			const y = 866 + row * (cardH + gapY);
			drawListCard(ctx, member, photos[index + 3], index + 4, x, y, colWidth, cardH, scoreFormatter);
		});
	} else {
		fillRound(ctx, 150, 885, 780, 170, 30, "#EAF2EE");
		ctx.fillStyle = "#155B49";
		ctx.font = "900 23px Georgia, serif";
		ctx.textAlign = "center";
		ctx.fillText("Konsistensi kecil yang dijaga setiap hari", 540, 950);
		ctx.fillText("akan menjadi kebiasaan besar.", 540, 985);
		ctx.fillStyle = "#71827B";
		ctx.font = "600 14px Arial, sans-serif";
		ctx.fillText("Leaderboard otomatis terisi hingga Top 10 saat anggota bertambah.", 540, 1023);
	}

	fillRound(ctx, 54, 1240, 972, 72, 30, "rgba(3,46,36,.78)");
	strokeRound(ctx, 54, 1240, 972, 72, 30, "rgba(215,180,95,.30)", 1.5);
	drawDiamond(ctx, 380, 1260, 8, "#D7B45F", .8);
	drawDiamond(ctx, 700, 1260, 8, "#D7B45F", .8);
	ctx.fillStyle = "#FFFFFF";
	ctx.font = "900 18px Georgia, serif";
	ctx.textAlign = "center";
	ctx.textBaseline = "middle";
	ctx.fillText("Ruang Sanad Ops", 540, 1263);
	ctx.fillStyle = "rgba(255,255,255,.72)";
	ctx.font = "500 13px Arial, sans-serif";
	ctx.fillText("Overall = rata-rata Tahajud · Tilawah + Tadabur · Dzikir Pagi · Dzikir Petang", 540, 1292);

	return canvas;
}
