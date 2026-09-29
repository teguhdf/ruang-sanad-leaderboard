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
		return appHandler.Dashboard(c)
	})

	protected.Get("/profile", appHandler.Profile)
	protected.Put("/profile", appHandler.UpdateProfile)
	protected.Put("/profile/password", appHandler.UpdatePassword)
	protected.Post("/upload", uploadHandler.AvatarUpload)

	management := protected.Group("", middlewares.RoleRequired(userService, models.RoleSPV, models.RoleAdmin))
	management.Get("/members", appHandler.Members)
	management.Post("/members", appHandler.CreateMember)
	management.Put("/members/:id", appHandler.UpdateMember)
	management.Get("/attendance", appHandler.AttendanceFoundation)
	management.Put("/attendance/profiles/:memberID", appHandler.UpdateAttendanceProfile)
	management.Post("/attendance/invite/:memberID", employeeAttendanceHandler.InviteEmployee)
	management.Get("/attendance/reviews", employeeAttendanceHandler.AttendanceReviews)
	management.Post("/attendance/reviews/adjustments/:requestID", employeeAttendanceHandler.ReviewAdjustment)
	management.Post("/attendance/reviews/overtime/:requestID", employeeAttendanceHandler.ReviewOvertime)
	management.Get("/fingerprint", appHandler.FingerprintImport)
	management.Post("/fingerprint/preview", appHandler.PreviewFingerprint)
	management.Post("/fingerprint/import", appHandler.SaveFingerprintImport)
	management.Get("/upload", appHandler.UploadTest)

	employee := protected.Group("", middlewares.RoleRequired(userService, models.RoleEmployee))
	employee.Get("/my-attendance", employeeAttendanceHandler.MyAttendance)
	employee.Post("/my-attendance/start", employeeAttendanceHandler.StartWork)
	employee.Post("/my-attendance/stop", employeeAttendanceHandler.StopWork)
	employee.Post("/my-attendance/adjustments", employeeAttendanceHandler.RequestAdjustment)
	employee.Post("/my-attendance/overtime", employeeAttendanceHandler.RequestOvertime)

	authMiddleware := middlewares.AuthRequired(store)
	uploadHandler.RegisterTUSRoutes(app, authMiddleware)

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
