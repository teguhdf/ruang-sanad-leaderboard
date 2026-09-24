<script lang="ts">
    import AppLayout from "@layouts/AppLayout.svelte";
    import { getCSRFToken } from "@lib/utils/csrf";
    import type {
        FingerprintDayPreview,
        FingerprintImportHistory,
        FingerprintPersonPreview,
        FingerprintPreview,
        User,
    } from "@lib/types";
    import {
        CalendarDays,
        CheckCircle2,
        FileSpreadsheet,
        Fingerprint,
        Pencil,
        RotateCcw,
        Save,
        Upload,
        UserCheck,
        UserX,
        X,
    } from "lucide-svelte";

    interface Props {
        user?: User;
        imports?: FingerprintImportHistory[];
    }

    interface CorrectionDraft {
        check_in: string;
        check_out: string;
        reason: string;
        editing: boolean;
        applied: boolean;
        error: string;
    }

    let { user, imports: initialImports = [] }: Props = $props();

    let selectedFile = $state<File | null>(null);
    let preview = $state<FingerprintPreview | null>(null);
    let loading = $state(false);
    let saving = $state(false);
    let errorMessage = $state("");
    let saveMessage = $state("");
    let saved = $state(false);
    let corrections = $state<Record<string, CorrectionDraft>>({});
    let importHistory = $state<FingerprintImportHistory[]>(initialImports);

    function correctionKey(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        return `${person.fingerprint_id}:${day.date || day.day}`;
    }

    function rawCheckIn(day: FingerprintDayPreview) {
        return day.times[0] ?? "";
    }

    function rawCheckOut(day: FingerprintDayPreview) {
        return day.times[1] ?? "";
    }

    function makeCorrection(day: FingerprintDayPreview): CorrectionDraft {
        return {
            check_in: rawCheckIn(day),
            check_out: rawCheckOut(day),
            reason: "",
            editing: false,
            applied: false,
            error: "",
        };
    }

    function initializeCorrections(data: FingerprintPreview) {
        const next: Record<string, CorrectionDraft> = {};

        for (const person of data.people) {
            for (const day of person.days) {
                next[correctionKey(person, day)] = makeCorrection(day);
            }
        }

        corrections = next;
    }

    function getCorrection(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        return corrections[correctionKey(person, day)];
    }

    function isChanged(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        const draft = getCorrection(person, day);
        return draft.check_in !== rawCheckIn(day) || draft.check_out !== rawCheckOut(day);
    }

    function startCorrection(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        const draft = getCorrection(person, day);
        draft.editing = true;
        draft.error = "";
    }

    function cancelCorrection(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        const draft = getCorrection(person, day);
        draft.check_in = rawCheckIn(day);
        draft.check_out = rawCheckOut(day);
        draft.reason = "";
        draft.editing = false;
        draft.applied = false;
        draft.error = "";
    }

    function applyCorrection(person: FingerprintPersonPreview, day: FingerprintDayPreview) {
        const draft = getCorrection(person, day);

        if (!isChanged(person, day)) {
            draft.error = "Belum ada perubahan waktu.";
            return;
        }

        if (!draft.reason.trim()) {
            draft.error = "Alasan koreksi wajib diisi.";
            return;
        }

        draft.reason = draft.reason.trim();
        draft.applied = true;
        draft.editing = false;
        draft.error = "";
    }

    function correctedCount() {
        return Object.values(corrections).filter((item) => item.applied).length;
    }

    function handleFileChange(event: Event) {
        const input = event.target as HTMLInputElement;
        selectedFile = input.files?.[0] ?? null;
        preview = null;
        errorMessage = "";
        saveMessage = "";
        saved = false;
        corrections = {};
    }

    async function createPreview() {
        if (!selectedFile || loading) return;

        loading = true;
        preview = null;
        errorMessage = "";
        saveMessage = "";
        saved = false;
        corrections = {};

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

            const parsedPreview = payload as FingerprintPreview;
            initializeCorrections(parsedPreview);
            preview = parsedPreview;
        } catch {
            errorMessage = "Tidak dapat terhubung ke server.";
        } finally {
            loading = false;
        }
    }


    async function saveImport() {
        if (!selectedFile || !preview || saving || saved) return;

        if (Object.values(corrections).some((item) => item.editing)) {
            errorMessage = "Selesaikan atau batalkan koreksi yang sedang diedit sebelum menyimpan.";
            return;
        }

        const appliedCorrections: Array<{
            fingerprint_id: string;
            date: string;
            check_in: string;
            check_out: string;
            reason: string;
        }> = [];

        for (const person of preview.people) {
            for (const day of person.days) {
                const draft = getCorrection(person, day);
                if (!draft?.applied) continue;

                appliedCorrections.push({
                    fingerprint_id: person.fingerprint_id,
                    date: day.date,
                    check_in: draft.check_in,
                    check_out: draft.check_out,
                    reason: draft.reason,
                });
            }
        }

        saving = true;
        errorMessage = "";
        saveMessage = "";

        const formData = new FormData();
        formData.append("file", selectedFile);
        formData.append("corrections", JSON.stringify(appliedCorrections));

        try {
            const response = await fetch("/app/fingerprint/import", {
                method: "POST",
                headers: {
                    "X-XSRF-TOKEN": getCSRFToken(),
                },
                body: formData,
            });

            const payload = await response.json();
            if (!response.ok) {
                errorMessage = payload?.error || "Import fingerprint gagal disimpan.";
                return;
            }

            saveMessage = payload?.message || "Import fingerprint berhasil disimpan.";
            saved = true;

            const savedImport = payload?.import as FingerprintImportHistory | undefined;
            if (savedImport && !importHistory.some((item) => item.id === savedImport.id)) {
                importHistory = [savedImport, ...importHistory].slice(0, 10);
            }
        } catch {
            errorMessage = "Tidak dapat terhubung ke server saat menyimpan import.";
        } finally {
            saving = false;
        }
    }

    function formatImportedAt(value: string) {
        if (!value) return "";
        const parsed = new Date(value);
        if (Number.isNaN(parsed.getTime())) return value;
        return parsed.toLocaleString("id-ID");
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
                    Sel kosong tidak dianggap absen. Waktu pertama dibaca sebagai jam masuk, waktu kedua sebagai jam pulang. Koreksi di halaman ini masih berupa draft dan belum masuk database.
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
                            <span>Koreksi manual: {correctedCount()}</span>
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
                                <div class="space-y-3">
                                    {#each person.days as day}
                                        {@const draft = getCorrection(person, day)}
                                        <article class="rounded-xl border border-neutral-200 dark:border-white/[0.06] bg-neutral-50/60 dark:bg-neutral-900/30 p-4">
                                            <div class="flex flex-col lg:flex-row lg:items-start lg:justify-between gap-4">
                                                <div class="min-w-36">
                                                    <div class="inline-flex items-center gap-1.5 text-sm font-semibold text-neutral-800 dark:text-neutral-200">
                                                        <CalendarDays class="w-3.5 h-3.5 text-neutral-400" />
                                                        {day.date || `Hari ${day.day}`}
                                                    </div>
                                                    <div class="mt-1 text-xs text-neutral-500">
                                                        {stateLabel(day.read_state)}
                                                    </div>
                                                </div>

                                                <div class="flex-1 grid gap-3 md:grid-cols-2">
                                                    <div class="rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-925 p-3">
                                                        <div class="text-[11px] uppercase tracking-wider text-neutral-500 mb-2">Data asli fingerprint</div>
                                                        <div class="grid grid-cols-2 gap-3">
                                                            <div>
                                                                <div class="text-xs text-neutral-500 mb-1">Jam masuk</div>
                                                                <div class="font-mono font-semibold text-neutral-900 dark:text-white">
                                                                    {rawCheckIn(day) || "Tidak ada"}
                                                                </div>
                                                            </div>
                                                            <div>
                                                                <div class="text-xs text-neutral-500 mb-1">Jam pulang</div>
                                                                <div class="font-mono font-semibold text-neutral-900 dark:text-white">
                                                                    {rawCheckOut(day) || "Tidak ada"}
                                                                </div>
                                                            </div>
                                                        </div>
                                                        {#if day.times.length > 2}
                                                            <div class="mt-2 text-xs text-amber-700 dark:text-amber-400">
                                                                File memiliki waktu tambahan: {day.times.slice(2).join(", ")}
                                                            </div>
                                                        {/if}
                                                    </div>

                                                    <div class="rounded-lg border {draft.applied ? 'border-brand-400/40 bg-brand-400/5' : 'border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-925'} p-3">
                                                        <div class="flex items-center justify-between gap-2 mb-2">
                                                            <div class="text-[11px] uppercase tracking-wider text-neutral-500">Hasil setelah koreksi</div>
                                                            {#if draft.applied}
                                                                <span class="inline-flex items-center gap-1 text-[11px] font-semibold text-brand-700 dark:text-brand-300">
                                                                    <CheckCircle2 class="w-3.5 h-3.5" />
                                                                    Dikoreksi manual
                                                                </span>
                                                            {/if}
                                                        </div>
                                                        <div class="grid grid-cols-2 gap-3">
                                                            <div>
                                                                <div class="text-xs text-neutral-500 mb-1">Jam masuk</div>
                                                                <div class="font-mono font-semibold text-neutral-900 dark:text-white">
                                                                    {draft.applied ? (draft.check_in || "Tidak ada") : (rawCheckIn(day) || "Tidak ada")}
                                                                </div>
                                                            </div>
                                                            <div>
                                                                <div class="text-xs text-neutral-500 mb-1">Jam pulang</div>
                                                                <div class="font-mono font-semibold text-neutral-900 dark:text-white">
                                                                    {draft.applied ? (draft.check_out || "Tidak ada") : (rawCheckOut(day) || "Tidak ada")}
                                                                </div>
                                                            </div>
                                                        </div>
                                                        {#if draft.applied}
                                                            <div class="mt-2 text-xs text-neutral-500">
                                                                Alasan: <span class="text-neutral-700 dark:text-neutral-300">{draft.reason}</span>
                                                            </div>
                                                        {/if}
                                                    </div>
                                                </div>

                                                <div class="lg:w-36 lg:text-right">
                                                    {#if !draft.editing}
                                                        <button
                                                            type="button"
                                                            onclick={() => startCorrection(person, day)}
                                                            class="inline-flex items-center justify-center gap-1.5 rounded-lg border border-neutral-300 dark:border-neutral-700 px-3 py-2 text-sm font-medium text-neutral-600 dark:text-neutral-300 hover:border-brand-400/60 hover:text-brand-600 dark:hover:text-brand-400"
                                                        >
                                                            <Pencil class="w-4 h-4" />
                                                            {draft.applied ? "Edit koreksi" : "Koreksi"}
                                                        </button>
                                                    {/if}
                                                </div>
                                            </div>

                                            {#if draft.editing}
                                                <div class="mt-4 pt-4 border-t border-neutral-200 dark:border-white/[0.05]">
                                                    <div class="grid gap-4 md:grid-cols-2">
                                                        <label class="space-y-1.5">
                                                            <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Jam masuk</span>
                                                            <input
                                                                bind:value={draft.check_in}
                                                                type="time"
                                                                class="w-full rounded-lg border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-950 px-3 py-2.5 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                                                            />
                                                        </label>
                                                        <label class="space-y-1.5">
                                                            <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Jam pulang</span>
                                                            <input
                                                                bind:value={draft.check_out}
                                                                type="time"
                                                                class="w-full rounded-lg border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-950 px-3 py-2.5 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15"
                                                            />
                                                            <span class="block text-xs text-neutral-500">Jam kedua pada fingerprint diperlakukan sebagai jam pulang.</span>
                                                        </label>
                                                    </div>

                                                    <label class="block mt-4 space-y-1.5">
                                                        <span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">
                                                            Alasan koreksi <span class="text-red-500">*</span>
                                                        </span>
                                                        <textarea
                                                            bind:value={draft.reason}
                                                            rows="2"
                                                            placeholder="Contoh: lupa fingerprint pulang, sudah dikonfirmasi admin."
                                                            class="w-full rounded-lg border border-neutral-300 dark:border-neutral-700 bg-white dark:bg-neutral-950 px-3 py-2.5 text-neutral-900 dark:text-white outline-none focus:border-brand-400 focus:ring-2 focus:ring-brand-400/15 resize-y"
                                                        ></textarea>
                                                    </label>

                                                    {#if draft.error}
                                                        <div class="mt-2 text-sm font-medium text-red-600 dark:text-red-400">{draft.error}</div>
                                                    {/if}

                                                    <div class="mt-4 flex flex-wrap justify-end gap-2">
                                                        <button
                                                            type="button"
                                                            onclick={() => cancelCorrection(person, day)}
                                                            class="inline-flex items-center gap-1.5 rounded-lg border border-neutral-300 dark:border-neutral-700 px-3 py-2 text-sm font-medium text-neutral-600 dark:text-neutral-300"
                                                        >
                                                            {#if draft.applied}
                                                                <RotateCcw class="w-4 h-4" />
                                                                Kembalikan ke data asli
                                                            {:else}
                                                                <X class="w-4 h-4" />
                                                                Batal
                                                            {/if}
                                                        </button>
                                                        <button
                                                            type="button"
                                                            onclick={() => applyCorrection(person, day)}
                                                            class="inline-flex items-center gap-1.5 rounded-lg bg-brand-600 hover:bg-brand-700 dark:bg-brand-500 dark:hover:bg-brand-400 px-4 py-2 text-sm font-semibold text-white"
                                                        >
                                                            <Save class="w-4 h-4" />
                                                            Terapkan koreksi
                                                        </button>
                                                    </div>
                                                </div>
                                            {/if}
                                        </article>
                                    {/each}
                                </div>
                            </div>
                        </details>
                    {/each}
                </div>
            </section>

            <section class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 p-5 sm:p-6">
                <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
                    <div>
                        <h2 class="text-lg font-semibold text-neutral-900 dark:text-white">Simpan hasil import</h2>
                        <p class="mt-1 text-sm text-neutral-500 max-w-2xl">
                            Data asli tetap disimpan sebagai evidence. Koreksi manual disimpan terpisah beserta alasannya. File .xls yang sama tidak akan membuat import duplikat.
                        </p>
                        {#if preview.summary.unmatched > 0}
                            <p class="mt-2 text-sm text-amber-700 dark:text-amber-400">
                                Ada {preview.summary.unmatched} orang yang belum matched. Record-nya tetap tersimpan sebagai evidence, tetapi belum terhubung ke anggota tim.
                            </p>
                        {/if}
                    </div>

                    <button
                        type="button"
                        onclick={saveImport}
                        disabled={saving || saved || !selectedFile}
                        class="inline-flex items-center justify-center gap-2 rounded-xl bg-brand-600 hover:bg-brand-700 dark:bg-brand-500 dark:hover:bg-brand-400 px-5 py-3 font-semibold text-white disabled:opacity-50 disabled:cursor-not-allowed shrink-0"
                    >
                        <Save class="w-4 h-4" />
                        {saving ? "Menyimpan..." : saved ? "Sudah disimpan" : "Simpan import"}
                    </button>
                </div>

                {#if saveMessage}
                    <div class="mt-4 rounded-xl border border-green-500/20 bg-green-500/10 px-4 py-3 text-sm font-medium text-green-700 dark:text-green-400">
                        {saveMessage}
                    </div>
                {/if}
            </section>
        {/if}

        <section class="rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 overflow-hidden">
            <div class="p-5 sm:p-6 border-b border-neutral-200/80 dark:border-white/[0.05]">
                <h2 class="text-lg font-semibold text-neutral-900 dark:text-white">Riwayat import</h2>
                <p class="mt-1 text-sm text-neutral-500">10 import fingerprint terbaru untuk akun ini.</p>
            </div>

            {#if importHistory.length === 0}
                <div class="p-6 text-sm text-neutral-500">Belum ada import yang disimpan.</div>
            {:else}
                <div class="divide-y divide-neutral-200/80 dark:divide-white/[0.05]">
                    {#each importHistory as item}
                        <div class="p-5 sm:px-6 flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3">
                            <div class="min-w-0">
                                <div class="font-semibold text-neutral-900 dark:text-white truncate">{item.file_name}</div>
                                <div class="mt-1 text-sm text-neutral-500">
                                    {item.period_start} sampai {item.period_end} · {item.total_slots} slot
                                </div>
                            </div>
                            <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-neutral-500">
                                <span>Matched {item.matched}</span>
                                <span>Unmatched {item.unmatched}</span>
                                <span>Koreksi {item.corrected_count}</span>
                                <span>{formatImportedAt(item.created_at)}</span>
                            </div>
                        </div>
                    {/each}
                </div>
            {/if}
        </section>
    </div>
</AppLayout>
