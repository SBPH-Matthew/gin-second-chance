package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type categoryStatus string

const (
	ACTIVE   categoryStatus = "ACTIVE"
	INACTIVE categoryStatus = "INACTIVE"
	DRAFT    categoryStatus = "DRAFT"
)

func (cs *categoryStatus) Scan(value any) error {
	if value == nil {
		*cs = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*cs = categoryStatus(v)
	case string:
		*cs = categoryStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into categoryStatus", value)
	}

	return nil
}

func (cs categoryStatus) Value() (driver.Value, error) {
	switch cs {
	case ACTIVE, INACTIVE, DRAFT, "":
		return string(cs), nil
	default:
		return nil, fmt.Errorf("invalid categoryStatus: %s", cs)
	}
}

func (cs categoryStatus) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", cs)), nil
}

func (cs *categoryStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*cs = categoryStatus(s)
	return nil
}

type category struct {
	ID     uint           `gorm:"primaryKey"`
	Name   string         `gorm:"not null"`
	Status categoryStatus `gorm:"type:category_status"`
}
