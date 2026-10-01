<script lang="ts">
    import { router, useForm } from "@inertiajs/svelte";
    import { fly } from "svelte/transition";
    import AppLayout from "@layouts/AppLayout.svelte";
    import type { Member, User } from "@lib/types";
    import {
        Briefcase,
        CalendarDays,
        Camera,
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
    let uploadingPhotoId = $state<number | null>(null);

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
        requestAnimationFrame(() => document.getElementById("member-form-card")?.scrollIntoView({ behavior: "smooth", block: "start" }));
    }

    function resetForm() {
        editingId = null;
        memberForm.reset();
        memberForm.clearErrors();
    }

    function startCreate() {
        resetForm();
        requestAnimationFrame(() => document.getElementById("member-form-card")?.scrollIntoView({ behavior: "smooth", block: "start" }));
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

    function initials(name: string) {
        return name.trim().split(/\s+/).slice(0, 2).map((part) => part[0]?.toUpperCase() ?? "").join("") || "?";
    }

    function uploadPhoto(memberId: number, event: Event) {
        const input = event.currentTarget as HTMLInputElement;
        const file = input.files?.[0];
        if (!file) return;
        if (!["image/jpeg", "image/png", "image/webp"].includes(file.type)) {
            alert("Format foto harus JPG, PNG, atau WEBP.");
            input.value = "";
            return;
        }
        if (file.size > 3 * 1024 * 1024) {
            alert("Ukuran foto maksimal 3 MB.");
            input.value = "";
            return;
        }
        const data = new FormData();
        data.append("file", file);
        uploadingPhotoId = memberId;
        router.post(`/app/members/${memberId}/photo`, data, {
            forceFormData: true,
            preserveScroll: true,
            onFinish: () => {
                uploadingPhotoId = null;
                input.value = "";
            },
        });
    }
</script>

<AppLayout {user} group="members">
    <div class="border-b border-neutral-200/80 dark:border-white/[0.04]">
        <div class="max-w-6xl mx-auto px-6 pt-8 pb-8">
            <div class="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
                <div>
                    <div class="flex items-center gap-2 text-sm font-medium text-emerald-700 mb-2"><Users class="w-4 h-4" /> Tim Ruang Sanad</div>
                    <h1 class="text-3xl font-bold tracking-tight text-neutral-900 dark:text-white">Anggota Tim</h1>
                    <p class="mt-2 text-neutral-600 dark:text-neutral-400 max-w-2xl">Simpan identitas anggota, foto profil, jadwal kerja tetap, ID fingerprint, dan workspace Google Sheets.</p>
                </div>
                <div class="flex items-center gap-3">
                    <button type="button" onclick={startCreate} class="inline-flex min-h-11 items-center gap-2 rounded-xl bg-emerald-700 px-4 text-sm font-bold text-white hover:bg-emerald-800"><Plus class="w-4 h-4" /> Tambah anggota baru</button>
                    <div class="rounded-xl border border-neutral-200 dark:border-white/[0.06] bg-white dark:bg-neutral-925 px-4 py-3 min-w-32"><div class="text-xs uppercase tracking-wider text-neutral-500">Total</div><div class="text-2xl font-bold text-neutral-900 dark:text-white mt-1">{members.length}</div></div>
                </div>
            </div>
        </div>
    </div>

    <div class="max-w-6xl mx-auto px-6 py-8 space-y-8">
        {#if success}<div class="rounded-xl border border-green-500/20 bg-green-500/10 px-4 py-3 text-sm font-medium text-green-700" in:fly={{ y: -8, duration: 220 }}>{success}</div>{/if}
        {#if error}<div class="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm font-medium text-red-700" in:fly={{ y: -8, duration: 220 }}>{error}</div>{/if}

        <section id="member-form-card" class="scroll-mt-24 rounded-2xl border border-neutral-200/80 dark:border-white/[0.06] bg-white dark:bg-neutral-925/60 overflow-hidden">
            <div class="px-6 py-5 border-b border-neutral-200/80 dark:border-white/[0.05] flex items-center justify-between gap-4">
                <div><h2 class="text-lg font-semibold text-neutral-900 dark:text-white">{editingId ? "Edit anggota" : "Tambah anggota baru"}</h2><p class="text-sm text-neutral-500 mt-1">{editingId ? "Sedang mengedit anggota. Klik Tambah anggota baru untuk keluar dari mode edit." : "Isi data anggota. Foto dapat ditambahkan setelah anggota tersimpan."}</p></div>
                {#if editingId}<button type="button" onclick={startCreate} class="inline-flex items-center gap-1.5 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm font-bold text-emerald-800"><Plus class="w-4 h-4" /> Tambah anggota baru</button>{/if}
            </div>

            <form onsubmit={submitMember} class="p-6 grid gap-5 md:grid-cols-2">
                <label class="space-y-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Nama anggota</span><input bind:value={memberForm.name} type="text" autocomplete="off" placeholder="Contoh: Fathiyah" class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/15" /></label>
                <label class="space-y-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Divisi</span><input bind:value={memberForm.division} type="text" autocomplete="off" placeholder="Contoh: SMM" class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/15" /></label>

                <div class="space-y-2 md:col-span-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Hari kerja tetap</span><div class="flex flex-wrap gap-2">{#each workDayOptions as day}<button type="button" onclick={() => toggleWorkDay(day.value)} class="min-w-12 rounded-lg border px-3 py-2 text-sm font-semibold transition-colors {memberForm.work_days.includes(day.value) ? 'border-emerald-400 bg-emerald-50 text-emerald-800' : 'border-neutral-300 text-neutral-500 hover:border-neutral-400'}">{day.label}</button>{/each}</div></div>

                <label class="space-y-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Jam masuk</span><input bind:value={memberForm.start_time} type="time" class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/15" /></label>
                <label class="space-y-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">ID fingerprint</span><input bind:value={memberForm.fingerprint_id} type="text" autocomplete="off" placeholder="Sesuai ID pada mesin fingerprint" class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/15" /></label>
                <label class="space-y-2 md:col-span-2"><span class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Link Google Sheets workspace</span><input bind:value={memberForm.google_sheet_url} type="url" placeholder="https://docs.google.com/spreadsheets/d/..." class="w-full rounded-xl border border-neutral-300 dark:border-neutral-700 bg-neutral-50 dark:bg-neutral-900 px-4 py-3 text-neutral-900 dark:text-white outline-none focus:border-emerald-400 focus:ring-2 focus:ring-emerald-400/15" /></label>

                <div class="md:col-span-2 flex justify-end gap-2">
                    {#if editingId}<button type="button" onclick={startCreate} class="inline-flex items-center gap-2 rounded-xl border border-neutral-300 px-5 py-3 font-semibold text-neutral-600"><X class="w-4 h-4" /> Batal edit</button>{/if}
                    <button type="submit" disabled={memberForm.processing} class="inline-flex items-center gap-2 rounded-xl bg-emerald-700 hover:bg-emerald-800 px-5 py-3 font-semibold text-white transition-colors disabled:opacity-50">{#if editingId}<Save class="w-4 h-4" />{memberForm.processing ? "Menyimpan..." : "Simpan perubahan"}{:else}<Plus class="w-4 h-4" />{memberForm.processing ? "Menyimpan..." : "Tambah anggota"}{/if}</button>
                </div>
            </form>
        </section>

        <section>
            <div class="flex items-center justify-between mb-4"><div><h2 class="text-lg font-semibold text-neutral-900 dark:text-white">Daftar anggota</h2><p class="text-sm text-neutral-500 mt-1">Foto digunakan pada poster leaderboard. Tanpa foto, avatar inisial tetap digunakan.</p></div><button type="button" onclick={startCreate} class="hidden sm:inline-flex items-center gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm font-bold text-emerald-800"><Plus class="w-4 h-4" /> Tambah anggota</button></div>

            {#if members.length === 0}
                <div class="rounded-2xl border border-dashed border-neutral-300 py-14 px-6 text-center"><Users class="w-8 h-8 mx-auto text-neutral-400 mb-3" /><h3 class="font-semibold text-neutral-800">Belum ada anggota</h3><p class="text-sm text-neutral-500 mt-1">Tambahkan anggota pertama.</p></div>
            {:else}
                <div class="grid gap-4 lg:grid-cols-2">
                    {#each members as member (member.id)}
                        <article class="rounded-2xl border border-neutral-200/80 bg-white p-5 dark:bg-neutral-925/60 dark:border-white/[0.06]">
                            <div class="flex items-start justify-between gap-4">
                                <div class="flex min-w-0 items-center gap-3"><div class="grid h-16 w-16 shrink-0 place-items-center overflow-hidden rounded-full bg-emerald-50 text-lg font-black text-emerald-800 ring-2 ring-white shadow-sm">{#if member.photo_url}<img src={member.photo_url} alt={`Foto ${member.name}`} class="h-full w-full object-cover" />{:else}{initials(member.name)}{/if}</div><div class="min-w-0"><h3 class="text-lg font-semibold text-neutral-900 dark:text-white truncate">{member.name}</h3><div class="mt-1 inline-flex items-center gap-1.5 text-sm text-neutral-500"><Briefcase class="w-4 h-4" />{member.division}</div></div></div>
                                <div class="flex shrink-0 gap-2"><label class="inline-flex cursor-pointer items-center gap-1.5 rounded-lg border border-neutral-200 px-3 py-2 text-sm font-medium text-neutral-600 hover:border-emerald-400/50 hover:text-emerald-700"><Camera class="w-4 h-4" />{uploadingPhotoId === member.id ? "Upload..." : "Upload Foto"}<input type="file" accept="image/jpeg,image/png,image/webp" class="hidden" disabled={uploadingPhotoId !== null} onchange={(event) => uploadPhoto(member.id, event)} /></label><button type="button" onclick={() => startEdit(member)} class="inline-flex items-center gap-1.5 rounded-lg border border-neutral-200 px-3 py-2 text-sm font-medium text-neutral-600 hover:border-emerald-400/50 hover:text-emerald-700"><Pencil class="w-4 h-4" /> Edit</button></div>
                            </div>
                            <div class="mt-5 grid gap-3 sm:grid-cols-2 text-sm"><div class="flex items-start gap-2 text-neutral-600"><CalendarDays class="w-4 h-4 mt-0.5 shrink-0" /><span>{member.work_days.map(dayLabel).join(", ")}</span></div><div class="flex items-start gap-2 text-neutral-600"><Clock class="w-4 h-4 mt-0.5 shrink-0" /><span>Masuk {member.start_time}</span></div><div class="flex items-start gap-2 text-neutral-600"><Fingerprint class="w-4 h-4 mt-0.5 shrink-0" /><span class="break-all">{member.fingerprint_id}</span></div><a href={member.google_sheet_url} target="_blank" rel="noopener" class="flex items-start gap-2 text-emerald-700 hover:underline"><ExternalLink class="w-4 h-4 mt-0.5 shrink-0" /> Buka workspace</a></div>
                        </article>
                    {/each}
                </div>
            {/if}
        </section>
    </div>
</AppLayout>
