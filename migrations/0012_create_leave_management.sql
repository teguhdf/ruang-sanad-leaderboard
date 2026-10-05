-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS leave_entitlements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    year INTEGER NOT NULL CHECK (year BETWEEN 2000 AND 2100),
    entitled_days INTEGER NOT NULL DEFAULT 0 CHECK (entitled_days >= 0),
    carryover_days INTEGER NOT NULL DEFAULT 0 CHECK (carryover_days >= 0),
    adjustment_days INTEGER NOT NULL DEFAULT 0,
    note TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    UNIQUE(owner_user_id, member_id, year)
);

CREATE INDEX IF NOT EXISTS idx_leave_entitlements_owner_year
ON leave_entitlements(owner_user_id, year, member_id);

CREATE TABLE IF NOT EXISTS leave_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_user_id INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    requested_by_user_id INTEGER NOT NULL,
    entitlement_year INTEGER NOT NULL CHECK (entitlement_year BETWEEN 2000 AND 2100),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    chargeable_days INTEGER NOT NULL CHECK (chargeable_days > 0),
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    reviewed_by_user_id INTEGER,
    reviewed_at DATETIME,
    review_note TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE,
    FOREIGN KEY (requested_by_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by_user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_leave_requests_owner_status
ON leave_requests(owner_user_id, status, start_date);

CREATE INDEX IF NOT EXISTS idx_leave_requests_member_year
ON leave_requests(member_id, entitlement_year, status, start_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_leave_requests_member_year;
DROP INDEX IF EXISTS idx_leave_requests_owner_status;
DROP TABLE IF EXISTS leave_requests;
DROP INDEX IF EXISTS idx_leave_entitlements_owner_year;
DROP TABLE IF EXISTS leave_entitlements;
-- +goose StatementEnd
