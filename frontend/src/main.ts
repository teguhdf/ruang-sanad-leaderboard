import { createInertiaApp } from "@inertiajs/svelte";

createInertiaApp();

const flashSelector = ".ring-emerald-200.bg-emerald-50, .ring-red-200.bg-red-50";
const dismissTimers = new WeakMap<HTMLElement, number>();

function dismissFlash(element: HTMLElement) {
	const timer = dismissTimers.get(element);
	if (timer) window.clearTimeout(timer);

	element.style.transition = "opacity 180ms ease, transform 180ms ease";
	element.style.opacity = "0";
	element.style.transform = "translateY(-4px)";
	window.setTimeout(() => {
		element.style.display = "none";
	}, 180);
}

function enhanceFlash(element: HTMLElement) {
	if (element.dataset.autoDismissFlash === "1") return;
	element.dataset.autoDismissFlash = "1";

	if (!element.style.position) element.style.position = "relative";
	element.style.paddingRight = "3rem";

	const close = document.createElement("button");
	close.type = "button";
	close.setAttribute("aria-label", "Tutup notifikasi");
	close.textContent = "×";
	close.style.position = "absolute";
	close.style.top = "50%";
	close.style.right = "0.9rem";
	close.style.transform = "translateY(-50%)";
	close.style.width = "1.75rem";
	close.style.height = "1.75rem";
	close.style.border = "0";
	close.style.borderRadius = "9999px";
	close.style.background = "rgba(255,255,255,0.55)";
	close.style.color = "inherit";
	close.style.font = "700 20px/1 Arial, sans-serif";
	close.style.cursor = "pointer";
	close.style.display = "grid";
	close.style.placeItems = "center";
	close.onclick = () => dismissFlash(element);
	element.appendChild(close);

	const timer = window.setTimeout(() => dismissFlash(element), 8000);
	dismissTimers.set(element, timer);
}

function scanFlash(root: ParentNode = document) {
	root.querySelectorAll<HTMLElement>(flashSelector).forEach(enhanceFlash);
}

function startFlashEnhancer() {
	scanFlash();

	const existing = (window as Window & { __ruangSanadFlashObserver?: MutationObserver }).__ruangSanadFlashObserver;
	existing?.disconnect();

	const observer = new MutationObserver(() => scanFlash());
	observer.observe(document.body, { childList: true, subtree: true });
	(window as Window & { __ruangSanadFlashObserver?: MutationObserver }).__ruangSanadFlashObserver = observer;
}

if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", startFlashEnhancer, { once: true });
} else {
	startFlashEnhancer();
}
