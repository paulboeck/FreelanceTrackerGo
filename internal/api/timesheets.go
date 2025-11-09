package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/paulboeck/FreelanceTrackerGo/internal/models"
	"github.com/paulboeck/FreelanceTrackerGo/internal/validator"
)

// TimesheetHandlers handles timesheet API endpoints
type TimesheetHandlers struct {
	timesheets *models.TimesheetModel
	projects   *models.ProjectModel
}

// NewTimesheetHandlers creates a new TimesheetHandlers instance
func NewTimesheetHandlers(timesheets *models.TimesheetModel, projects *models.ProjectModel) *TimesheetHandlers {
	return &TimesheetHandlers{
		timesheets: timesheets,
		projects:   projects,
	}
}

// TimesheetResponse represents a timesheet in API responses
type TimesheetResponse struct {
	ID          int     `json:"id"`
	ProjectID   int     `json:"projectId"`
	WorkDate    string  `json:"workDate"`
	HoursWorked float64 `json:"hoursWorked"`
	HourlyRate  float64 `json:"hourlyRate"`
	Description string  `json:"description"`
	Created     string  `json:"created"`
	Updated     string  `json:"updated"`
}

// TimesheetRequest represents the request body for creating/updating timesheets
type TimesheetRequest struct {
	WorkDate    string  `json:"workDate"`
	HoursWorked float64 `json:"hoursWorked"`
	HourlyRate  float64 `json:"hourlyRate"`
	Description string  `json:"description"`
}

// GetTimesheet returns a single timesheet by ID
// GET /api/v1/timesheets/:id
func (h *TimesheetHandlers) GetTimesheet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid timesheet ID")
		return
	}

	timesheet, err := h.timesheets.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Timesheet not found")
			return
		}
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, timesheetToResponse(&timesheet), nil)
}

// CreateTimesheet creates a new timesheet for a project
// POST /api/v1/projects/:id/timesheets
func (h *TimesheetHandlers) CreateTimesheet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	projectID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid project ID")
		return
	}

	// Check if project exists
	_, err = h.projects.Get(projectID)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Project not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req TimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.WorkDate), "workDate", "Work date is required")
	v.CheckField(req.HoursWorked > 0, "hoursWorked", "Hours worked must be greater than 0")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse work date
	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		v.AddFieldError("workDate", "Invalid date format (use YYYY-MM-DD)")
		ErrorValidation(w, *v)
		return
	}

	// Insert timesheet
	id, err := h.timesheets.Insert(projectID, workDate, req.HoursWorked, req.HourlyRate, req.Description)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return created timesheet
	createdTimesheet, err := h.timesheets.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusCreated, timesheetToResponse(&createdTimesheet), nil)
}

// UpdateTimesheet updates an existing timesheet
// PUT /api/v1/timesheets/:id
func (h *TimesheetHandlers) UpdateTimesheet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid timesheet ID")
		return
	}

	// Check if timesheet exists
	_, err = h.timesheets.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Timesheet not found")
			return
		}
		ErrorInternal(w)
		return
	}

	var req TimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorBadRequest(w, "Invalid JSON")
		return
	}

	// Validate
	v := &validator.Validator{}
	v.CheckField(validator.NotBlank(req.WorkDate), "workDate", "Work date is required")
	v.CheckField(req.HoursWorked > 0, "hoursWorked", "Hours worked must be greater than 0")
	v.CheckField(req.HourlyRate >= 0, "hourlyRate", "Hourly rate must be non-negative")

	if !v.Valid() {
		ErrorValidation(w, *v)
		return
	}

	// Parse work date
	workDate, err := time.Parse("2006-01-02", req.WorkDate)
	if err != nil {
		v.AddFieldError("workDate", "Invalid date format (use YYYY-MM-DD)")
		ErrorValidation(w, *v)
		return
	}

	// Update timesheet
	err = h.timesheets.Update(id, workDate, req.HoursWorked, req.HourlyRate, req.Description)
	if err != nil {
		ErrorInternal(w)
		return
	}

	// Return updated timesheet
	updatedTimesheet, err := h.timesheets.Get(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, timesheetToResponse(&updatedTimesheet), nil)
}

// DeleteTimesheet soft deletes a timesheet
// DELETE /api/v1/timesheets/:id
func (h *TimesheetHandlers) DeleteTimesheet(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		ErrorBadRequest(w, "Invalid timesheet ID")
		return
	}

	// Check if timesheet exists
	_, err = h.timesheets.Get(id)
	if err != nil {
		if err == models.ErrNoRecord {
			ErrorNotFound(w, "Timesheet not found")
			return
		}
		ErrorInternal(w)
		return
	}

	// Delete timesheet
	err = h.timesheets.Delete(id)
	if err != nil {
		ErrorInternal(w)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Timesheet deleted successfully",
	}, nil)
}

// timesheetToResponse converts a models.Timesheet to a TimesheetResponse
func timesheetToResponse(t *models.Timesheet) TimesheetResponse {
	return TimesheetResponse{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		WorkDate:    t.WorkDate.Format("2006-01-02"),
		HoursWorked: t.HoursWorked,
		HourlyRate:  t.HourlyRate,
		Description: t.Description,
		Created:     t.Created.Format("2006-01-02T15:04:05Z"),
		Updated:     t.Updated.Format("2006-01-02T15:04:05Z"),
	}
}
