package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/api"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// apiRoutes sets up all API routes and returns an HTTP handler
func (app *application) apiRoutes(apiKeyModel *models.APIKeyModel) http.Handler {
	// Create httprouter for API endpoints
	router := httprouter.New()

	// Create rate limiter (100 requests per minute)
	rateLimiter := api.NewRateLimiter(100, 120)

	// Create CORS config (by default, no origins allowed - can be configured later)
	corsConfig := api.DefaultCORSConfig()

	// Create handler instances
	apiKeyHandlers := api.NewAPIKeyHandlers(apiKeyModel, app.users.(*models.UserModel))
	clientHandlers := api.NewClientHandlers(app.clients.(*models.ClientModel), app.projects.(*models.ProjectModel))
	projectHandlers := api.NewProjectHandlersFull(
		app.projects.(*models.ProjectModel),
		app.timesheets.(*models.TimesheetModel),
		app.invoices.(*models.InvoiceModel),
	)
	timesheetHandlers := api.NewTimesheetHandlers(app.timesheets.(*models.TimesheetModel), app.projects.(*models.ProjectModel))
	invoiceHandlers := api.NewInvoiceHandlers(
		app.invoices.(*models.InvoiceModel),
		app.projects.(*models.ProjectModel),
		app.clients.(*models.ClientModel),
		app.settings.(*models.AppSettingModel),
	)
	settingsHandlers := api.NewSettingsHandlers(app.settings.(*models.AppSettingModel))
	reportsHandlers := api.NewReportsHandlers(app.invoices.(*models.InvoiceModel))

	// ========================================
	// Public API routes (no authentication)
	// ========================================
	router.POST("/api/v1/auth/login", apiKeyHandlers.Login)

	// ========================================
	// Protected API routes (authentication required)
	// ========================================

	// API Keys management (requires apikeys:read or apikeys:write scope)
	router.POST("/api/v1/auth/apikeys", wrapWithMiddleware(
		apiKeyHandlers.CreateAPIKey,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("apikeys:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/auth/apikeys", wrapWithMiddleware(
		apiKeyHandlers.ListAPIKeys,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("apikeys:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.DELETE("/api/v1/auth/apikeys/:id", wrapWithMiddleware(
		apiKeyHandlers.DeleteAPIKey,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("apikeys:write"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Clients
	router.GET("/api/v1/clients", wrapWithMiddleware(
		clientHandlers.ListClients,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/clients/:id", wrapWithMiddleware(
		clientHandlers.GetClient,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.POST("/api/v1/clients", wrapWithMiddleware(
		clientHandlers.CreateClient,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/clients/:id", wrapWithMiddleware(
		clientHandlers.UpdateClient,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.DELETE("/api/v1/clients/:id", wrapWithMiddleware(
		clientHandlers.DeleteClient,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/clients/:id/projects", wrapWithMiddleware(
		clientHandlers.GetClientProjects,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:read", "projects:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/clients/:id/hourlyrate", wrapWithMiddleware(
		clientHandlers.GetClientHourlyRate,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("clients:read"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Projects
	router.GET("/api/v1/projects", wrapWithMiddleware(
		projectHandlers.ListProjects,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/projects/:id", wrapWithMiddleware(
		projectHandlers.GetProject,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.POST("/api/v1/projects", wrapWithMiddleware(
		projectHandlers.CreateProject,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/projects/:id", wrapWithMiddleware(
		projectHandlers.UpdateProject,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PATCH("/api/v1/projects/:id/status", wrapWithMiddleware(
		projectHandlers.UpdateProjectStatus,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.DELETE("/api/v1/projects/:id", wrapWithMiddleware(
		projectHandlers.DeleteProject,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/projects/:id/timesheets", wrapWithMiddleware(
		projectHandlers.GetProjectTimesheets,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:read", "timesheets:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/projects/:id/invoices", wrapWithMiddleware(
		projectHandlers.GetProjectInvoices,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("projects:read", "invoices:read"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Timesheets
	router.GET("/api/v1/timesheets/:id", wrapWithMiddleware(
		timesheetHandlers.GetTimesheet,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("timesheets:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.POST("/api/v1/projects/:id/timesheets", wrapWithMiddleware(
		timesheetHandlers.CreateTimesheet,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("timesheets:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/timesheets/:id", wrapWithMiddleware(
		timesheetHandlers.UpdateTimesheet,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("timesheets:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.DELETE("/api/v1/timesheets/:id", wrapWithMiddleware(
		timesheetHandlers.DeleteTimesheet,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("timesheets:write"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Invoices
	router.GET("/api/v1/invoices/:id", wrapWithMiddleware(
		invoiceHandlers.GetInvoice,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.POST("/api/v1/projects/:id/invoices", wrapWithMiddleware(
		invoiceHandlers.CreateInvoice,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/invoices/:id", wrapWithMiddleware(
		invoiceHandlers.UpdateInvoice,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.DELETE("/api/v1/invoices/:id", wrapWithMiddleware(
		invoiceHandlers.DeleteInvoice,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/invoices/:id/pdf", wrapWithMiddleware(
		invoiceHandlers.GenerateInvoicePDF,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.POST("/api/v1/invoices/:id/email", wrapWithMiddleware(
		invoiceHandlers.EmailInvoice,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("invoices:write"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Reports
	router.GET("/api/v1/reports/income", wrapWithMiddleware(
		reportsHandlers.GetIncomeReport,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("reports:read"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// Settings
	router.GET("/api/v1/settings", wrapWithMiddleware(
		settingsHandlers.GetAllSettings,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("settings:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.GET("/api/v1/settings/:key", wrapWithMiddleware(
		settingsHandlers.GetSetting,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("settings:read"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/settings", wrapWithMiddleware(
		settingsHandlers.UpdateAllSettings,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("settings:write"),
		api.RateLimitMiddleware(rateLimiter),
	))
	router.PUT("/api/v1/settings/:key", wrapWithMiddleware(
		settingsHandlers.UpdateSetting,
		api.AuthMiddleware(apiKeyModel),
		api.RequireScopes("settings:write"),
		api.RateLimitMiddleware(rateLimiter),
	))

	// ========================================
	// Utility endpoints (no authentication)
	// ========================================

	// Health check endpoint for monitoring and deployment
	router.GET("/health", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`))
	})

	// API documentation endpoint
	router.GET("/api/docs", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>FreelanceTracker API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            window.ui = SwaggerUIBundle({
                url: "https://raw.githubusercontent.com/yourusername/freelancetracker/main/docs/openapi.yaml",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`))
	})

	// Wrap router with CORS middleware
	return api.CORSMiddleware(corsConfig)(router)
}

// wrapWithMiddleware wraps an httprouter.Handle with multiple middleware functions
func wrapWithMiddleware(handler httprouter.Handle, middlewares ...func(http.Handler) http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Convert httprouter.Handle to http.Handler
		var h http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler(w, r, ps)
		})

		// Apply middlewares in reverse order (so they execute in the order provided)
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}

		h.ServeHTTP(w, r)
	}
}
