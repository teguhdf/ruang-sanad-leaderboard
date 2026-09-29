<script lang="ts">
	import { router } from "@inertiajs/svelte";
	import AppLayout from "@layouts/AppLayout.svelte";
	import type { AttendanceFoundation, User } from "@lib/types";
	import {
		BadgeCheck,
		BriefcaseBusiness,
		Building2,
		CalendarClock,
		Clock3,
		Coffee,
		Laptop,
		Save,
		ShieldCheck,
		Sparkles,
		TimerReset,
		UsersRound,
	} from "lucide-svelte";

	interface Props {
		user?: User;
		foundation: AttendanceFoundation;
		success?: string;
		error?: string;
	}

	let { user, foundation, success, error }: Props = $props();

	let drafts = $state<Record<number, { work_mode: string; schedule_type: string }>>(
		Object.fromEntries(
			foundation.profiles.map((profile) => [
				profile.member_id,
				{ work_mode: profile.work_mode, schedule_type: profile.schedule_type },
			]),
		),
	);
	let savingMemberID = $state<number | null>(null);

	function formatMinutes(total: number) {
		const hours = Math.floor(total / 60);
		const minutes = total % 60;
		return minutes === 0 ? `${hours} jam` : `${hours}j ${minutes}m`;
	}

	function updateDraft(memberID: number, field: "work_mode" | "schedule_type", value: string) {
		drafts[memberID] = {
			...drafts[memberID],
			[field]: value,
		};
	}

	function saveProfile(memberID: number) {
		savingMemberID = memberID;
		router.put(`/app/attendance/profiles/${memberID}`, drafts[memberID], {
			preserveScroll: true,
			onFinish: () => {
				savingMemberID = null;
			},
		});
	}

	function locationLabel(location: string) {
		return location === "office" ? "Kantor" : "Remote";
	}
</script>

<AppLayout {user} group="attendance">
	<div class="md-page">
		<section class="md-hero">
			<div>
				<div class="md-eyebrow">
					<Sparkles size="16" />
					Sprint 1 · Attendance Foundation
				</div>
				<h1>Fondasi absensi yang mengikuti cara tim benar-benar bekerja.</h1>
				<p>
					Shift kantor, remote fleksibel, Sabtu setengah hari, pengganti kerja Ahad, carry-over mingguan, dan lembur dengan approval SPV disiapkan dalam satu model.
				</p>
			</div>
			<div class="md-hero-badge">
				<BadgeCheck size="22" />
				<div>
					<strong>Policy aktif</strong>
					<span>44 jam / minggu</span>
				</div>
			</div>
		</section>

		{#if success}
			<div class="md-alert md-alert-success">{success}</div>
		{/if}
		{#if error}
			<div class="md-alert md-alert-error">{error}</div>
		{/if}

		<section class="md-stat-grid">
			<article class="md-stat-card">
				<div class="md-icon-box"><Clock3 size="20" /></div>
				<span>Senin - Jumat</span>
				<strong>{formatMinutes(foundation.policy.weekday_target_minutes)}</strong>
				<small>Target harian normal</small>
			</article>
			<article class="md-stat-card">
				<div class="md-icon-box"><Coffee size="20" /></div>
				<span>Sabtu</span>
				<strong>{formatMinutes(foundation.policy.saturday_target_minutes)}</strong>
				<small>Setengah hari</small>
			</article>
			<article class="md-stat-card md-stat-primary">
				<div class="md-icon-box"><CalendarClock size="20" /></div>
				<span>Target mingguan</span>
				<strong>{formatMinutes(foundation.policy.weekly_target_minutes)}</strong>
				<small>Carry-over antarhari aktif</small>
			</article>
			<article class="md-stat-card">
				<div class="md-icon-box"><ShieldCheck size="20" /></div>
				<span>Lembur</span>
				<strong>Approval SPV</strong>
				<small>Terpisah dari jam normal</small>
			</article>
		</section>

		<section class="md-card">
			<div class="md-section-head">
				<div>
					<div class="md-eyebrow"><TimerReset size="16" /> Aturan waktu</div>
					<h2>Policy perusahaan</h2>
					<p>Rule ini menjadi dasar engine absensi di sprint berikutnya.</p>
				</div>
			</div>
			<div class="md-chip-row">
				<span class="md-chip">Toleransi telat {foundation.policy.late_tolerance_minutes} menit</span>
				<span class="md-chip">Toleransi pulang cepat {foundation.policy.early_leave_tolerance_minutes} menit</span>
				<span class="md-chip">Carry-over hanya dalam minggu yang sama</span>
				<span class="md-chip">Ahad kerja = roster + OFF pengganti</span>
				<span class="md-chip md-chip-accent">Lembur wajib approval SPV</span>
			</div>
			<div class="md-note">
				<strong>Pemisahan penting:</strong> kekurangan jam yang diganti pada hari berikutnya tetap dihitung sebagai pemenuhan target reguler. Jam baru disebut lembur jika pekerjaan tambahan memang disetujui sebagai overtime, sehingga tidak mencampur “bayar hutang jam” dengan lembur perusahaan.
			</div>
		</section>

		<section class="md-card">
			<div class="md-section-head">
				<div>
					<div class="md-eyebrow"><BriefcaseBusiness size="16" /> Template shift</div>
					<h2>Pola kerja operasional</h2>
					<p>Template awal mengikuti pola shift yang sudah dikunci.</p>
				</div>
			</div>

			<div class="md-shift-grid">
				{#each foundation.shift_templates as shift}
					<article class="md-shift-card">
						<div class="md-shift-title">
							<div>
								<span>{shift.name}</span>
								<strong>{formatMinutes(shift.required_minutes)}</strong>
							</div>
							<span class="md-status">Aktif</span>
						</div>
						<p>{shift.description}</p>
						<div class="md-segment-list">
							{#each shift.segments as segment}
								<div class="md-segment-row">
									<div class="md-segment-icon">
										{#if segment.location_mode === "office"}
											<Building2 size="17" />
										{:else}
											<Laptop size="17" />
										{/if}
									</div>
									<div>
										<strong>{locationLabel(segment.location_mode)}</strong>
										<span>
											{#if segment.is_flexible}
												Fleksibel · {formatMinutes(segment.target_minutes)}
											{:else}
												{segment.start_time} - {segment.end_time} · {formatMinutes(segment.target_minutes)}
											{/if}
										</span>
									</div>
								</div>
							{/each}
						</div>
					</article>
				{/each}
			</div>
		</section>

		<section class="md-card">
			<div class="md-section-head md-section-head-split">
				<div>
					<div class="md-eyebrow"><UsersRound size="16" /> Profil absensi anggota</div>
					<h2>Tentukan mode kerja tiap anggota</h2>
					<p>Roster per tanggal akan dibangun pada sprint roster. Di sini kita kunci karakter kerja dasarnya.</p>
				</div>
				<span class="md-count">{foundation.profiles.length} anggota</span>
			</div>

			{#if foundation.profiles.length === 0}
				<div class="md-empty">Belum ada anggota. Tambahkan anggota lebih dulu dari menu Anggota Tim.</div>
			{:else}
				<div class="md-profile-list">
					{#each foundation.profiles as profile}
						<article class="md-profile-row">
							<div class="md-avatar">{profile.member_name.charAt(0).toUpperCase()}</div>
							<div class="md-profile-info">
								<strong>{profile.member_name}</strong>
								<span>{profile.division}</span>
							</div>
							<label>
								<span>Mode kerja</span>
								<select
									value={drafts[profile.member_id]?.work_mode}
									onchange={(event) => updateDraft(profile.member_id, "work_mode", event.currentTarget.value)}
								>
									<option value="onsite">Onsite</option>
									<option value="remote">Remote</option>
									<option value="hybrid">Hybrid</option>
								</select>
							</label>
							<label>
								<span>Tipe jadwal</span>
								<select
									value={drafts[profile.member_id]?.schedule_type}
									onchange={(event) => updateDraft(profile.member_id, "schedule_type", event.currentTarget.value)}
								>
									<option value="shift">Shift / roster</option>
									<option value="flexible">Fleksibel</option>
								</select>
							</label>
							<div class="md-target">
								<span>Target</span>
								<strong>{formatMinutes(profile.weekly_target_minutes)}/minggu</strong>
							</div>
							<button
								type="button"
								class="md-button md-button-tonal"
								disabled={savingMemberID === profile.member_id}
								onclick={() => saveProfile(profile.member_id)}
							>
								<Save size="16" />
								{savingMemberID === profile.member_id ? "Menyimpan" : "Simpan"}
							</button>
						</article>
					{/each}
				</div>
			{/if}
		</section>
	</div>
</AppLayout>
