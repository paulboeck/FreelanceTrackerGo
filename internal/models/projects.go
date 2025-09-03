package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Project represents a project in the system
type Project struct {
	ID                     int
	Name                   string
	ClientID               int
	Status                 string
	HourlyRate             float64
	Deadline               *time.Time
	ScheduledStart         *time.Time
	InvoiceCCEmail         string
	InvoiceCCDescription   string
	ScheduleComments       string
	AdditionalInfo         string
	AdditionalInfo2        string
	DiscountPercent        *float64
	DiscountReason         string
	AdjustmentAmount       *float64
	AdjustmentReason       string
	CurrencyDisplay        string
	CurrencyConversionRate float64
	FlatFeeInvoice         bool
	Notes                  string
	Updated                time.Time
	Created                time.Time
	DeletedAt              *time.Time
}

// ProjectWithClient represents a project with client information for list views
type ProjectWithClient struct {
	ID                     int
	Name                   string
	ClientID               int
	ClientName             string
	Status                 string
	HourlyRate             float64
	Deadline               *time.Time
	ScheduledStart         *time.Time
	InvoiceCCEmail         string
	InvoiceCCDescription   string
	ScheduleComments       string
	AdditionalInfo         string
	AdditionalInfo2        string
	DiscountPercent        *float64
	DiscountReason         string
	AdjustmentAmount       *float64
	AdjustmentReason       string
	CurrencyDisplay        string
	CurrencyConversionRate float64
	FlatFeeInvoice         bool
	Notes                  string
	Updated                time.Time
	Created                time.Time
	DeletedAt              *time.Time
}

// ProjectModel wraps the generated SQLC Queries for project operations
type ProjectModel struct {
	queries *db.Queries
}

// NewProjectModel creates a new ProjectModel
func NewProjectModel(database *sql.DB) *ProjectModel {
	return &ProjectModel{
		queries: db.New(database),
	}
}

// Insert adds a new project to the database and returns its ID
func (p *ProjectModel) Insert(project Project) (int, error) {
	ctx := context.Background()
	
	// Helper function to convert *time.Time to sql.NullString for dates
	timeToNullString := func(t *time.Time) sql.NullString {
		if t == nil {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: t.Format("2006-01-02"), Valid: true}
	}
	
	// Helper function to convert string to sql.NullString
	stringToNullString := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}
	
	// Helper function to convert *float64 to sql.NullFloat64
	floatToNullFloat64 := func(f *float64) sql.NullFloat64 {
		if f == nil {
			return sql.NullFloat64{Valid: false}
		}
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	
	params := db.InsertProjectParams{
		Name:                   project.Name,
		ClientID:               int64(project.ClientID),
		Status:                 project.Status,
		HourlyRate:             project.HourlyRate,
		Deadline:               timeToNullString(project.Deadline),
		ScheduledStart:         timeToNullString(project.ScheduledStart),
		InvoiceCcEmail:         stringToNullString(project.InvoiceCCEmail),
		InvoiceCcDescription:   stringToNullString(project.InvoiceCCDescription),
		ScheduleComments:       stringToNullString(project.ScheduleComments),
		AdditionalInfo:         stringToNullString(project.AdditionalInfo),
		AdditionalInfo2:        stringToNullString(project.AdditionalInfo2),
		DiscountPercent:        floatToNullFloat64(project.DiscountPercent),
		DiscountReason:         stringToNullString(project.DiscountReason),
		AdjustmentAmount:       floatToNullFloat64(project.AdjustmentAmount),
		AdjustmentReason:       stringToNullString(project.AdjustmentReason),
		CurrencyDisplay:        project.CurrencyDisplay,
		CurrencyConversionRate: project.CurrencyConversionRate,
		FlatFeeInvoice:         0, // Convert bool to int64 (0 = false, 1 = true)
		Notes:                  stringToNullString(project.Notes),
	}
	
	// Convert bool to int64 for SQLite
	if project.FlatFeeInvoice {
		params.FlatFeeInvoice = 1
	}
	
	id, err := p.queries.InsertProject(ctx, params)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// Get retrieves a project by ID
func (p *ProjectModel) Get(id int) (Project, error) {
	ctx := context.Background()
	row, err := p.queries.GetProject(ctx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Project{}, ErrNoRecord
		}
		return Project{}, err
	}

	// Helper function to convert sql.NullString to *time.Time for dates
	nullStringToTime := func(ns sql.NullString) *time.Time {
		if !ns.Valid || ns.String == "" {
			return nil
		}
		if t, err := time.Parse("2006-01-02", ns.String); err == nil {
			return &t
		}
		return nil
	}

	// Helper function to convert sql.NullFloat64 to *float64
	nullFloat64ToFloat := func(nf sql.NullFloat64) *float64 {
		if !nf.Valid {
			return nil
		}
		return &nf.Float64
	}

	var deletedAt *time.Time
	if row.DeletedAt != nil {
		if dt, ok := row.DeletedAt.(time.Time); ok {
			deletedAt = &dt
		}
	}

	project := Project{
		ID:                     int(row.ID),
		Name:                   row.Name,
		ClientID:               int(row.ClientID),
		Status:                 row.Status,
		HourlyRate:             row.HourlyRate,
		Deadline:               nullStringToTime(row.Deadline),
		ScheduledStart:         nullStringToTime(row.ScheduledStart),
		InvoiceCCEmail:         row.InvoiceCcEmail.String,
		InvoiceCCDescription:   row.InvoiceCcDescription.String,
		ScheduleComments:       row.ScheduleComments.String,
		AdditionalInfo:         row.AdditionalInfo.String,
		AdditionalInfo2:        row.AdditionalInfo2.String,
		DiscountPercent:        nullFloat64ToFloat(row.DiscountPercent),
		DiscountReason:         row.DiscountReason.String,
		AdjustmentAmount:       nullFloat64ToFloat(row.AdjustmentAmount),
		AdjustmentReason:       row.AdjustmentReason.String,
		CurrencyDisplay:        row.CurrencyDisplay,
		CurrencyConversionRate: row.CurrencyConversionRate,
		FlatFeeInvoice:         row.FlatFeeInvoice != 0,
		Notes:                  row.Notes.String,
		Updated:                row.UpdatedAt,
		Created:                row.CreatedAt,
		DeletedAt:              deletedAt,
	}

	return project, nil
}

// GetByClient retrieves all projects for a specific client
func (p *ProjectModel) GetByClient(clientID int) ([]Project, error) {
	ctx := context.Background()
	rows, err := p.queries.GetProjectsByClient(ctx, int64(clientID))
	if err != nil {
		return nil, err
	}

	// Helper functions (reused from Get method)
	nullStringToTime := func(ns sql.NullString) *time.Time {
		if !ns.Valid || ns.String == "" {
			return nil
		}
		if t, err := time.Parse("2006-01-02", ns.String); err == nil {
			return &t
		}
		return nil
	}

	nullFloat64ToFloat := func(nf sql.NullFloat64) *float64 {
		if !nf.Valid {
			return nil
		}
		return &nf.Float64
	}

	projects := make([]Project, len(rows))
	for i, row := range rows {
		var deletedAt *time.Time
		if row.DeletedAt != nil {
			if dt, ok := row.DeletedAt.(time.Time); ok {
				deletedAt = &dt
			}
		}

		projects[i] = Project{
			ID:                     int(row.ID),
			Name:                   row.Name,
			ClientID:               int(row.ClientID),
			Status:                 row.Status,
			HourlyRate:             row.HourlyRate,
			Deadline:               nullStringToTime(row.Deadline),
			ScheduledStart:         nullStringToTime(row.ScheduledStart),
			InvoiceCCEmail:         row.InvoiceCcEmail.String,
			InvoiceCCDescription:   row.InvoiceCcDescription.String,
			ScheduleComments:       row.ScheduleComments.String,
			AdditionalInfo:         row.AdditionalInfo.String,
			AdditionalInfo2:        row.AdditionalInfo2.String,
			DiscountPercent:        nullFloat64ToFloat(row.DiscountPercent),
			DiscountReason:         row.DiscountReason.String,
			AdjustmentAmount:       nullFloat64ToFloat(row.AdjustmentAmount),
			AdjustmentReason:       row.AdjustmentReason.String,
			CurrencyDisplay:        row.CurrencyDisplay,
			CurrencyConversionRate: row.CurrencyConversionRate,
			FlatFeeInvoice:         row.FlatFeeInvoice != 0,
			Notes:                  row.Notes.String,
			Updated:                row.UpdatedAt,
			Created:                row.CreatedAt,
			DeletedAt:              deletedAt,
		}
	}

	return projects, nil
}

// Update modifies an existing project in the database
func (p *ProjectModel) Update(project Project) error {
	ctx := context.Background()
	
	// Helper functions (reused from Insert method)
	timeToNullString := func(t *time.Time) sql.NullString {
		if t == nil {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: t.Format("2006-01-02"), Valid: true}
	}
	
	stringToNullString := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}
	
	floatToNullFloat64 := func(f *float64) sql.NullFloat64 {
		if f == nil {
			return sql.NullFloat64{Valid: false}
		}
		return sql.NullFloat64{Float64: *f, Valid: true}
	}
	
	params := db.UpdateProjectParams{
		Name:                   project.Name,
		Status:                 project.Status,
		HourlyRate:             project.HourlyRate,
		Deadline:               timeToNullString(project.Deadline),
		ScheduledStart:         timeToNullString(project.ScheduledStart),
		InvoiceCcEmail:         stringToNullString(project.InvoiceCCEmail),
		InvoiceCcDescription:   stringToNullString(project.InvoiceCCDescription),
		ScheduleComments:       stringToNullString(project.ScheduleComments),
		AdditionalInfo:         stringToNullString(project.AdditionalInfo),
		AdditionalInfo2:        stringToNullString(project.AdditionalInfo2),
		DiscountPercent:        floatToNullFloat64(project.DiscountPercent),
		DiscountReason:         stringToNullString(project.DiscountReason),
		AdjustmentAmount:       floatToNullFloat64(project.AdjustmentAmount),
		AdjustmentReason:       stringToNullString(project.AdjustmentReason),
		CurrencyDisplay:        project.CurrencyDisplay,
		CurrencyConversionRate: project.CurrencyConversionRate,
		FlatFeeInvoice:         0,
		Notes:                  stringToNullString(project.Notes),
		ID:                     int64(project.ID),
	}
	
	// Convert bool to int64 for SQLite
	if project.FlatFeeInvoice {
		params.FlatFeeInvoice = 1
	}
	
	return p.queries.UpdateProject(ctx, params)
}

// Delete soft deletes a project by setting the deleted_at timestamp
func (p *ProjectModel) Delete(id int) error {
	ctx := context.Background()
	return p.queries.DeleteProject(ctx, int64(id))
}

// GetWithPagination retrieves projects with client information using pagination
func (p *ProjectModel) GetWithPagination(limit, offset int64) ([]ProjectWithClient, error) {
	ctx := context.Background()
	rows, err := p.queries.GetProjectsWithClientPagination(ctx, db.GetProjectsWithClientPaginationParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	projects := make([]ProjectWithClient, len(rows))
	for i, row := range rows {
		project, err := p.convertPaginationRowToProjectWithClient(row)
		if err != nil {
			return nil, err
		}
		projects[i] = project
	}

	return projects, nil
}

// convertPaginationRowToProjectWithClient converts a pagination database row to a ProjectWithClient struct
func (p *ProjectModel) convertPaginationRowToProjectWithClient(row db.GetProjectsWithClientPaginationRow) (ProjectWithClient, error) {
	// Helper function to convert sql.NullString to string
	nullStringToString := func(ns sql.NullString) string {
		if ns.Valid {
			return ns.String
		}
		return ""
	}
	
	// Helper function to convert sql.NullFloat64 to *float64
	nullFloat64ToFloat := func(nf sql.NullFloat64) *float64 {
		if nf.Valid {
			return &nf.Float64
		}
		return nil
	}
	
	// Helper function to convert sql.NullString date to *time.Time
	nullStringToTime := func(ns sql.NullString) *time.Time {
		if !ns.Valid || ns.String == "" {
			return nil
		}
		if t, err := time.Parse("2006-01-02", ns.String); err == nil {
			return &t
		}
		return nil
	}
	
	return ProjectWithClient{
		ID:                     int(row.ID),
		Name:                   row.Name,
		ClientID:               int(row.ClientID),
		ClientName:             row.ClientName,
		Status:                 row.Status,
		HourlyRate:             row.HourlyRate,
		Deadline:               nullStringToTime(row.Deadline),
		ScheduledStart:         nullStringToTime(row.ScheduledStart),
		InvoiceCCEmail:         nullStringToString(row.InvoiceCcEmail),
		InvoiceCCDescription:   nullStringToString(row.InvoiceCcDescription),
		ScheduleComments:       nullStringToString(row.ScheduleComments),
		AdditionalInfo:         nullStringToString(row.AdditionalInfo),
		AdditionalInfo2:        nullStringToString(row.AdditionalInfo2),
		DiscountPercent:        nullFloat64ToFloat(row.DiscountPercent),
		DiscountReason:         nullStringToString(row.DiscountReason),
		AdjustmentAmount:       nullFloat64ToFloat(row.AdjustmentAmount),
		AdjustmentReason:       nullStringToString(row.AdjustmentReason),
		CurrencyDisplay:        row.CurrencyDisplay,
		CurrencyConversionRate: row.CurrencyConversionRate,
		FlatFeeInvoice:         row.FlatFeeInvoice == 1,
		Notes:                  nullStringToString(row.Notes),
		Updated:                row.UpdatedAt,
		Created:                row.CreatedAt,
	}, nil
}

// GetCount returns the total count of non-deleted projects
func (p *ProjectModel) GetCount() (int64, error) {
	ctx := context.Background()
	return p.queries.GetProjectsCount(ctx)
}

// GetAll retrieves all projects with their client information
func (p *ProjectModel) GetAll() ([]ProjectWithClient, error) {
	ctx := context.Background()
	rows, err := p.queries.GetAllProjectsWithClient(ctx)
	if err != nil {
		return nil, err
	}
	
	projects := make([]ProjectWithClient, len(rows))
	for i, row := range rows {
		project, err := p.convertRowToProjectWithClient(row)
		if err != nil {
			return nil, err
		}
		projects[i] = project
	}
	
	return projects, nil
}

// convertRowToProjectWithClient converts a database row to a ProjectWithClient struct
func (p *ProjectModel) convertRowToProjectWithClient(row db.GetAllProjectsWithClientRow) (ProjectWithClient, error) {
	// Helper function to convert sql.NullString to string
	nullStringToString := func(ns sql.NullString) string {
		if ns.Valid {
			return ns.String
		}
		return ""
	}
	
	// Helper function to convert sql.NullFloat64 to *float64
	nullFloat64ToFloat := func(nf sql.NullFloat64) *float64 {
		if nf.Valid {
			return &nf.Float64
		}
		return nil
	}
	
	// Helper function to convert sql.NullString date to *time.Time
	nullStringToTime := func(ns sql.NullString) *time.Time {
		if !ns.Valid || ns.String == "" {
			return nil
		}
		if t, err := time.Parse("2006-01-02", ns.String); err == nil {
			return &t
		}
		return nil
	}
	
	return ProjectWithClient{
		ID:                     int(row.ID),
		Name:                   row.Name,
		ClientID:               int(row.ClientID),
		ClientName:             row.ClientName,
		Status:                 row.Status,
		HourlyRate:             row.HourlyRate,
		Deadline:               nullStringToTime(row.Deadline),
		ScheduledStart:         nullStringToTime(row.ScheduledStart),
		InvoiceCCEmail:         nullStringToString(row.InvoiceCcEmail),
		InvoiceCCDescription:   nullStringToString(row.InvoiceCcDescription),
		ScheduleComments:       nullStringToString(row.ScheduleComments),
		AdditionalInfo:         nullStringToString(row.AdditionalInfo),
		AdditionalInfo2:        nullStringToString(row.AdditionalInfo2),
		DiscountPercent:        nullFloat64ToFloat(row.DiscountPercent),
		DiscountReason:         nullStringToString(row.DiscountReason),
		AdjustmentAmount:       nullFloat64ToFloat(row.AdjustmentAmount),
		AdjustmentReason:       nullStringToString(row.AdjustmentReason),
		CurrencyDisplay:        row.CurrencyDisplay,
		CurrencyConversionRate: row.CurrencyConversionRate,
		FlatFeeInvoice:         row.FlatFeeInvoice != 0,
		Notes:                  nullStringToString(row.Notes),
		Updated:                row.UpdatedAt,
		Created:                row.CreatedAt,
	}, nil
}

// ProjectModelInterface defines the interface for project operations
type ProjectModelInterface interface {
	Insert(project Project) (int, error)
	Get(id int) (Project, error)
	GetByClient(clientID int) ([]Project, error)
	GetAll() ([]ProjectWithClient, error)
	GetWithPagination(limit, offset int64) ([]ProjectWithClient, error)
	GetCount() (int64, error)
	Update(project Project) error
	Delete(id int) error
}

// Ensure implementation satisfies the interface
var _ ProjectModelInterface = (*ProjectModel)(nil)