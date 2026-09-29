/**
 * Shared frontend types.
 *
 * Must stay in sync with Go backend `app/models/dto.go` UserResponse.
 */

export interface User {
	id: number;
	email: string;
	name: string;
	avatar: string;
	role: string;
	email_verified: boolean;
}

export interface Flash {
	error?: string;
	success?: string;
	warning?: string;
	info?: string;
}

export interface Member {
	id: number;
	name: string;
	division: string;
	work_days: string[];
	start_time: string;
	fingerprint_id: string;
	google_sheet_url: string;
	created_at: string;
	updated_at: string;
}

export interface AttendancePolicy {
	id: number;
	weekday_target_minutes: number;
	saturday_target_minutes: number;
	weekly_target_minutes: number;
	late_tolerance_minutes: number;
	early_leave_tolerance_minutes: number;
	allow_weekly_carryover: boolean;
	overtime_requires_approval: boolean;
	week_starts_on: number;
	created_at: string;
	updated_at: string;
}

export interface MemberAttendanceProfile {
	id: number;
	member_id: number;
	member_name: string;
	division: string;
	linked_user_id?: number;
	work_mode: "onsite" | "remote" | "hybrid";
	schedule_type: "shift" | "flexible";
	required_daily_minutes: number;
	weekly_target_minutes: number;
	attendance_enabled: boolean;
}

export interface ShiftSegment {
	id: number;
	sort_order: number;
	location_mode: "office" | "remote";
	start_time?: string;
	end_time?: string;
	target_minutes: number;
	is_flexible: boolean;
}

export interface ShiftTemplate {
	id: number;
	code: string;
	name: string;
	required_minutes: number;
	description: string;
	is_active: boolean;
	segments: ShiftSegment[];
}

export interface AttendanceFoundation {
	policy: AttendancePolicy;
	profiles: MemberAttendanceProfile[];
	shift_templates: ShiftTemplate[];
}

export interface FingerprintDayPreview {
	day: number;
	date: string;
	raw_value: string;
	times: string[];
	read_state: "empty" | "one_time" | "two_times" | "multiple_times";
}

export interface FingerprintPersonPreview {
	fingerprint_id: string;
	file_name: string;
	department: string;
	matched: boolean;
	member_id?: number;
	member_name?: string;
	days: FingerprintDayPreview[];
	summary: {
		two_times: number;
		one_time: number;
		empty: number;
		multiple_times: number;
	};
}

export interface FingerprintPreview {
	file_name: string;
	period_start: string;
	period_end: string;
	people: FingerprintPersonPreview[];
	summary: {
		file_people: number;
		matched: number;
		unmatched: number;
		total_slots: number;
		two_times: number;
		one_time: number;
		empty: number;
		multiple_times: number;
	};
}

export interface FingerprintImportHistory {
	id: number;
	file_name: string;
	period_start: string;
	period_end: string;
	file_people: number;
	matched: number;
	unmatched: number;
	total_slots: number;
	two_times: number;
	one_time: number;
	empty: number;
	multiple_times: number;
	corrected_count: number;
	created_at: string;
}
