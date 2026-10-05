/**
 * Shared frontend types.
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
	photo_url: string;
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

export interface AttendanceDayPlan {
	work_date: string;
	day_type: "work" | "off" | "replacement_off" | "holiday" | "sick" | "leave";
	shift_template_id?: number;
	work_mode: "onsite" | "remote" | "hybrid";
	required_minutes: number;
	source: "default" | "roster" | "approved_leave";
}

export interface MemberAttendanceWeek {
	member_id: number;
	member_name: string;
	division: string;
	week_start: string;
	week_end: string;
	plan: {
		days: AttendanceDayPlan[];
		required_minutes: number;
	};
}

export interface AttendanceFoundation {
	policy: AttendancePolicy;
	profiles: MemberAttendanceProfile[];
	shift_templates: ShiftTemplate[];
	roster_week_start: string;
	roster_week_end: string;
	roster_weeks: MemberAttendanceWeek[];
}

export interface EmployeeAttendanceIdentity {
	member_id: number;
	member_name: string;
	division: string;
	work_mode: "onsite" | "remote" | "hybrid";
	schedule_type: "shift" | "flexible";
	weekly_target_minutes: number;
}

export interface AttendanceSession {
	id: number;
	member_id: number;
	work_date: string;
	location_mode: "office" | "remote";
	check_in_at: string;
	check_out_at?: string;
	status: "active" | "completed" | "incomplete" | "superseded";
	source: "app" | "manual_approved";
	duration_minutes: number;
}

export interface AttendanceAdjustmentRequest {
	id: number;
	member_id: number;
	member_name: string;
	work_date: string;
	location_mode: "office" | "remote";
	requested_check_in: string;
	requested_check_out: string;
	reason: string;
	status: "pending" | "approved" | "rejected";
	review_note: string;
	reviewed_at?: string;
	created_at: string;
}

export interface OvertimeRequest {
	id: number;
	member_id: number;
	work_date: string;
	requested_minutes: number;
	reason: string;
	status: "pending" | "approved" | "rejected";
	review_note: string;
	approved_at?: string;
	created_at: string;
}

export interface MyAttendanceSummary {
	identity: EmployeeAttendanceIdentity;
	today: string;
	active_session?: AttendanceSession;
	today_sessions: AttendanceSession[];
	today_minutes: number;
	week_start: string;
	week_end: string;
	weekly_minutes: number;
	weekly_target_minutes: number;
	weekly_balance_minutes: number;
	adjustments: AttendanceAdjustmentRequest[];
	overtime_requests: OvertimeRequest[];
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
