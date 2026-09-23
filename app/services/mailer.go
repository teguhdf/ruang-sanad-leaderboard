package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/maulanashalihin/laju-go/app/queries"
)

// MailDriver selects the email backend.
type MailDriver string

const (
	MailDriverLog      MailDriver = "log"
	MailDriverResend   MailDriver = "resend"
	MailDriverMailtrap MailDriver = "mailtrap"
)

// MailConfig holds mailer configuration.
type MailConfig struct {
	Driver          MailDriver
	From            string
	FromName        string
	ResendAPIKey    string
	MailtrapToken   string
	MailtrapInboxID string
}

// MailMessage is a single email to send.
type MailMessage struct {
	To      string
	Subject string
	Text    string
	HTML    string
}

// MailerService handles email sending and password reset tokens.
// The mail driver is selected via MAIL_DRIVER env var:
//   - log:      print to console + record in SentMails (dev / tests)
//   - resend:   https://resend.com (RESEND_API_KEY)
//   - mailtrap: https://mailtrap.io (MAILTRAP_API_TOKEN, optional MAILTRAP_INBOX_ID)
type MailerService struct {
	querier    *queries.Querier
	appURL     string
	config     MailConfig
	httpClient *http.Client
}

// SentMails records mail sent via the log driver (assertable in dev and tests).
var (
	sentMails   []MailMessage
	sentMailsMu sync.Mutex
)

// ResetTokenEntry is a validated password reset token.
type ResetTokenEntry struct {
	UserID    int64
	Email     string
	Token     string
	ExpiresAt time.Time
}

// NewMailerService creates a new MailerService with the given config.
func NewMailerService(querier *queries.Querier, cfg MailConfig, appURL string) *MailerService {
	return &MailerService{
		querier:    querier,
		appURL:     appURL,
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendMail dispatches a message via the configured driver.
func (m *MailerService) SendMail(msg MailMessage) error {
	switch m.config.Driver {
	case MailDriverLog:
		sentMailsMu.Lock()
		sentMails = append(sentMails, msg)
		sentMailsMu.Unlock()
		slog.Info("mail (log driver)", "to", msg.To, "subject", msg.Subject)
		fmt.Println(formatMailLog(msg))
		return nil
	case MailDriverResend:
		return m.postJSON("https://api.resend.com/emails", m.config.ResendAPIKey, map[string]any{
			"from":    m.fromHeader(),
			"to":      []string{msg.To},
			"subject": msg.Subject,
			"text":    msg.Text,
			"html":    msg.HTML,
		})
	case MailDriverMailtrap:
		url := "https://send.api.mailtrap.io/api/send"
		if m.config.MailtrapInboxID != "" {
			url = fmt.Sprintf("https://sandbox.api.mailtrap.io/api/send/%s", m.config.MailtrapInboxID)
		}
		return m.postJSON(url, m.config.MailtrapToken, map[string]any{
			"from":    map[string]string{"email": m.config.From},
			"to":      []map[string]string{{"email": msg.To}},
			"subject": msg.Subject,
			"text":    msg.Text,
			"html":    msg.HTML,
		})
	default:
		return fmt.Errorf("unsupported mail driver: %s", m.config.Driver)
	}
}

// GetSentMails returns mail sent via the log driver (for tests).
func GetSentMails() []MailMessage {
	sentMailsMu.Lock()
	defer sentMailsMu.Unlock()
	cp := make([]MailMessage, len(sentMails))
	copy(cp, sentMails)
	return cp
}

// ResetSentMails clears the log driver's sent mail buffer (for tests).
func ResetSentMails() {
	sentMailsMu.Lock()
	sentMails = nil
	sentMailsMu.Unlock()
}

func (m *MailerService) fromHeader() string {
	if m.config.FromName != "" {
		return fmt.Sprintf("%s <%s>", m.config.FromName, m.config.From)
	}
	return m.config.From
}

func (m *MailerService) postJSON(url, token string, body any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal mail body: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("create mail request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 500))
		return fmt.Errorf("mail provider error %d: %s", resp.StatusCode, string(detail))
	}
	return nil
}

func formatMailLog(msg MailMessage) string {
	body := msg.HTML
	if body == "" {
		body = msg.Text
	}
	body = strings.ReplaceAll(body, "\n", "\n│ ")
	return fmt.Sprintf("┌─ mail (log driver) ────────────────────────────\n│ to:      %s\n│ subject: %s\n│ %s\n└────────────────────────────────────────────────",
		msg.To, msg.Subject, body)
}

// ── Password Reset ──────────────────────────────────────────────

// SendPasswordResetEmail generates a token, stores it in DB, and sends the reset email.
func (m *MailerService) SendPasswordResetEmail(ctx context.Context, email string, userID int64) error {
	token, err := generateResetToken()
	if err != nil {
		return err
	}

	if err := m.querier.CreatePasswordReset(ctx, hashToken(token), userID, email, time.Now().Add(1*time.Hour)); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	resetURL := fmt.Sprintf("%s/reset-password/%s", m.appURL, token)

	subject := "Reset Your Password"
	body := fmt.Sprintf(passwordResetEmailHTML, resetURL, resetURL)

	return m.SendMail(MailMessage{
		To:      email,
		Subject: subject,
		Text:    fmt.Sprintf("Reset your password: %s", resetURL),
		HTML:    body,
	})
}

// ValidateResetToken validates a reset token against the database.
func (m *MailerService) ValidateResetToken(ctx context.Context, token string) (*ResetTokenEntry, error) {
	pr, err := m.querier.GetPasswordReset(ctx, hashToken(token))
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token")
	}

	return &ResetTokenEntry{
		UserID:    pr.UserID,
		Email:     pr.Email,
		Token:     pr.Token,
		ExpiresAt: pr.ExpiresAt,
	}, nil
}

// InvalidateResetToken marks a reset token as used.
func (m *MailerService) InvalidateResetToken(ctx context.Context, token string) {
	if err := m.querier.MarkPasswordResetUsed(ctx, hashToken(token)); err != nil {
		slog.Error("failed to invalidate reset token", "error", err)
	}
}

// CleanupExpiredTokens removes expired tokens from the database.
func (m *MailerService) CleanupExpiredTokens(ctx context.Context) {
	if err := m.querier.DeleteExpiredPasswordResets(ctx); err != nil {
		slog.Error("cleanup: failed to delete expired password resets", "error", err)
	}
}

// ── Email Verification ──────────────────────────────────────────

// SendVerificationEmail generates a verification token, stores it in DB,
// and sends the verification email.
func (m *MailerService) SendVerificationEmail(ctx context.Context, email string, userID int64) error {
	token, err := generateResetToken()
	if err != nil {
		return err
	}

	if err := m.querier.CreateEmailVerification(ctx, hashToken(token), userID, email, time.Now().Add(24*time.Hour)); err != nil {
		return fmt.Errorf("failed to store verification token: %w", err)
	}

	verifyURL := fmt.Sprintf("%s/verify-email/%s", m.appURL, token)

	subject := "Verify Your Email"
	body := fmt.Sprintf(verificationEmailHTML, verifyURL, verifyURL)

	return m.SendMail(MailMessage{
		To:      email,
		Subject: subject,
		Text:    fmt.Sprintf("Verify your email: %s", verifyURL),
		HTML:    body,
	})
}

// VerifyEmail validates a verification token and marks the user's email as verified.
func (m *MailerService) VerifyEmail(ctx context.Context, token string) (int64, error) {
	ev, err := m.querier.GetEmailVerification(ctx, hashToken(token))
	if err != nil {
		return 0, fmt.Errorf("invalid or expired verification token")
	}

	if err := m.querier.MarkEmailVerified(ctx, ev.UserID); err != nil {
		return 0, fmt.Errorf("failed to verify email: %w", err)
	}

	m.querier.MarkEmailVerificationUsed(ctx, hashToken(token))
	return ev.UserID, nil
}

// CleanupExpiredVerifications removes expired verification tokens from the database.
func (m *MailerService) CleanupExpiredVerifications(ctx context.Context) {
	if err := m.querier.DeleteExpiredEmailVerifications(ctx); err != nil {
		slog.Error("cleanup: failed to delete expired email verifications", "error", err)
	}
}

// ── Helpers ─────────────────────────────────────────────────────

// generateResetToken generates a secure random token (used for both
// password reset and email verification).
func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken returns the SHA-256 hex digest of a raw token.
// The DB stores only the hash — if the database leaks, the attacker
// cannot reconstruct valid reset/verification URLs from the hashed tokens.
func hashToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}

// ── Email Templates ─────────────────────────────────────────────

const passwordResetEmailHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #f97316, #ea580c); padding: 30px; text-align: center; border-radius: 12px 12px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 24px; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 12px 12px; }
        .button { display: inline-block; background: #f97316; color: white; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-weight: 600; margin: 20px 0; }
        .button:hover { background: #ea580c; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
        .warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hi there,</p>
            <p>We received a request to reset your password. Click the button below to reset it:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #f97316;">%s</p>
            <div class="warning">
                <strong>⚠️ Important:</strong> This link will expire in 1 hour.
            </div>
            <p>If you didn't request this, you can safely ignore this email. Your password will remain unchanged.</p>
            <p>Best regards,<br>The Laju Team</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Laju. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

const verificationEmailHTML = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #6366f1, #4f46e5); padding: 30px; text-align: center; border-radius: 12px 12px 0 0; }
        .header h1 { color: white; margin: 0; font-size: 24px; }
        .content { background: #f9fafb; padding: 30px; border-radius: 0 0 12px 12px; }
        .button { display: inline-block; background: #6366f1; color: white; padding: 12px 30px; text-decoration: none; border-radius: 8px; font-weight: 600; margin: 20px 0; }
        .button:hover { background: #4f46e5; }
        .footer { text-align: center; padding: 20px; color: #6b7280; font-size: 14px; }
        .warning { background: #dbeafe; border-left: 4px solid #3b82f6; padding: 15px; margin: 20px 0; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✉️ Verify Your Email</h1>
        </div>
        <div class="content">
            <p>Welcome to Laju!</p>
            <p>Please verify your email address by clicking the button below:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #6366f1;">%s</p>
            <div class="warning">
                <strong>ℹ️ Note:</strong> This link will expire in 24 hours.
            </div>
            <p>If you didn't create an account, you can safely ignore this email.</p>
            <p>Best regards,<br>The Laju Team</p>
        </div>
        <div class="footer">
            <p>&copy; 2026 Laju. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
