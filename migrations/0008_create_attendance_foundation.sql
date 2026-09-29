-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS attendance_policies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL UNIQUE,
    weekday_target_minutes INTEGER NOT NULL DEFAULT 480,
    saturday_target_minutes INTEGER NOT NULL DEFAULT 240,
    weekly_target_minutes INTEGER NOT NULL DEFAULT 2640,
    late_tolerance_minutes INTEGER NOT NULL DEFAULT 5,
    early_leave_tolerance_minutes INTEGER NOT NULL DEFAULT 5,
    allow_weekly_carryover BOOLEAN NOT NULL DEFAULT TRUE,
    overtime_requires_approval BOOLEAN NOT NULL DEFAULT TRUE,
    week_starts_on INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS member_attendance_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL UNIQUE,
    linked_user_id INTEGER,
    work_mode TEXT NOT NULL DEFAULT 'onsite',
    schedule_type TEXT NOT NULL DEFAULT 'shift',
    required_daily_minutes INTEGER NOT NULL DEFAULT 480,
    weekly_target_minutes INTEGER NOT NULL DEFAULT 2640,
    attendance_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (linked_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_member_attendance_profiles_linked_user
ON member_attendance_profiles(linked_user_id)
WHERE linked_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_member_attendance_profiles_owner
ON member_attendance_profiles(owner_user_id);

CREATE TABLE IF NOT EXISTS shift_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    required_minutes INTEGER NOT NULL DEFAULT 480,
    description TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(owner_user_id, code)
);

CREATE INDEX IF NOT EXISTS idx_shift_templates_owner
ON shift_templates(owner_user_id);

CREATE TABLE IF NOT EXISTS shift_segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    shift_template_id INTEGER NOT NULL,
    sort_order INTEGER NOT NULL,
    location_mode TEXT NOT NULL,
    start_time TEXT,
    end_time TEXT,
    target_minutes INTEGER NOT NULL,
    is_flexible BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (shift_template_id) REFERENCES shift_templates(id) ON DELETE CASCADE,
    UNIQUE(shift_template_id, sort_order)
);

CREATE TABLE IF NOT EXISTS rosters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    work_date TEXT NOT NULL,
    day_type TEXT NOT NULL DEFAULT 'work',
    shift_template_id INTEGER,
    work_mode TEXT NOT NULL DEFAULT 'onsite',
    required_minutes INTEGER NOT NULL DEFAULT 480,
    note TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (shift_template_id) REFERENCES shift_templates(id) ON DELETE SET NULL,
    UNIQUE(member_id, work_date)
);

CREATE INDEX IF NOT EXISTS idx_rosters_owner_date
ON rosters(owner_user_id, work_date);

CREATE INDEX IF NOT EXISTS idx_rosters_member_date
ON rosters(member_id, work_date);

CREATE TABLE IF NOT EXISTS overtime_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    work_date TEXT NOT NULL,
    requested_minutes INTEGER NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_by_user_id INTEGER,
    approved_by_user_id INTEGER,
    approved_at DATETIME,
    review_note TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (approved_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_overtime_requests_owner_status
ON overtime_requests(owner_user_id, status, work_date);

CREATE INDEX IF NOT EXISTS idx_overtime_requests_member_date
ON overtime_requests(member_id, work_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_overtime_requests_member_date;
DROP INDEX IF EXISTS idx_overtime_requests_owner_status;
DROP TABLE IF EXISTS overtime_requests;

DROP INDEX IF EXISTS idx_rosters_member_date;
DROP INDEX IF EXISTS idx_rosters_owner_date;
DROP TABLE IF EXISTS rosters;

DROP TABLE IF EXISTS shift_segments;
DROP INDEX IF EXISTS idx_shift_templates_owner;
DROP TABLE IF EXISTS shift_templates;

DROP INDEX IF EXISTS idx_member_attendance_profiles_owner;
DROP INDEX IF EXISTS idx_member_attendance_profiles_linked_user;
DROP TABLE IF EXISTS member_attendance_profiles;
DROP TABLE IF EXISTS attendance_policies;
-- +goose StatementEnd
