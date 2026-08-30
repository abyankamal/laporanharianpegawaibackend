package domain

import (
	"testing"
)

func TestIsValidRole(t *testing.T) {
	validRoles := []string{"admin", "lurah", "sekertaris", "sekretaris", "kasi", "staf", "LURAH", " Admin "}
	for _, role := range validRoles {
		if !IsValidRole(role) {
			t.Errorf("expected %s to be a valid role", role)
		}
	}

	invalidRoles := []string{"", "guest", "superadmin", "unknown"}
	for _, role := range invalidRoles {
		if IsValidRole(role) {
			t.Errorf("expected %s to be an invalid role", role)
		}
	}
}

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sekretaris", RoleSekertaris},
		{"Sekretaris", RoleSekertaris},
		{"SEKERTARIS", RoleSekertaris},
		{"Lurah", RoleLurah},
		{" ADMIN ", RoleAdmin},
		{"kasi", RoleKasi},
		{"staf", RoleStaf},
	}

	for _, tt := range tests {
		got := NormalizeRole(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeRole(%q) = %q; expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestRoleHelpers(t *testing.T) {
	if !IsAdmin("admin") || !IsAdmin("ADMIN") {
		t.Error("IsAdmin failed")
	}
	if !IsLurah("lurah") || !IsLurah("Lurah") {
		t.Error("IsLurah failed")
	}
	if !IsSekertaris("sekertaris") || !IsSekertaris("sekretaris") {
		t.Error("IsSekertaris failed")
	}
	if !IsKasi("kasi") || !IsKasi("KASI") {
		t.Error("IsKasi failed")
	}
	if !IsStaf("staf") || !IsStaf("Staf") {
		t.Error("IsStaf failed")
	}
}
