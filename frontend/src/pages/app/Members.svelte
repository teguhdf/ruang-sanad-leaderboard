<script lang="ts">
    import { inertia, useForm } from "@inertiajs/svelte";
    import { fly } from "svelte/transition";
    import AppLayout from "@layouts/AppLayout.svelte";
    import type { Member, User } from "@lib/types";
    import {
        Briefcase,
        CalendarDays,
        Clock,
        ExternalLink,
        Fingerprint,
        Pencil,
        Plus,
        Save,
        Users,
        X,
    } from "lucide-svelte";

    interface Props {
        user?: User;
        members?: Member[];
        success?: string;
        error?: string;
    }

    let { user, members = [], success, error }: Props = $props();
    let editingId = $state<number | null>(null);

    const workDayOptions = [
        { value: "mon", label: "Sen" },
        { value: "tue", label: "Sel" },
        { value: "wed", label: "Rab" },
        { value: "thu", label: "Kam" },
        { value: "fri", label: "Jum" },
        { value: "sat", label: "Sab" },
        { value: "sun", label: "Min" },
    ];

    const memberForm = useForm({
        name: "",
        division: "",
        work_days: [] as string[],
        start_time: "",
        fingerprint_id: "",
        google_sheet_url: "",
    });

    function toggleWorkDay(day: string) {
        if (memberForm.work_days.includes(day)) {
            memberForm.work_days = memberForm.work_days.filter((item) => item !== day);
        } else {
            memberForm.work_days = [...memberForm.work_days, day];
        }
    }

    function startEdit(member: Member) {
        editingId = member.id;
        memberForm.name = member.name;
        memberForm.division = member.division;
        memberForm.work_days = [...member.work_days];
        memberForm.start_time = member.start_time;
        memberForm.fingerprint_id = member.fingerprint_id;
        memberForm.google_sheet_url = member.google_sheet_url;
        window.scrollTo({ top: 0, behavior: "smooth" });
    }

    function resetForm() {
        editingId = null;
        memberForm.reset();
        memberForm.clearErrors();
    }

    function submitMember(event: SubmitEvent) {
        event.preventDefault();

        if (editingId) {
            memberForm.put(`/app/members/${editingId}`, {
                preserveScroll: true,
                onSuccess: resetForm,
            });
            return;
        }

        memberForm.post("/app/members", {
            preserveScroll: true,
            onSuccess: resetForm,
        });
    }

    function dayLabel(day: string) {
        return workDayOptions.find((item) => item.value === day)?.label ?? day;
    }
</script>

<AppLayout {user} group="members">
    <div class="border-b border-neutral-200/80 dark:border-white/[0.04]">
        <div class="max-w-6xl mx-auto px-6 pt-8 pb-8">
            <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
                <div>
                    <div class="flex items-center gap-2 text-sm font-medium text-brand-600 dark:text-brand-400 mb-2">
                        <Users class="w-4 h-4" />
                        Tim Ruang Sanad
                    </div>
                    <h1 class="text-3xl font-bold tracking-tight text-neutral-900 dark:text-white">
                        Anggota Tim
                    </h1>
                    <p class="mt-2 text-neutral-600 dark:text-neutral-400 max-w-2xl">
                        Simpan identitas anggota, jadwal kerja tetap, ID fingerprint, dan workspace Google Sheets.
                    </p>
                </div>

                <div class="rounded-xl border border-neutral-200 dark:border-white/[0.06] bg-white dark:bg-neutral-925 px-4 py-3 min-w-36">
                    <div class="text-xs uppercase tracking-wider text-neutral-500">Total anggota</div>
                    <div class="text-2xl font-bold text-neutral-900 dark:text-white mt-1">{members.length}</div>
                </div>
            </div>
        </div>
    </div>

    <div class="max-w-6xl mx-auto px-6 py-8 space-y-8">
        {#if success}
            <div class="rounded-xl border border-green-500/20 bg-green-500/10 px-4 py-3 text-sm font-medium text-green-700 dark:text-green-400" in:fly={{ y: -8, duration: 220 }}>
                {success}
            </div>
        {/if}

        {#if error}
            <div class="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm font-medium text-red-700 dark:text-red-400" in:fly={{ y: -8, duration: 220 }}>
                {error}
            </div>
        {/if}

        <section class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 overflow-hidden">
            <div class="px-6 py-5 border-b border-neutral-200/80 dark:border-white/[0.05] flex items-center justify-between gap-4">
                <div>
                    <h2 class="text-lg font-semibold text-neutral-900 dark:text-white">
                        {editingId ? "Edit anggota" : "Tambah anggota"}
                    </h2>
                    <p class="text-sm text-neutral-500 mt-1">
                        Durasi kerja tetap 8 jam. Di tahap ini kita hanya menyimpan jam masuk, bukan membuat jam pulang.
                    </p>
                </div>
                {#if editingId}
                    <button type="button" onclick={resetForm} class="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-900 dark:hover:text-white">
                        <X class="w-4 h-4" />
                        Batal edit
                    </button>
                {/if}
            </div>

            <form onsubmit={submitMember} class="p-6 grid gap-5 md:grid-cols-2">
                <label class="space-y-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Nama anggota</span>
                    <input
                        bind:value={memberForm.name}
                        type="text"
                        autocomplete="off"
                        placeholder="Contoh: Fathiyah"
                        class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                    />
                </label>

                <label class="space-y-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Divisi</span>
                    <input
                        bind:value={memberForm.division}
                        type="text"
                        autocomplete="off"
                        placeholder="Contoh: SMM"
                        class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                    />
                </label>

                <div class="space-y-2 md:col-span-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Hari kerja tetap</span>
                    <div class="flex flex-wrap gap-2">
                        {#each workDayOptions as day}
                            <button
                                type="button"
                                onclick={() => toggleWorkDay(day.value)}
                                class="min-w-12 rounded-lg border px-3 py-2 text-sm font-semibold transition-colors {memberForm.work_days.includes(day.value)
                                    ? 'border-brand-400 bg-brand-400/10 text-brand-700 dark:text-brand-300'
                                    : 'border-neutral-300 dark:border-neutral-700 text-neutral-500 hover:border-neutral-400 dark:hover:border-neutral-600'}"
                            >
                                {day.label}
                            </button>
                        {/each}
                    </div>
                </div>

                <label class="space-y-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Jam masuk</span>
                    <input
                        bind:value={memberForm.start_time}
                        type="time"
                        class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                    />
                </label>

                <label class="space-y-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">ID fingerprint</span>
                    <input
                        bind:value={memberForm.fingerprint_id}
                        type="text"
                        autocomplete="off"
                        placeholder="Sesuai ID pada mesin fingerprint"
                        class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                    />
                </label>

                <label class="space-y-2 md:col-span-2">
                    <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Link Google Sheets workspace</span>
                    <input
                        bind:value={memberForm.google_sheet_url}
                        type="url"
                        placeholder="https://docs.google.com/spreadsheets/d/..."
                        class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                    />
                </label>

                <div class="md:col-span-2 flex justify-end">
                    <button
                        type="submit"
                        disabled={memberForm.processing}
                        class="inline-flex items-center gap-2 rounded-xl bg-brand-600 hover:bg-brand-700 dark:bg-brand-500 dark:hover:bg-brand-400 px-5 py-3 font-semibold text-white transition-colors disabled:opacity-50"
                    >
                        {#if editingId}
                            <Save class="w-4 h-4" />
                            {memberForm.processing ? "Menyimpan..." : "Simpan perubahan"}
                        {:else}
                            <Plus class="w-4 h-4" />
                            {memberForm.processing ? "Menyimpan..." : "Tambah anggota"}
                        {/if}
                    </button>
                </div>
            </form>
        </section>

        <section>
            <div class="flex items-center justify-between mb-4">
                <div>
                    <h2 class="text-lg font-semibold text-neutral-900 dark:text-white">Daftar anggota</h2>
                    <p class="text-sm text-neutral-500 mt-1">Belum ada skor performa atau amaliyah pada data asli.</p>
                </div>
            </div>

            {#if members.length === 0}
                <div class="rounded-2xl border border-dashed border-neutral-300 dark:border-neutral-700 py-14 px-6 text-center">
                    <Users class="w-8 h-8 mx-auto text-neutral-400 mb-3" />
                    <h3 class="font-semibold text-neutral-800 dark:text-neutral-200">Belum ada anggota</h3>
                    <p class="text-sm text-neutral-500 mt-1">Tambahkan satu anggota untuk menguji penyimpanan dan refresh.</p>
                </div>
            {:else}
                <div class="grid gap-4 lg:grid-cols-2">
                    {#each members as member (member.id)}
                        <article class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 p-5">
                            <div class="flex items-start justify-between gap-4">
                                <div class="min-w-0">
                                    <h3 class="text-lg font-semibold text-neutral-900 dark:text-white truncate">{member.name}</h3>
                                    <div class="mt-1 inline-flex items-center gap-1.5 text-sm text-neutral-500">
                                        <Briefcase class="w-4 h-4" />
                                        {member.division}
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    onclick={() => startEdit(member)}
                                    class="shrink-0 inline-flex items-center gap-1.5 rounded-lg border border-neutral-200 dark:border-neutral-700 px-3 py-2 text-sm font-medium text-neutral-600 dark:text-neutral-300 hover:border-brand-400/50 hover:text-brand-600 dark:hover:text-brand-400"
                                >
                                    <Pencil class="w-4 h-4" />
                                    Edit
                                </button>
                            </div>

                            <div class="mt-5 grid gap-3 sm:grid-cols-2 text-sm">
                                <div class="flex items-start gap-2 text-neutral-600 dark:text-neutral-400">
                                    <CalendarDays class="w-4 h-4 mt-0.5 shrink-0" />
                                    <span>{member.work_days.map(dayLabel).join(", ")}</span>
                                </div>
                                <div class="flex items-start gap-2 text-neutral-600 dark:text-neutral-400">
                                    <Clock class="w-4 h-4 mt-0.5 shrink-0" />
                                    <span>Masuk {member.start_time}</span>
                                </div>
                                <div class="flex items-start gap-2 text-neutral-600 dark:text-neutral-400">
                                    <Fingerprint class="w-4 h-4 mt-0.5 shrink-0" />
                                    <span class="break-all">{member.fingerprint_id}</span>
                                </div>
                                <a
                                    href={member.google_sheet_url}
                                    target="_blank"
                                    rel="noopener"
                                    class="flex items-start gap-2 text-brand-600 dark:text-brand-400 hover:underline"
                                >
                                    <ExternalLink class="w-4 h-4 mt-0.5 shrink-0" />
                                    Buka workspace
                                </a>
                            </div>
                        </article>
                    {/each}
                </div>
            {/if}
        </section>
    </div>
</AppLayout>
