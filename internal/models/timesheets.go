package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Timesheet represents a timesheet in the system
type Timesheet struct {
	ID          int
	ProjectID   int
	WorkDate    time.Time
	HoursWorked float64
	Updated     time.Time
	Created     time.Time
	DeletedAt   *time.Time
}

// TimesheetModel wraps the generated SQLC Queries for timesheet operations
type TimesheetModel struct {
	queries *db.Queries
}

// NewTimesheetModel creates a new TimesheetModel
func NewTimesheetModel(database *sql.DB) *TimesheetModel {
	return &TimesheetModel{
		queries: db.New(database),
	}
}

// Insert adds a new timesheet to the database and returns its ID
func (t *TimesheetModel) Insert(projectID int, workDate time.Time, hoursWorked float64) (int, error) {
	ctx := context.Background()
	params := db.InsertTimesheetParams{
		ProjectID:   int64(projectID),
		WorkDate:    workDate,
		HoursWorked: hoursWorked,
	}
	id, err := t.queries.InsertTimesheet(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get retrieves a timesheet by ID
func (t *TimesheetModel) Get(id int) (Timesheet, error) {
	ctx := context.Background()
	row, err := t.queries.GetTimesheet(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Timesheet{}, ErrNoRecord
		}
		return Timesheet{}, err
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	timesheet := Timesheet{
		ID:          int(row.ID),
		ProjectID:   int(row.ProjectID),
		WorkDate:    row.WorkDate,
		HoursWorked: row.HoursWorked,
		Updated:     row.UpdatedAt,
		Created:     row.CreatedAt,
		DeletedAt:   deletedAt,
	}

	return timesheet, nil
}

// GetByProject retrieves all timesheets for a specific project
func (t *TimesheetModel) GetByProject(projectID int) ([]Timesheet, error) {
	ctx := context.Background()
	rows, err := t.queries.GetTimesheetsByProject(ctx, int64(projectID))
	if err != nil {
		return nil, err
	}

	timesheets := make([]Timesheet, len(rows))
	for i, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if dt, ok := row.DeletedAt.(time.Time); ok {
				deletedAt = &dt
			}
		}

		timesheets[i] = Timesheet{
			ID:          int(row.ID),
			ProjectID:   int(row.ProjectID),
			WorkDate:    row.WorkDate,
			HoursWorked: row.HoursWorked,
			Updated:     row.UpdatedAt,
			Created:     row.CreatedAt,
			DeletedAt:   deletedAt,
		}
	}

	return timesheets, nil
}

// Update modifies an existing timesheet in the database
func (t *TimesheetModel) Update(id int, workDate time.Time, hoursWorked float64) error {
	ctx := context.Background()
	params := db.UpdateTimesheetParams{
		ID:          int64(id),
		WorkDate:    workDate,
		HoursWorked: hoursWorked,
	}
	return t.queries.UpdateTimesheet(ctx, params)
}

// Delete soft deletes a timesheet by setting the deleted_at timestamp
func (t *TimesheetModel) Delete(id int) error {
	ctx := context.Background()
	return t.queries.DeleteTimesheet(ctx, int64(id))
}

// TimesheetModelInterface defines the interface for timesheet operations
type TimesheetModelInterface interface {
	Insert(projectID int, workDate time.Time, hoursWorked float64) (int, error)
	Get(id int) (Timesheet, error)
	GetByProject(projectID int) ([]Timesheet, error)
	Update(id int, workDate time.Time, hoursWorked float64) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ TimesheetModelInterface = (*TimesheetModel)(nil)