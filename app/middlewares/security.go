package middlewares

import "github.com/gofiber/fiber/v2"

// SecurityHeaders sets standard HTTP security headers on every response.
// These protect against MIME sniffing, clickjacking, protocol downgrade,
// and referrer leakage when the app runs without a reverse proxy that
// would otherwise set them.
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("X-XSS-Protection", "0") // modern browsers use CSP instead; old IE flag can introduce vulnerabilities
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		return c.Next()
	}
}
