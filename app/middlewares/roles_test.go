package middlewares

import (
	"testing"

	"github.com/maulanashalihin/laju-go/app/models"
)

func TestRoleAllowedSuperAdminBypassesRoleGroups(t *testing.T) {
	management := map[models.UserRole]struct{}{
		models.RoleSPV:   {},
		models.RoleAdmin: {},
	}
	employee := map[models.UserRole]struct{}{
		models.RoleEmployee: {},
	}

	if !roleAllowed(models.RoleSuperAdmin, management) {
		t.Fatal("super admin should pass management role guard")
	}
	if !roleAllowed(models.RoleSuperAdmin, employee) {
		t.Fatal("super admin should pass employee role guard")
	}
	if roleAllowed(models.RoleEmployee, management) {
		t.Fatal("employee must not pass management role guard")
	}
	if !roleAllowed(models.RoleSPV, management) {
		t.Fatal("spv should pass management role guard")
	}
}
