package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/maulanashalihin/laju-go/app/config"
	"github.com/maulanashalihin/laju-go/app/models"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	emailFlag := flag.String("email", "superadmin@local.test", "local test super admin email")
	nameFlag := flag.String("name", "Super Admin Test", "local test super admin name")
	flag.Parse()

	cfg := config.Load()
	if strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "production") {
		fmt.Fprintln(os.Stderr, "refusing to create or promote a super admin in production")
		os.Exit(1)
	}

	email := strings.ToLower(strings.TrimSpace(*emailFlag))
	name := strings.TrimSpace(*nameFlag)
	if email == "" || !strings.Contains(email, "@") {
		fmt.Fprintln(os.Stderr, "invalid email")
		os.Exit(1)
	}
	if name == "" {
		name = "Super Admin Test"
	}

	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		fail(err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		fail(err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		fail(err)
	}

	var userID int64
	err = db.QueryRow("SELECT id FROM users WHERE lower(email) = lower(?)", email).Scan(&userID)
	now := time.Now().UTC()

	switch {
	case err == nil:
		result, updateErr := db.Exec(
			"UPDATE users SET role = ?, name = ?, updated_at = ? WHERE id = ?",
			string(models.RoleSuperAdmin), name, now, userID,
		)
		if updateErr != nil {
			fail(updateErr)
		}
		affected, affectedErr := result.RowsAffected()
		if affectedErr != nil || affected != 1 {
			if affectedErr != nil {
				fail(affectedErr)
			}
			fail(fmt.Errorf("expected one user row updated, got %d", affected))
		}
		fmt.Printf("SUPER_ADMIN_READY=YES\nUSER_ID=%d\nEMAIL=%s\nACTION=promoted_existing\n", userID, email)

	case err == sql.ErrNoRows:
		result, insertErr := db.Exec(
			"INSERT INTO users (email, name, password, avatar, role, email_verified, created_at, updated_at) VALUES (?, ?, NULL, '', ?, 1, ?, ?)",
			email, name, string(models.RoleSuperAdmin), now, now,
		)
		if insertErr != nil {
			fail(insertErr)
		}
		userID, insertErr = result.LastInsertId()
		if insertErr != nil {
			fail(insertErr)
		}
		fmt.Printf("SUPER_ADMIN_READY=YES\nUSER_ID=%d\nEMAIL=%s\nACTION=created_without_password\n", userID, email)
		fmt.Println("NEXT=use /forgot-password to create the first password")

	default:
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "ERROR=%v\n", err)
	os.Exit(1)
}
