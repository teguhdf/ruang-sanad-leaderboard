-- +goose Up
-- +goose StatementBegin
UPDATE users
SET role = 'spv', updated_at = CURRENT_TIMESTAMP
WHERE role = 'user'
  AND id IN (SELECT DISTINCT owner_user_id FROM members);

CREATE TABLE IF NOT EXISTS attendance_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    work_date TEXT NOT NULL,
    location_mode TEXT NOT NULL CHECK (location_mode IN ('office', 'remote')),
    check_in_at DATETIME NOT NULL,
    check_out_at DATETIME,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'incomplete', 'superseded')),
    source TEXT NOT NULL DEFAULT 'app' CHECK (source IN ('app', 'manual_approved')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_attendance_sessions_one_active
ON attendance_sessions(member_id)
WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_member_date
ON attendance_sessions(member_id, work_date);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_owner_date
ON attendance_sessions(owner_user_id, work_date);

CREATE TABLE IF NOT EXISTS attendance_adjustment_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    requested_by_user_id INTEGER NOT NULL,
    related_session_id INTEGER,
    work_date TEXT NOT NULL,
    location_mode TEXT NOT NULL CHECK (location_mode IN ('office', 'remote')),
    requested_check_in DATETIME NOT NULL,
    requested_check_out DATETIME NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by_user_id INTEGER,
    reviewed_at DATETIME,
    review_note TEXT NOT NULL DEFAULT '',
    created_session_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (requested_by_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (related_session_id) REFERENCES attendance_sessions(id) ON DELETE SET NULL,
    FOREIGN KEY (reviewed_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_session_id) REFERENCES attendance_sessions(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_attendance_adjustments_owner_status
ON attendance_adjustment_requests(owner_user_id, status, work_date);

CREATE INDEX IF NOT EXISTS idx_attendance_adjustments_member_date
ON attendance_adjustment_requests(member_id, work_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_attendance_adjustments_member_date;
DROP INDEX IF EXISTS idx_attendance_adjustments_owner_status;
DROP TABLE IF EXISTS attendance_adjustment_requests;

DROP INDEX IF EXISTS idx_attendance_sessions_owner_date;
DROP INDEX IF EXISTS idx_attendance_sessions_member_date;
DROP INDEX IF EXISTS idx_attendance_sessions_one_active;
DROP TABLE IF EXISTS attendance_sessions;
-- +goose StatementEnd
