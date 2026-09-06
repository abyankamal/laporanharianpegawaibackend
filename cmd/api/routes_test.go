package main

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/handler"
)

func TestRoutes_SetupAndRESTfulAliases(t *testing.T) {
	app := fiber.New()

	// Initialize dummy empty handlers
	h := Handlers{
		Auth:      &handler.AuthHandler{},
		User:      &handler.UserHandler{},
		Notif:     &handler.NotificationHandler{},
		WorkHour:  &handler.WorkHourHandler{},
		Holiday:   &handler.HolidayHandler{},
		Report:    &handler.ReportHandler{},
		Review:    &handler.ReviewHandler{},
		Task:      &handler.TaskHandler{},
		Dashboard: &handler.DashboardHandler{},
		Jabatan:   &handler.JabatanHandler{},
		Admin:     &handler.AdminHandler{},
		Absensi:   &handler.AbsensiHandler{},
		Izin:      &handler.IzinHandler{},
	}

	setupRoutes(app, h)

	// List of critical routes that must be registered (not 404 MethodNotAllowed / RouteNotFound)
	// Because Protected() middleware is attached, unauthorized calls return 401 Unauthorized, NOT 404.
	testCases := []struct {
		method string
		path   string
	}{
		// Mobile routes
		{"PUT", "/api/mobile/profile/password"},
		{"PATCH", "/api/mobile/profile/password"},
		{"PATCH", "/api/mobile/notifications/1/read"},
		{"PATCH", "/api/mobile/notifications/1"},
		{"PUT", "/api/mobile/izin/1/approve"},
		{"PATCH", "/api/mobile/izin/1/approve"},
		{"PATCH", "/api/mobile/izin/1"},
		{"PATCH", "/api/mobile/tasks/1"},

		// Web routes
		{"PATCH", "/api/web/users/1/password"},
		{"PATCH", "/api/web/users/1"},
		{"PATCH", "/api/web/izin/1/approve"},
		{"PATCH", "/api/web/izin/1"},

		// Top-level RESTful aliases
		{"PUT", "/api/profile/password"},
		{"PATCH", "/api/profile/password"},
		{"GET", "/api/notifications"},
		{"GET", "/api/notifications/1"},
		{"PATCH", "/api/notifications/1/read"},
		{"PATCH", "/api/notifications/1"},
		{"PUT", "/api/izin/1/approve"},
		{"PATCH", "/api/izin/1/approve"},
		{"PATCH", "/api/izin/1"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			// Protected middleware intercepts before handler, returning 401 Unauthorized
			// Crucially: It is NOT 404 Not Found, proving the route is registered!
			assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
		})
	}
}
