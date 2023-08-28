package database

import (
	"time"

	"gorm.io/gorm"
)

type CustomBaseModel struct {
	ID        uint           `json:"-" gorm:"primarykey"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type User struct {
	CustomBaseModel
	Login        string `gorm:"not null;unique;default:null"`
	PasswordHash string
	Orders       []Order
	Withdrawals  []Withdraw
}

type OrderStatus string

const (
	NEW        OrderStatus = "NEW"
	PROCESSING OrderStatus = "PROCESSING"
	INVALID    OrderStatus = "INVALID"
	PROCESSED  OrderStatus = "PROCESSED"
)

type Order struct {
	CustomBaseModel
	Number    string      `json:"number" gorm:"not null;unique;default:null"`
	Status    OrderStatus `json:"status" gorm:"default:NEW"`
	Accrual   float64     `json:"accrual,omitempty"`
	UserID    uint        `json:"-"`
	CreatedAt time.Time   `json:"uploaded_at"`
}

type Withdraw struct {
	CustomBaseModel
	Order     int       `json:"order" gorm:"not null;unique;default:null"`
	Sum       float64   `json:"sum" gorm:"not null;default:null"`
	UserID    uint      `json:"-"`
	CreatedAt time.Time `json:"processed_at"`
}
