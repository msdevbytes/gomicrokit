package model

import (
	"time"

	uuid "github.com/gofrs/uuid/v5"
	"gorm.io/gorm"
)

type UUIDBaseModel struct {
	ID        string         `json:"id" gorm:"primaryKey;unique;size:191;"`
	Status    *bool          `json:"status" gorm:"default:false;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type BaseModel struct {
	ID        int            `json:"id" gorm:"primaryKey;unique;"`
	Status    *bool          `json:"status" gorm:"default:true;"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (base *UUIDBaseModel) BeforeCreate(db *gorm.DB) error {
	if base.ID == "" {
		db.Statement.SetColumn("ID", uuid.Must(uuid.NewV7()).String())
	}
	return nil
}
