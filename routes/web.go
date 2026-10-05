package routes

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/handlers"
	"github.com/maulanashalihin/laju-go/app/middlewares"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
)

type Handlers struct {
	Public             *handlers.PublicHandler
	Auth               *handlers.AuthHandler
	App                *handlers.AppHandler
	EmployeeAttendance *handlers.EmployeeAttendanceHandler
	Upload             *handlers.UploadHandler
	PasswordReset      *handlers.PasswordResetHandler
}

func SetupRoutes(app *fiber.App, handlers Handlers, store *session.Store, userService *services.UserService, mailerService *services.MailerService, csrfMiddleware *middlewares.CSRFMiddleware) {
	setupStaticRoutes(app)
	setupPublicRoutes(app, handlers.Public)
	setupAuthRoutes(app, handlers.Auth, handlers.PasswordReset, store, mailerService, csrfMiddleware)
	setupAppRoutes(app, handlers.App, handlers.EmployeeAttendance, handlers.Upload, store, userService, csrfMiddleware)
}

func setupStaticRoutes(app *fiber.App) {
	app.Static("/dist", "./dist", fiber.Static{CacheDuration: 365 * 24 * time.Hour, MaxAge: 31536000, Compress: true})
	app.Static("/assets", "./dist/assets", fiber.Static{CacheDuration: 365 * 24 * time.Hour, MaxAge: 31536000, Compress: true})
	app.Static("/public", "./public", fiber.Static{CacheDuration: 1 * time.Hour, MaxAge: 3600})
	app.Static("/storage/avatars", "./storage/avatars", fiber.Static{CacheDuration: 24 * time.Hour, MaxAge: 86400})
	app.Static("/storage/member-photos", "./storage/member-photos", fiber.Static{CacheDuration: 24 * time.Hour, MaxAge: 86400})
	app.Use("/storage/completed", func(c *fiber.Ctx) error {
		c.Set("Content-Disposition", "attachment")
		return c.Next()
	})
	app.Static("/storage/completed", "./storage/completed", fiber.Static{CacheDuration: 24 * time.Hour, MaxAge: 86400})
}

func setupPublicRoutes(app *fiber.App, handler *handlers.PublicHandler) {
	app.Get("/", handler.Index)
	app.Get("/about", handler.About)
}

func setupAuthRoutes(app *fiber.App, authHandler *handlers.AuthHandler, passwordResetHandler *handlers.PasswordResetHandler, store *session.Store, mailerService *services.MailerService, csrfMiddleware *middlewares.CSRFMiddleware) {
	app.Get("/login", middlewares.Guest(store), authHandler.ShowLoginForm)
	app.Post("/login", middlewares.Guest(store), authHandler.Login, middlewares.AuthRateLimit.Limit())

	// Invite-only: no public registration and no public OAuth account creation.
	app.Post("/logout", middlewares.AuthRequired(store), csrfMiddleware.Protect(), authHandler.Logout)

	app.Get("/forgot-password", passwordResetHandler.ShowForgotPasswordForm)
	app.Post("/forgot-password", passwordResetHandler.SendResetLink, middlewares.PasswordResetRateLimit.Limit())
	app.Get("/reset-password/:token", passwordResetHandler.ShowResetPasswordForm)
	app.Post("/reset-password/:token", passwordResetHandler.ResetPassword)
	app.Get("/verify-email/:token", authHandler.VerifyEmail)
}

func setupAppRoutes(app *fiber.App, appHandler *handlers.AppHandler, employeeAttendanceHandler *handlers.EmployeeAttendanceHandler, uploadHandler *handlers.UploadHandler, store *session.Store, userService *services.UserService, csrfMiddleware *middlewares.CSRFMiddleware) {
	protected := app.Group("/app", middlewares.AuthRequired(store), middlewares.NoCache())
	protected.Use(csrfMiddleware.Protect())

	protected.Get("/", func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int64)
		if ok {
			profile, err := userService.GetProfile(userID)
			if err == nil && profile.Role == models.RoleEmployee {
				return c.Redirect("/app/my-attendance")
			}
		}
		return c.Redirect("/app/amaliyah")
	})

	protected.Get("/profile", appHandler.Profile)
	protected.Put("/profile", appHandler.UpdateProfile)
	protected.Put("/profile/password", appHandler.UpdatePassword)
	protected.Post("/upload", uploadHandler.AvatarUpload)

	// Do not mount role middleware on an empty-prefix subgroup. In Fiber that
	// middleware also intercepts sibling /app routes registered afterwards.
	// Attach role guards per route so employee and management routes can coexist
	// under the same /app prefix without cross-blocking each other.
	managementOnly := middlewares.RoleRequired(userService, models.RoleSPV, models.RoleAdmin)
	protected.Get("/members", managementOnly, appHandler.Members)
	protected.Post("/members", managementOnly, appHandler.CreateMember)
	protected.Put("/members/:id", managementOnly, appHandler.UpdateMember)
	protected.Post("/members/:id/photo", managementOnly, appHandler.UploadMemberPhoto)
	protected.Get("/amaliyah", managementOnly, appHandler.AmaliyahLeaderboard)
	protected.Post("/amaliyah/sync", managementOnly, appHandler.SyncAmaliyahWorkspaces)
	protected.Post("/amaliyah/import", managementOnly, appHandler.ImportAmaliyahWorkspaces)
	protected.Get("/attendance", managementOnly, appHandler.AttendanceFoundation)
	protected.Put("/attendance/profiles/:memberID", managementOnly, appHandler.UpdateAttendanceProfile)
	protected.Post("/attendance/invite/:memberID", managementOnly, employeeAttendanceHandler.InviteEmployee)
	protected.Get("/attendance/reviews", managementOnly, employeeAttendanceHandler.AttendanceReviews)
	protected.Post("/attendance/reviews/adjustments/:requestID", managementOnly, employeeAttendanceHandler.ReviewAdjustment)
	protected.Post("/attendance/reviews/overtime/:requestID", managementOnly, employeeAttendanceHandler.ReviewOvertime)
	protected.Get("/leave", managementOnly, employeeAttendanceHandler.LeaveAdministration)
	protected.Post("/leave/entitlements/:memberID", managementOnly, employeeAttendanceHandler.SetLeaveEntitlement)
	protected.Post("/leave/reviews/:requestID", managementOnly, employeeAttendanceHandler.ReviewLeave)
	protected.Get("/fingerprint", managementOnly, appHandler.FingerprintImport)
	protected.Post("/fingerprint/preview", managementOnly, appHandler.PreviewFingerprint)
	protected.Post("/fingerprint/import", managementOnly, appHandler.SaveFingerprintImport)
	protected.Get("/upload", managementOnly, appHandler.UploadTest)

	employeeOnly := middlewares.RoleRequired(userService, models.RoleEmployee)
	protected.Get("/my-attendance", employeeOnly, employeeAttendanceHandler.MyAttendance)
	protected.Post("/my-attendance/start", employeeOnly, employeeAttendanceHandler.StartWork)
	protected.Post("/my-attendance/stop", employeeOnly, employeeAttendanceHandler.StopWork)
	protected.Post("/my-attendance/forgot-checkout", employeeOnly, employeeAttendanceHandler.RequestForgotCheckout)
	protected.Post("/my-attendance/adjustments", employeeOnly, employeeAttendanceHandler.RequestAdjustment)
	protected.Post("/my-attendance/overtime", employeeOnly, employeeAttendanceHandler.RequestOvertime)
	protected.Get("/my-leave", employeeOnly, employeeAttendanceHandler.MyLeave)
	protected.Post("/my-leave/requests", employeeOnly, employeeAttendanceHandler.RequestLeave)
	protected.Post("/my-leave/requests/:requestID/cancel", employeeOnly, employeeAttendanceHandler.CancelLeave)

	uploadHandler.RegisterTUSRoutes(
		app,
		middlewares.AuthRequired(store),
		middlewares.RoleRequired(userService, models.RoleSPV, models.RoleAdmin),
	)

	admin := app.Group("/admin", middlewares.AdminRequired(store, userService), middlewares.NoCache())
	admin.Get("/", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "Admin dashboard"}) })
}

func SetupCSRFMiddleware(secret string, secure bool) *middlewares.CSRFMiddleware {
	config := middlewares.DefaultCSRFConfig(secret)
	config.Secure = secure
	config.SameSite = "Lax"
	return middlewares.NewCSRFMiddleware(config)
}

func SetupPasswordResetHandler(
	mailerService *services.MailerService,
	userService *services.UserService,
	store *session.Store,
	inertiaService *services.InertiaService,
) *handlers.PasswordResetHandler {
	return handlers.NewPasswordResetHandler(mailerService, userService, store, inertiaService)
}

func GetAppURL(appPort string, appEnv string) string {
	if appEnv == "production" {
		return "https://yourdomain.com"
	}
	return fmt.Sprintf("http://localhost:%s", appPort)
}
