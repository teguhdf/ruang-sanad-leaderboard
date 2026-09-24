-- +goose Up
-- +goose StatementBegin
CREATE TABLE fingerprint_imports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    file_name TEXT NOT NULL,
    file_sha256 TEXT NOT NULL,
    period_start TEXT NOT NULL,
    period_end TEXT NOT NULL,
    file_people INTEGER NOT NULL DEFAULT 0,
    matched INTEGER NOT NULL DEFAULT 0,
    unmatched INTEGER NOT NULL DEFAULT 0,
    total_slots INTEGER NOT NULL DEFAULT 0,
    two_times INTEGER NOT NULL DEFAULT 0,
    one_time INTEGER NOT NULL DEFAULT 0,
    empty INTEGER NOT NULL DEFAULT 0,
    multiple_times INTEGER NOT NULL DEFAULT 0,
    corrected_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(owner_user_id, file_sha256)
);

CREATE INDEX idx_fingerprint_imports_owner_created
    ON fingerprint_imports(owner_user_id, created_at DESC);

CREATE TABLE fingerprint_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    import_id INTEGER NOT NULL,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER,
    fingerprint_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    department TEXT NOT NULL DEFAULT '',
    attendance_date TEXT NOT NULL,
    day INTEGER NOT NULL,
    raw_value TEXT NOT NULL DEFAULT '',
    raw_check_in TEXT NOT NULL DEFAULT '',
    raw_check_out TEXT NOT NULL DEFAULT '',
    corrected_check_in TEXT NOT NULL DEFAULT '',
    corrected_check_out TEXT NOT NULL DEFAULT '',
    correction_reason TEXT NOT NULL DEFAULT '',
    is_corrected INTEGER NOT NULL DEFAULT 0,
    read_state TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (import_id) REFERENCES fingerprint_imports(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL,
    UNIQUE(import_id, fingerprint_id, attendance_date)
);

CREATE INDEX idx_fingerprint_records_import_id
    ON fingerprint_records(import_id);

CREATE INDEX idx_fingerprint_records_owner_member_date
    ON fingerprint_records(owner_user_id, member_id, attendance_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_fingerprint_records_owner_member_date;
DROP INDEX IF EXISTS idx_fingerprint_records_import_id;
DROP TABLE IF EXISTS fingerprint_records;
DROP INDEX IF EXISTS idx_fingerprint_imports_owner_created;
DROP TABLE IF EXISTS fingerprint_imports;
-- +goose StatementEnd
