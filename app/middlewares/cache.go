package middlewares

import "github.com/gofiber/fiber/v2"

// NoCache sets HTTP headers that prevent CDN/proxy caching of authenticated
// responses. Without this, a reverse proxy or Cloudflare can cache one user's
// Inertia page (HTML + JSON props containing user_id, email, role) and serve
// it to a different user — causing session mixing where user A sees user B's
// data.
//
// Applied to /app/* and /admin/* route groups.
func NoCache() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("Vary", "Cookie")
		return c.Next()
	}
}
