package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type SettingValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Setting struct {
	ID         int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Key        string        `json:"key" gorm:"not null;index"`
	Name       string        `json:"name" gorm:"size:255;not null"`
	Values     SettingValues `json:"values" gorm:"type:jsonb;not null;default:('[]'::jsonb)"`
	IsActive   bool          `json:"is_active" gorm:"default:true"`
	IsInternal bool          `json:"is_internal" gorm:"default:false"`

	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedBy *int64     `json:"created_by"`
	UpdatedBy *int64     `json:"updated_by"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
	DeletedBy *int64     `json:"deleted_by"`
}

type SettingValues []SettingValue

func (sv SettingValues) Value() (driver.Value, error) {
	if len(sv) == 0 {
		return "[]", nil
	}
	b, err := json.Marshal(sv)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (sv *SettingValues) Scan(src interface{}) error {
	if src == nil {
		*sv = SettingValues{}
		return nil
	}

	switch s := src.(type) {
	case []byte:
		return json.Unmarshal(s, sv)
	case string:
		return json.Unmarshal([]byte(s), sv)
	default:
		return fmt.Errorf("unsupported Scan type for SettingValues: %T", src)
	}
}
