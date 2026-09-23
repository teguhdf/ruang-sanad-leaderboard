<script lang="ts">
    import AppLayout from "@layouts/AppLayout.svelte";
    import { getCSRFToken } from "@lib/utils/csrf";
    import type { FingerprintPreview, User } from "@lib/types";
    import {
        CalendarDays,
        CheckCircle2,
        FileSpreadsheet,
        Fingerprint,
        Upload,
        UserCheck,
        UserX,
    } from "lucide-svelte";

    interface Props {
        user?: User;
    }

    let { user }: Props = $props();

    let selectedFile = $state<File | null>(null);
    let preview = $state<FingerprintPreview | null>(null);
    let loading = $state(false);
    let errorMessage = $state("");

    function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;
        selectedFile = input.files?.[0] ?? null;
        preview = null;
        errorMessage = "";
    }

    async function createPreview() {
        if (!selectedFile || loading) return;

        loading = true;
        preview = null;
        errorMessage = "";

        const formData = new FormData();
        formData.append("file", selectedFile);

        try {
            const response = await fetch("/app/fingerprint/preview", {
                method: "POST",
                headers: {
                    "X-XSRF-TOKEN": getCSRFToken(),
                },
                body: formData,
            });

            const payload = await response.json();
            if (!response.ok) {
                errorMessage = payload?.error || "Preview fingerprint gagal dibuat.";
                return;
            }

            preview = payload as FingerprintPreview;
        } catch {
            errorMessage = "Tidak dapat terhubung ke server.";
        } finally {
            loading = false;
        }
    }

    function stateLabel(state: string) {
        if (state === "two_times") return "2 waktu";
        if (state === "one_time") return "1 waktu";
        if (state === "multiple_times") return ">2 waktu";
        return "Kosong";
    }
</script>

<AppLayout {user} group="fingerprint">
    <div class="border-b border-neutral-200/80 dark:border-white/[0.04]">
        <div class="max-w-6xl mx-auto px-4 sm:px-6 pt-8 pb-8">
            <div class="flex items-center gap-2 text-sm font-medium text-brand-600 dark:text-brand-400 mb-2">
                <Fingerprint class="w-4 h-4" />
                Data Fingerprint
            </div>
            <h1 class="text-3xl font-bold tracking-tight text-neutral-900 dark:text-white">
                Import Fingerprint
            </h1>
            <p class="mt-2 text-neutral-600 dark:text-neutral-400 max-w-3xl">
                Baca file .xls, cocokkan ID fingerprint dengan anggota tim, lalu periksa datanya sebelum ada penyimpanan.
            </p>
        </div>
    </div>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 py-8 space-y-6">
        <section class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 p-5 sm:p-6">
            <div class="flex items-start gap-3 mb-5">
                <div class="w-10 h-10 rounded-xl bg-brand-400/10 flex items-center justify-center shrink-0">
                    <FileSpreadsheet class="w-5 h-5 text-brand-600 dark:text-brand-400" />
                </div>
                <div>
                    <h2 class="font-semibold text-neutral-900 dark:text-white">Pilih laporan fingerprint</h2>
                    <p class="text-sm text-neutral-500 mt-1">Hanya .xls. Maksimal 2 MB. File tidak disimpan pada tahap preview.</p>
                </div>
            </div>

            <div class="flex flex-col sm:flex-row gap-3">
                <label class="flex-1">
                    <input
                        type="file"
                        accept=".xls,application/vnd.ms-excel"
                        onchange={handleFileChange}
                        class="block w-full text-sm text-neutral-600 dark:text-neutral-300 file:mr-4 file:rounded-lg file:border-0 file:bg-neutral-100 dark:file:bg-neutral-800 file:px-4 file:py-2.5 file:font-semibold file:text-neutral-700 dark:file:text-neutral-200 hover:file:bg-neutral-200 dark:hover:file:bg-neutral-700"
                    />
                </label>

                <button
                    type="button"
                    onclick={createPreview}
                    disabled={!selectedFile || loading}
                    class="inline-flex items-center justify-center gap-2 rounded-xl bg-brand-600 hover:bg-brand-700 dark:bg-brand-500 dark:hover:bg-brand-400 px-5 py-3 font-semibold text-white disabled:opacity-50 disabled:cursor-not-allowed"
                >
                    <Upload class="w-4 h-4" />
                    {loading ? "Membaca file..." : "Buat preview"}
                </button>
            </div>

            {#if selectedFile}
                <p class="mt-3 text-xs text-neutral-500 break-all">{selectedFile.name}</p>
            {/if}

            {#if errorMessage}
                <div class="mt-4 rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm font-medium text-red-700 dark:text-red-400">
                    {errorMessage}
                </div>
            {/if}
        </section>

        {#if preview}
            <section class="rounded-2xl border border-amber-400/25 bg-amber-50/70 dark:bg-amber-500/5 px-5 py-4">
                <p class="text-sm font-semibold text-amber-900 dark:text-amber-300">Preview saja, belum disimpan</p>
                <p class="text-sm text-amber-800/80 dark:text-amber-300/70 mt-1">
                    Sel kosong tidak dianggap absen dan satu waktu tidak dibuatkan jam pulang.
                </p>
            </section>

            <section class="grid grid-cols-2 lg:grid-cols-4 gap-3">
                <div class="rounded-2xl border border-neutral-200 dark:border-white/[0.06] bg-white dark:bg-neutral-925 p-4">
                    <div class="text-xs uppercase tracking-wider text-neutral-500">Orang di file</div>
                    <div class="text-2xl font-bold mt-1 text-neutral-900 dark:text-white">{preview.summary.file_people}</div>
                </div>
                <div class="rounded-2xl border border-green-500/20 bg-green-500/5 p-4">
                    <div class="text-xs uppercase tracking-wider text-green-700 dark:text-green-400">Matched</div>
                    <div class="text-2xl font-bold mt-1 text-green-800 dark:text-green-300">{preview.summary.matched}</div>
                </div>
                <div class="rounded-2xl border border-red-500/20 bg-red-500/5 p-4">
                    <div class="text-xs uppercase tracking-wider text-red-700 dark:text-red-400">Unmatched</div>
                    <div class="text-2xl font-bold mt-1 text-red-800 dark:text-red-300">{preview.summary.unmatched}</div>
                </div>
                <div class="rounded-2xl border border-neutral-200 dark:border-white/[0.06] bg-white dark:bg-neutral-925 p-4">
                    <div class="text-xs uppercase tracking-wider text-neutral-500">Slot tanggal</div>
                    <div class="text-2xl font-bold mt-1 text-neutral-900 dark:text-white">{preview.summary.total_slots}</div>
                </div>
            </section>

            <section class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 overflow-hidden">
                <div class="p-5 sm:p-6 border-b border-neutral-200/80 dark:border-white/[0.05] flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                    <div>
                        <h2 class="text-lg font-semibold text-neutral-900 dark:text-white">Hasil pembacaan</h2>
                        <div class="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm text-neutral-500">
                            <span>{preview.period_start} sampai {preview.period_end}</span>
                            <span>2 waktu: {preview.summary.two_times}</span>
                            <span>1 waktu: {preview.summary.one_time}</span>
                            <span>Kosong: {preview.summary.empty}</span>
                        </div>
                    </div>
                    <div class="text-xs text-neutral-500 break-all">{preview.file_name}</div>
                </div>

                <div class="divide-y divide-neutral-200/80 dark:divide-white/[0.05]">
                    {#each preview.people as person}
                        <details class="group">
                            <summary class="list-none cursor-pointer p-5 sm:p-6 hover:bg-neutral-50 dark:hover:bg-neutral-900/40">
                                <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                                    <div class="flex items-start gap-3 min-w-0">
                                        <div class="w-10 h-10 rounded-full bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center shrink-0">
                                            <Fingerprint class="w-5 h-5 text-neutral-500" />
                                        </div>
                                        <div class="min-w-0">
                                            <div class="font-semibold text-neutral-900 dark:text-white truncate">
                                                ID {person.fingerprint_id} · {person.file_name}
                                            </div>
                                            <div class="text-sm text-neutral-500 mt-0.5">{person.department || "Tanpa departemen"}</div>
                                        </div>
                                    </div>

                                    <div class="flex flex-wrap items-center gap-2">
                                        {#if person.matched}
                                            <span class="inline-flex items-center gap-1.5 rounded-full bg-green-500/10 px-3 py-1.5 text-xs font-semibold text-green-700 dark:text-green-400">
                                                <UserCheck class="w-3.5 h-3.5" />
                                                Matched: {person.member_name}
                                            </span>
                                        {:else}
                                            <span class="inline-flex items-center gap-1.5 rounded-full bg-red-500/10 px-3 py-1.5 text-xs font-semibold text-red-700 dark:text-red-400">
                                                <UserX class="w-3.5 h-3.5" />
                                                Unmatched
                                            </span>
                                        {/if}
                                        <span class="text-xs text-neutral-500">
                                            2 waktu {person.summary.two_times} · 1 waktu {person.summary.one_time} · kosong {person.summary.empty}
                                        </span>
                                    </div>
                                </div>
                            </summary>

                            <div class="px-5 sm:px-6 pb-6">
                                <div class="overflow-x-auto rounded-xl border border-neutral-200 dark:border-white/[0.06]">
                                    <table class="min-w-full text-sm">
                                        <thead class="bg-neutral-50 dark:bg-neutral-900/70 text-left text-neutral-500">
                                            <tr>
                                                <th class="px-4 py-3 font-medium">Tanggal</th>
                                                <th class="px-4 py-3 font-medium">Data waktu asli</th>
                                                <th class="px-4 py-3 font-medium">Pembacaan</th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-neutral-200 dark:divide-white/[0.05]">
                                            {#each person.days as day}
                                                <tr>
                                                    <td class="px-4 py-3 whitespace-nowrap text-neutral-700 dark:text-neutral-300">
                                                        <span class="inline-flex items-center gap-1.5">
                                                            <CalendarDays class="w-3.5 h-3.5 text-neutral-400" />
                                                            {day.date || `Hari ${day.day}`}
                                                        </span>
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-pre-line font-mono text-neutral-700 dark:text-neutral-300">
                                                        {day.times.length ? day.times.join(" / ") : "Tidak ada data"}
                                                    </td>
                                                    <td class="px-4 py-3 whitespace-nowrap">
                                                        {#if day.read_state === "two_times"}
                                                            <span class="inline-flex items-center gap-1 text-green-700 dark:text-green-400">
                                                                <CheckCircle2 class="w-3.5 h-3.5" />
                                                                {stateLabel(day.read_state)}
                                                            </span>
                                                        {:else}
                                                            <span class="text-neutral-500">{stateLabel(day.read_state)}</span>
                                                        {/if}
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </details>
                    {/each}
                </div>
            </section>
        {/if}
    </div>
</AppLayout>
