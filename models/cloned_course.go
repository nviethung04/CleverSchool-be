package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type CloneInfo struct {
	ProgramId int64      `json:"program_id"`
	CourseId int64      `json:"course_id"`
	CloneId  int64      `json:"clone_id"`
	ClonedAt *time.Time `json:"cloned_at"`
}

// Scan để GORM scan từ DB JSONB
func (c *CloneInfo) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan CloneInfo: type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, c)
}

// Value để GORM lưu vào DB
func (c CloneInfo) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// UnmarshalJSON handle cả trường hợp rỗng hoặc "" cho ClonedAt
func (c *CloneInfo) UnmarshalJSON(data []byte) error {
	type Alias CloneInfo
	aux := &struct {
		ClonedAt *string `json:"cloned_at"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.ClonedAt != nil && *aux.ClonedAt != "" {
		t, err := time.Parse(time.RFC3339, *aux.ClonedAt)
		if err != nil {
			return err
		}
		c.ClonedAt = &t
	} else {
		c.ClonedAt = nil
	}

	return nil
}
