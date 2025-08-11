package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Rules struct {
	ID       uint   `gorm:"primaryKey"`
	Rule     []byte `gorm:"type:jsonb;not null"`
	IsActive bool   `gorm:"not null;default:false"`
	UserID   uint
	User     User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type List struct {
	ID          uint   `gorm:"primaryKey"`
	List_name   string `gorm:"not null"`
	Description string
	Phrases     StrSlice `gorm:"type:text"`
	UserID      uint
	User        User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Regulars struct {
	ID          uint   `gorm:"primaryKey"`
	RegName     string `gorm:"not null"`
	Description string
	Regs        StrSlice `gorm:"type:text"`
	UserID      uint
	User        User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type User struct {
	ID         uint   `gorm:"primaryKey"`
	Username   string `gorm:"unique;not null"`
	Password   string `gorm:"not null"`
	FullName   string
	Email      string
	JWT        string
	UserGroup  string `gorm:"not null"`
	IsActive   bool   `gorm:"not null;default:false"`
	ChangePass bool   `gorm:"not null;default:false"`
	GroupID    uint
	Group      Group `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Group struct {
	ID          uint   `gorm:"primaryKey"`
	GroupName   string `gorm:"not null"`
	Description string
	Permissions JSONMap `gorm:"type:jsonb;not null"`
	IsActive    bool    `gorm:"not null;default:false"`
}

type Alert struct {
	ID        uint      `gorm:"primaryKey"`
	Timestamp time.Time `gorm:"not null"`
	Text      []byte    `gorm:"type:jsonb;not null"`
	ExpiresAt time.Time
}

// GORM JSONB support
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONMap: value is not []byte, got %T", value)
	}
	return json.Unmarshal(bytes, j)
}

// GORM Slice of strings support
type StrSlice []string

func (s StrSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return nil, nil
	}
	return strings.Join(s, ","), nil
}

func (s *StrSlice) Scan(value interface{}) error {
	bytes, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed convert to []string: value is not []byte, got %T", value)
	}
	*s = strings.Split(string(bytes), ",")
	return nil
}
