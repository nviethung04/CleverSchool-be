package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type SettingKVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SettingValues []SettingKVPair

func (v *SettingValues) Scan(value interface{}) error {
	if value == nil {
		*v = []SettingKVPair{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan SettingValues: wrong type %T", value)
	}
	if len(bytes) == 0 {
		*v = []SettingKVPair{}
		return nil
	}
	return json.Unmarshal(bytes, v)
}

func (v SettingValues) Value() (driver.Value, error) {
	if v == nil {
		return json.Marshal([]SettingKVPair{})
	}
	return json.Marshal(v)
}

type Setting struct {
	ID        int64         `gorm:"primaryKey" json:"id"`
	Key       string        `gorm:"column:key;size:255;not null" json:"key"`
	Name      string        `json:"name"`
	Values    SettingValues `gorm:"type:jsonb;default:'[]'" json:"values"`
	IsActive  bool          `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	CreatedBy int64         `json:"created_by"`
	UpdatedBy int64         `json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
	DeletedBy int64          `gorm:"column:deleted_by"`
}
