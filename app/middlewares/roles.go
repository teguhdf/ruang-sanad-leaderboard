package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maulanashalihin/laju-go/app/models"
	"github.com/maulanashalihin/laju-go/app/services"
)

func RoleRequired(userService *services.UserService, allowed ...models.UserRole) fiber.Handler {
	allowedSet := make(map[models.UserRole]struct{}, len(allowed))
	for _, role := range allowed {
		allowedSet[role] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(int64)
		if !ok || userID <= 0 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Not authenticated"})
		}

		user, err := userService.GetProfile(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify role"})
		}
		if _, ok := allowedSet[user.Role]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
		}
		c.Locals("role", string(user.Role))
		return c.Next()
	}
}
