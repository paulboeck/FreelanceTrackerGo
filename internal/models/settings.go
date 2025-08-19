package models

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// Setting represents a configuration setting in the system
type Setting struct {
	Key         string
	Value       string
	DataType    string
	Description string
	Created     time.Time
	Updated     time.Time
}

// SettingValue provides type-safe access to setting values
type SettingValue struct {
	Value    string
	DataType string
}

// AsString returns the setting value as a string
func (sv SettingValue) AsString() string {
	return sv.Value
}

// AsInt returns the setting value as an integer
func (sv SettingValue) AsInt() (int, error) {
	if sv.DataType != "int" {
		return 0, errors.New("setting is not an integer")
	}
	return strconv.Atoi(sv.Value)
}

// AsFloat returns the setting value as a float64
func (sv SettingValue) AsFloat() (float64, error) {
	if sv.DataType != "float" && sv.DataType != "decimal" {
		return 0, errors.New("setting is not a float or decimal")
	}
	return strconv.ParseFloat(sv.Value, 64)
}

// AsDecimal returns the setting value as a float64 (alias for AsFloat)
func (sv SettingValue) AsDecimal() (float64, error) {
	return sv.AsFloat()
}

// AsBool returns the setting value as a boolean
func (sv SettingValue) AsBool() (bool, error) {
	if sv.DataType != "bool" {
		return false, errors.New("setting is not a boolean")
	}
	return strconv.ParseBool(sv.Value)
}

// SettingModel wraps the generated SQLC Queries for setting operations
type SettingModel struct {
	queries *db.Queries
}

// NewSettingModel creates a new SettingModel
func NewSettingModel(database *sql.DB) *SettingModel {
	return &SettingModel{
		queries: db.New(database),
	}
}

// Get retrieves a specific setting by key
func (s *SettingModel) Get(key string) (Setting, error) {
	ctx := context.Background()
	row, err := s.queries.GetSetting(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Setting{}, ErrNoRecord
		}
		return Setting{}, err
	}

	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	var created, updated time.Time
	if row.CreatedAt.Valid {
		created = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		updated = row.UpdatedAt.Time
	}

	setting := Setting{
		Key:         row.Key,
		Value:       row.Value,
		DataType:    row.DataType,
		Description: description,
		Created:     created,
		Updated:     updated,
	}

	return setting, nil
}

// GetValue retrieves a setting value with type information
func (s *SettingModel) GetValue(key string) (SettingValue, error) {
	setting, err := s.Get(key)
	if err != nil {
		return SettingValue{}, err
	}
	return SettingValue{Value: setting.Value, DataType: setting.DataType}, nil
}

// GetString retrieves a string setting value
func (s *SettingModel) GetString(key string) (string, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return "", err
	}
	if value.DataType != "string" {
		return "", errors.New("setting is not a string")
	}
	return value.AsString(), nil
}

// GetInt retrieves an integer setting value
func (s *SettingModel) GetInt(key string) (int, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsInt()
}

// GetFloat retrieves a float setting value
func (s *SettingModel) GetFloat(key string) (float64, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsFloat()
}

// GetDecimal retrieves a decimal setting value
func (s *SettingModel) GetDecimal(key string) (float64, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsDecimal()
}

// GetBool retrieves a boolean setting value
func (s *SettingModel) GetBool(key string) (bool, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return false, err
	}
	return value.AsBool()
}

// GetAll retrieves all settings as a map for bulk access
func (s *SettingModel) GetAll() (map[string]SettingValue, error) {
	ctx := context.Background()
	rows, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make(map[string]SettingValue)
	for _, row := range rows {
		settings[row.Key] = SettingValue{
			Value:    row.Value,
			DataType: row.DataType,
		}
	}

	return settings, nil
}

// GetAllDetailed retrieves all settings with full information
func (s *SettingModel) GetAllDetailed() ([]Setting, error) {
	ctx := context.Background()
	rows, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make([]Setting, len(rows))
	for i, row := range rows {
		description := ""
		if row.Description.Valid {
			description = row.Description.String
		}

		var created, updated time.Time
		if row.CreatedAt.Valid {
			created = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			updated = row.UpdatedAt.Time
		}

		settings[i] = Setting{
			Key:         row.Key,
			Value:       row.Value,
			DataType:    row.DataType,
			Description: description,
			Created:     created,
			Updated:     updated,
		}
	}

	return settings, nil
}

// UpdateValue modifies only the value of an existing setting
func (s *SettingModel) UpdateValue(key, value string) error {
	ctx := context.Background()
	params := db.UpdateSettingParams{
		Key:   key,
		Value: value,
	}
	return s.queries.UpdateSetting(ctx, params)
}

// SettingModelInterface defines the interface for setting operations
type SettingModelInterface interface {
	Get(key string) (Setting, error)
	GetValue(key string) (SettingValue, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)
	GetFloat(key string) (float64, error)
	GetDecimal(key string) (float64, error)
	GetBool(key string) (bool, error)
	GetAll() (map[string]SettingValue, error)
	GetAllDetailed() ([]Setting, error)
	UpdateValue(key, value string) error
}

// Ensure implementation satisfies the interface
var _ SettingModelInterface = (*SettingModel)(nil)