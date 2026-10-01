<script lang="ts">
	import type { Snippet } from "svelte";
	import { fade, fly } from "svelte/transition";
	import { inertia, router } from "@inertiajs/svelte";
	import {
		CalendarClock,
		ChevronRight,
		ClipboardCheck,
		FileSpreadsheet,
		LogOut,
		Menu,
		Settings,
		Trophy,
		UsersRound,
		X,
	} from "lucide-svelte";
	import DarkModeToggle from "@components/DarkModeToggle.svelte";
	import type { User as UserType } from "@lib/types";

	interface Props {
		user?: UserType;
		group?: string;
		children: Snippet;
	}

	let { user, group = "", children }: Props = $props();
	let mobileOpen = $state(false);

	const managementLinks = [
		{ href: "/app/amaliyah", label: "Leaderboard Amaliyah", group: "amaliyah", icon: Trophy },
		{ href: "/app/members", label: "Anggota Tim", group: "members", icon: UsersRound },
		{ href: "/app/attendance", label: "Attendance", group: "attendance", icon: CalendarClock },
		{ href: "/app/attendance/reviews", label: "Approval", group: "attendance-reviews", icon: ClipboardCheck },
		{ href: "/app/fingerprint", label: "Fingerprint", group: "fingerprint", icon: FileSpreadsheet },
		{ href: "/app/profile", label: "Pengaturan", group: "profile", icon: Settings },
	];

	const employeeLinks = [
		{ href: "/app/my-attendance", label: "Absensi Saya", group: "my-attendance", icon: CalendarClock },
		{ href: "/app/profile", label: "Pengaturan", group: "profile", icon: Settings },
	];

	function menuLinks() {
		return user?.role === "employee" ? employeeLinks : managementLinks;
	}

	function homeHref() {
		return user?.role === "employee" ? "/app/my-attendance" : "/app/amaliyah";
	}

	function pageTitle() {
		return menuLinks().find((item) => item.group === group)?.label ?? "Ruang Sanad Ops";
	}

	function handleLogout() {
		router.post("/logout");
	}

	$effect(() => {
		if (typeof document === "undefined") return;
		document.body.style.overflow = mobileOpen ? "hidden" : "";
		return () => {
			document.body.style.overflow = "";
		};
	});
</script>

<div class="min-h-screen bg-neutral-50 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-50">
	<aside class="fixed inset-y-0 left-0 z-40 hidden w-64 flex-col border-r border-neutral-200/80 bg-white/95 px-3 py-4 backdrop-blur-xl dark:border-white/[0.06] dark:bg-neutral-925/95 lg:flex">
		<a href={homeHref()} use:inertia class="mb-5 flex items-center gap-3 rounded-2xl px-3 py-3 no-underline transition-colors hover:bg-neutral-100 dark:hover:bg-white/[0.05]">
			<div class="grid h-11 w-11 place-items-center rounded-2xl bg-brand-600 text-sm font-black tracking-tight text-white shadow-sm">RS</div>
			<div class="min-w-0">
				<div class="truncate text-[15px] font-bold tracking-tight text-neutral-900 dark:text-white">Ruang Sanad Ops</div>
				<div class="mt-0.5 text-[11px] font-medium text-neutral-500 dark:text-neutral-400">{user?.role === "employee" ? "Employee Workspace" : "Team Operations"}</div>
			</div>
		</a>

		<div class="px-3 pb-2 text-[10px] font-bold uppercase tracking-[0.14em] text-neutral-400">Workspace</div>
		<nav class="grid gap-1.5">
			{#each menuLinks() as item}
				{@const Icon = item.icon}
				<a href={item.href} use:inertia class="group flex min-h-12 items-center gap-3 rounded-2xl px-3.5 text-sm font-semibold transition-all {item.group === group ? 'bg-brand-100 text-brand-800 dark:bg-brand-900/55 dark:text-brand-100' : 'text-neutral-600 hover:bg-neutral-100 hover:text-neutral-900 dark:text-neutral-400 dark:hover:bg-white/[0.05] dark:hover:text-white'}">
					<div class="grid h-8 w-8 place-items-center rounded-xl {item.group === group ? 'bg-white/65 dark:bg-white/[0.08]' : 'bg-transparent'}"><Icon size="19" stroke-width="2" /></div>
					<span>{item.label}</span>
					{#if item.group === group}<ChevronRight size="16" class="ml-auto opacity-70" />{/if}
				</a>
			{/each}
		</nav>

		<div class="mt-auto rounded-2xl bg-neutral-100/80 p-3 dark:bg-white/[0.04]">
			<div class="flex items-center gap-3">
				{#if user?.avatar}
					<img src={user.avatar} alt={user.name} class="h-10 w-10 rounded-2xl object-cover" />
				{:else}
					<div class="grid h-10 w-10 place-items-center rounded-2xl bg-brand-600 text-sm font-extrabold text-white">{user?.name?.charAt(0).toUpperCase() ?? "U"}</div>
				{/if}
				<div class="min-w-0 flex-1">
					<div class="truncate text-xs font-bold text-neutral-900 dark:text-white">{user?.name ?? "User"}</div>
					<div class="truncate text-[10px] text-neutral-500 dark:text-neutral-400">{user?.role ?? ""} · {user?.email ?? ""}</div>
				</div>
			</div>
			<div class="mt-3 flex items-center gap-2">
				<div class="flex-1"><DarkModeToggle /></div>
				<button type="button" onclick={handleLogout} class="grid h-9 w-9 place-items-center rounded-xl text-error transition-colors hover:bg-error/10" aria-label="Logout"><LogOut size="17" /></button>
			</div>
		</div>
	</aside>

	<header class="sticky top-0 z-30 flex h-16 items-center gap-3 border-b border-neutral-200/80 bg-neutral-50/92 px-4 backdrop-blur-xl dark:border-white/[0.06] dark:bg-neutral-950/92 lg:hidden">
		<button type="button" onclick={() => (mobileOpen = true)} class="grid h-10 w-10 place-items-center rounded-full hover:bg-neutral-200/70 dark:hover:bg-white/[0.06]" aria-label="Buka menu"><Menu size="21" /></button>
		<div class="grid h-9 w-9 place-items-center rounded-xl bg-brand-600 text-xs font-black text-white">RS</div>
		<div class="min-w-0 flex-1"><div class="truncate text-sm font-bold">{pageTitle()}</div><div class="text-[10px] text-neutral-500">Ruang Sanad Ops</div></div>
		<DarkModeToggle />
	</header>

	{#if mobileOpen}
		<div class="fixed inset-0 z-50 lg:hidden">
			<button type="button" class="absolute inset-0 bg-black/35 backdrop-blur-[2px]" onclick={() => (mobileOpen = false)} aria-label="Tutup menu" transition:fade={{ duration: 140 }}></button>
			<aside class="absolute inset-y-0 left-0 flex w-[min(86vw,320px)] flex-col bg-white p-4 shadow-2xl dark:bg-neutral-925" transition:fly={{ x: -24, duration: 190 }}>
				<div class="mb-5 flex items-center gap-3 px-2 py-2"><div class="grid h-10 w-10 place-items-center rounded-2xl bg-brand-600 text-xs font-black text-white">RS</div><div class="flex-1"><div class="text-sm font-bold">Ruang Sanad Ops</div><div class="text-[10px] text-neutral-500">{user?.role === "employee" ? "Employee Workspace" : "Team Operations"}</div></div><button type="button" onclick={() => (mobileOpen = false)} class="grid h-9 w-9 place-items-center rounded-full hover:bg-neutral-100 dark:hover:bg-white/[0.05]" aria-label="Tutup menu"><X size="20" /></button></div>
				<nav class="grid gap-1.5">
					{#each menuLinks() as item}
						{@const Icon = item.icon}
						<a href={item.href} use:inertia onclick={() => (mobileOpen = false)} class="flex min-h-12 items-center gap-3 rounded-2xl px-3.5 text-sm font-semibold {item.group === group ? 'bg-brand-100 text-brand-800 dark:bg-brand-900/55 dark:text-brand-100' : 'text-neutral-600 hover:bg-neutral-100 dark:text-neutral-300 dark:hover:bg-white/[0.05]'}"><Icon size="19" />{item.label}</a>
					{/each}
				</nav>
				<div class="mt-auto rounded-2xl bg-neutral-100 p-3 dark:bg-white/[0.04]"><div class="text-xs font-bold">{user?.name ?? "User"}</div><div class="mt-0.5 truncate text-[10px] text-neutral-500">{user?.email ?? ""}</div><button type="button" onclick={handleLogout} class="mt-3 flex min-h-10 w-full items-center justify-center gap-2 rounded-full bg-error/10 px-4 text-xs font-bold text-error"><LogOut size="16" /> Logout</button></div>
			</aside>
		</div>
	{/if}

	<div class="lg:pl-64">
		<header class="sticky top-0 z-20 hidden h-16 items-center border-b border-neutral-200/80 bg-neutral-50/90 px-7 backdrop-blur-xl dark:border-white/[0.06] dark:bg-neutral-950/90 lg:flex">
			<div><div class="text-sm font-bold tracking-tight text-neutral-900 dark:text-white">{pageTitle()}</div><div class="mt-0.5 text-[10px] font-medium text-neutral-500">Ruang Sanad · {user?.role === "employee" ? "Employee" : "Team Operations"}</div></div>
			<div class="ml-auto flex items-center gap-2"><DarkModeToggle /><div class="ml-1 h-8 w-px bg-neutral-200 dark:bg-white/[0.06]"></div><div class="flex items-center gap-2 rounded-full bg-white px-2.5 py-1.5 shadow-sm ring-1 ring-neutral-200 dark:bg-neutral-925 dark:ring-white/[0.06]"><div class="grid h-7 w-7 place-items-center rounded-full bg-brand-600 text-[10px] font-extrabold text-white">{user?.name?.charAt(0).toUpperCase() ?? "U"}</div><span class="max-w-36 truncate pr-1 text-xs font-bold">{user?.name ?? "User"}</span></div></div>
		</header>
		<main class="min-h-[calc(100vh-4rem)]">{@render children()}</main>
	</div>
</div>
