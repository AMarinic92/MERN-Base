package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Card struct {
	ID        string         `gorm:"primaryKey;type:varchar(255)"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	OracleID   *string  `gorm:"type:varchar(255)"`
	Name       string   `gorm:"type:varchar(500);not null"`
	ManaCost   *string  `gorm:"type:varchar(100)"`
	CMC        *float64 `gorm:"type:decimal(10,2)"`
	TypeLine   string   `gorm:"type:varchar(500);not null"`
	OracleText *string  `gorm:"type:text"`
	Power      *string  `gorm:"type:varchar(20)"`
	Toughness  *string  `gorm:"type:varchar(20)"`
	Loyalty    *string  `gorm:"type:varchar(20)"`

	// Array fields - use pq.StringArray
	Colors        pq.StringArray `gorm:"type:text[]"`
	ColorIdentity pq.StringArray `gorm:"type:text[]"`
	Keywords      pq.StringArray `gorm:"type:text[]"`

	// JSON fields stored as text
	CardFaces  *string `gorm:"type:jsonb"`
	ImageURIs  *string `gorm:"type:jsonb"`
	Legalities *string `gorm:"type:jsonb"`
	Prices     *string `gorm:"type:jsonb"`

	SetCode         string  `gorm:"type:varchar(50);not null"`
	SetName         *string `gorm:"type:varchar(500)"`
	CollectorNumber *string `gorm:"type:varchar(50)"`
	Rarity          string  `gorm:"type:varchar(50);not null"`
	Artist          *string `gorm:"type:varchar(500)"`
	FlavorText      *string `gorm:"type:text"`
	ReleasedAt      *string `gorm:"type:varchar(50)"`
	Lang            string  `gorm:"type:varchar(10);default:''"`
	CachedAt        int64   `gorm:"type:bigint;default:0"`
}