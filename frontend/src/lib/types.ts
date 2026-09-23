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

/**
 * Common flash message shape from Inertia.
 * Backend sets via `s.store.Flash(c, "error", ...)` and `inertiaService.Render`
 * merges into `props.flash` automatically.
 */
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
