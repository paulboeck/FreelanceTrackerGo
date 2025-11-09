package api

import (
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
)

// ProjectHandlers handles project API endpoints
type ProjectHandlers struct {
	projects *models.ProjectModel
}

// NewProjectHandlers creates a new ProjectHandlers instance
func NewProjectHandlers(projects *models.ProjectModel) *ProjectHandlers {
	return &ProjectHandlers{
		projects: projects,
	}
}

// ProjectResponse represents a project in API responses
type ProjectResponse struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	ClientID       int      `json:"clientId"`
	Status         string   `json:"status"`
	HourlyRate     float64  `json:"hourlyRate"`
	Deadline       *string  `json:"deadline"`
	ScheduledStart *string  `json:"scheduledStart"`
	Notes          string   `json:"notes"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
}

// projectToResponse converts a models.Project to a ProjectResponse
func projectToResponse(p *models.Project) ProjectResponse {
	var deadline, scheduledStart *string
	if p.Deadline != nil {
		d := p.Deadline.Format("2006-01-02T15:04:05Z")
		deadline = &d
	}
	if p.ScheduledStart != nil {
		s := p.ScheduledStart.Format("2006-01-02T15:04:05Z")
		scheduledStart = &s
	}

	return ProjectResponse{
		ID:             p.ID,
		Name:           p.Name,
		ClientID:       p.ClientID,
		Status:         p.Status,
		HourlyRate:     p.HourlyRate,
		Deadline:       deadline,
		ScheduledStart: scheduledStart,
		Notes:          p.Notes,
		Created:        p.Created.Format("2006-01-02T15:04:05Z"),
		Updated:        p.Updated.Format("2006-01-02T15:04:05Z"),
	}
}

// TODO: Implement full CRUD operations for projects
// - ListProjects
// - GetProject
// - CreateProject
// - UpdateProject
// - UpdateProjectStatus
// - DeleteProject
// - GetProjectTimesheets
// - GetProjectInvoices
