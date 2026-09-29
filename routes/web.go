package routes

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/handlers"
	"github.com/maulanashalihin/laju-go/app/middlewares"
	"github.com/maulanashalihin/laju-go/app/services"
	"github.com/maulanashalihin/laju-go/app/session"
)

type Handlers struct {
	Public        *handlers.PublicHandler
	Auth          *handlers.AuthHandler
	App           *handlers.AppHandler
	Upload        *handlers.UploadHandler
	PasswordReset *handlers.PasswordResetHandler
}

func SetupRoutes(app *fiber.App, handlers Handlers, store *session.Store, userService *services.UserService, mailerService *services.MailerService, csrfMiddleware *middlewares.CSRFMiddleware) {
	setupStaticRoutes(app)
	setupPublicRoutes(app, handlers.Public)
	setupAuthRoutes(app, handlers.Auth, handlers.PasswordReset, store, mailerService, csrfMiddleware)
	setupAppRoutes(app, handlers.App, handlers.Upload, store, userService, csrfMiddleware)
}

func setupStaticRoutes(app *fiber.App) {
	app.Static("/dist", "./dist", fiber.Static{
		CacheDuration: 365 * 24 * time.Hour,
		MaxAge:        31536000,
		Compress:      true,
	})
	app.Static("/assets", "./dist/assets", fiber.Static{
		CacheDuration: 365 * 24 * time.Hour,
		MaxAge:        31536000,
		Compress:      true,
	})
	app.Static("/public", "./public", fiber.Static{
		CacheDuration: 1 * time.Hour,
		MaxAge:        3600,
	})
	app.Static("/storage/avatars", "./storage/avatars", fiber.Static{
		CacheDuration: 24 * time.Hour,
		MaxAge:        86400,
	})
	app.Use("/storage/completed", func(c *fiber.Ctx) error {
		c.Set("Content-Disposition", "attachment")
		return c.Next()
	})
	app.Static("/storage/completed", "./storage/completed", fiber.Static{
		CacheDuration: 24 * time.Hour,
		MaxAge:        86400,
	})
}

func setupPublicRoutes(app *fiber.App, handler *handlers.PublicHandler) {
	app.Get("/", handler.Index)
	app.Get("/about", handler.About)
}

func setupAuthRoutes(app *fiber.App, authHandler *handlers.AuthHandler, passwordResetHandler *handlers.PasswordResetHandler, store *session.Store, mailerService *services.MailerService, csrfMiddleware *middlewares.CSRFMiddleware) {
	app.Get("/login", middlewares.Guest(store), authHandler.ShowLoginForm)
	app.Post("/login", middlewares.Guest(store), authHandler.Login, middlewares.AuthRateLimit.Limit())

	app.Get("/register", middlewares.Guest(store), authHandler.ShowRegisterForm)
	app.Post("/register", middlewares.Guest(store), authHandler.Register, middlewares.AuthRateLimit.Limit())

	app.Get("/auth/google", authHandler.GoogleLogin)
	app.Get("/auth/google/callback", authHandler.GoogleCallback)

	app.Post("/logout", middlewares.AuthRequired(store), csrfMiddleware.Protect(), authHandler.Logout)

	app.Get("/forgot-password", passwordResetHandler.ShowForgotPasswordForm)
	app.Post("/forgot-password", passwordResetHandler.SendResetLink, middlewares.PasswordResetRateLimit.Limit())
	app.Get("/reset-password/:token", passwordResetHandler.ShowResetPasswordForm)
	app.Post("/reset-password/:token", passwordResetHandler.ResetPassword)

	app.Get("/verify-email/:token", authHandler.VerifyEmail)
}

func setupAppRoutes(app *fiber.App, appHandler *handlers.AppHandler, uploadHandler *handlers.UploadHandler, store *session.Store, userService *services.UserService, csrfMiddleware *middlewares.CSRFMiddleware) {
	protected := app.Group("/app", middlewares.AuthRequired(store), middlewares.NoCache())
	protected.Use(csrfMiddleware.Protect())

	protected.Get("/", appHandler.Dashboard)

	protected.Get("/members", appHandler.Members)
	protected.Post("/members", appHandler.CreateMember)
	protected.Put("/members/:id", appHandler.UpdateMember)

	protected.Get("/attendance", appHandler.AttendanceFoundation)
	protected.Put("/attendance/profiles/:memberID", appHandler.UpdateAttendanceProfile)

	protected.Get("/fingerprint", appHandler.FingerprintImport)
	protected.Post("/fingerprint/preview", appHandler.PreviewFingerprint)
	protected.Post("/fingerprint/import", appHandler.SaveFingerprintImport)

	protected.Get("/profile", appHandler.Profile)
	protected.Put("/profile", appHandler.UpdateProfile)
	protected.Put("/profile/password", appHandler.UpdatePassword)

	protected.Get("/upload", appHandler.UploadTest)
	protected.Post("/upload", uploadHandler.AvatarUpload)

	authMiddleware := middlewares.AuthRequired(store)
	uploadHandler.RegisterTUSRoutes(app, authMiddleware)

	admin := app.Group("/admin", middlewares.AdminRequired(store, userService), middlewares.NoCache())
	admin.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Admin dashboard",
		})
	})
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
	return handlers.NewPasswordResetHandler(
		mailerService,
		userService,
		store,
		inertiaService,
	)
}

func GetAppURL(appPort string, appEnv string) string {
	if appEnv == "production" {
		return "https://yourdomain.com"
	}
	return fmt.Sprintf("http://localhost:%s", appPort)
}
