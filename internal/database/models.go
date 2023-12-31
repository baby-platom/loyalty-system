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
	Balance      Balance
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
	Accrual   float32     `json:"accrual,omitempty"`
	UserID    uint        `json:"-"`
	CreatedAt time.Time   `json:"uploaded_at"`
}

type Withdraw struct {
	CustomBaseModel
	Order     string    `json:"order" gorm:"not null;unique;default:null"`
	Sum       float32   `json:"sum" gorm:"not null;default:null"`
	UserID    uint      `json:"-"`
	CreatedAt time.Time `json:"processed_at"`
}

type Balance struct {
	CustomBaseModel
	Accumulated float32 `gorm:"not null;default:0"`
	Withdrawn   float32 `gorm:"not null;default:0"`
	UserID      uint
}
