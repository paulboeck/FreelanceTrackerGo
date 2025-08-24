package models

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/paulboeck/FreelanceTrackerGo/internal/db"
)

// AppSetting represents a configuration setting in the system
type AppSetting struct {
	Key         string
	Value       string
	DataType    string
	Description string
	Created     time.Time
	Updated     time.Time
}

// AppSettingValue provides type-safe access to setting values
type AppSettingValue struct {
	Value    string
	DataType string
}

// AsString returns the setting value as a string
func (sv AppSettingValue) AsString() string {
	return sv.Value
}

// AsInt returns the setting value as an integer
func (sv AppSettingValue) AsInt() (int, error) {
	if sv.DataType != "int" {
		return 0, errors.New("setting is not an integer")
	}
	return strconv.Atoi(sv.Value)
}

// AsFloat returns the setting value as a float64
func (sv AppSettingValue) AsFloat() (float64, error) {
	if sv.DataType != "float" && sv.DataType != "decimal" {
		return 0, errors.New("setting is not a float or decimal")
	}
	return strconv.ParseFloat(sv.Value, 64)
}

// AsDecimal returns the setting value as a float64 (alias for AsFloat)
func (sv AppSettingValue) AsDecimal() (float64, error) {
	return sv.AsFloat()
}

// AsBool returns the setting value as a boolean
func (sv AppSettingValue) AsBool() (bool, error) {
	if sv.DataType != "bool" {
		return false, errors.New("setting is not a boolean")
	}
	return strconv.ParseBool(sv.Value)
}

// AppSettingModel wraps the generated SQLC Queries for setting operations
type AppSettingModel struct {
	queries *db.Queries
}

// NewAppSettingModel creates a new AppSettingModel
func NewAppSettingModel(database *sql.DB) *AppSettingModel {
	return &AppSettingModel{
		queries: db.New(database),
	}
}

// Get retrieves a specific setting by key
func (s *AppSettingModel) Get(key string) (AppSetting, error) {
	ctx := context.Background()
	row, err := s.queries.GetSetting(ctx, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AppSetting{}, ErrNoRecord
		}
		return AppSetting{}, err
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

	setting := AppSetting{
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
func (s *AppSettingModel) GetValue(key string) (AppSettingValue, error) {
	setting, err := s.Get(key)
	if err != nil {
		return AppSettingValue{}, err
	}
	return AppSettingValue{Value: setting.Value, DataType: setting.DataType}, nil
}

// GetString retrieves a string setting value
func (s *AppSettingModel) GetString(key string) (string, error) {
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
func (s *AppSettingModel) GetInt(key string) (int, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsInt()
}

// GetFloat retrieves a float setting value
func (s *AppSettingModel) GetFloat(key string) (float64, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsFloat()
}

// GetDecimal retrieves a decimal setting value
func (s *AppSettingModel) GetDecimal(key string) (float64, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return 0, err
	}
	return value.AsDecimal()
}

// GetBool retrieves a boolean setting value
func (s *AppSettingModel) GetBool(key string) (bool, error) {
	value, err := s.GetValue(key)
	if err != nil {
		return false, err
	}
	return value.AsBool()
}

// GetAll retrieves all settings as a map for bulk access
func (s *AppSettingModel) GetAll() (map[string]AppSettingValue, error) {
	ctx := context.Background()
	rows, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make(map[string]AppSettingValue)
	for _, row := range rows {
		settings[row.Key] = AppSettingValue{
			Value:    row.Value,
			DataType: row.DataType,
		}
	}

	return settings, nil
}

// GetAllDetailed retrieves all settings with full information
func (s *AppSettingModel) GetAllDetailed() ([]AppSetting, error) {
	ctx := context.Background()
	rows, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := make([]AppSetting, len(rows))
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

		settings[i] = AppSetting{
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
func (s *AppSettingModel) UpdateValue(key, value string) error {
	ctx := context.Background()
	params := db.UpdateSettingParams{
		Key:   key,
		Value: value,
	}
	return s.queries.UpdateSetting(ctx, params)
}

// AppSettingModelInterface defines the interface for setting operations
type AppSettingModelInterface interface {
	Get(key string) (AppSetting, error)
	GetValue(key string) (AppSettingValue, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)
	GetFloat(key string) (float64, error)
	GetDecimal(key string) (float64, error)
	GetBool(key string) (bool, error)
	GetAll() (map[string]AppSettingValue, error)
	GetAllDetailed() ([]AppSetting, error)
	UpdateValue(key, value string) error
}

// Ensure implementation satisfies the interface
var _ AppSettingModelInterface = (*AppSettingModel)(nil)