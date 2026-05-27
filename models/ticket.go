package models

import (
	"time"
)

type TicketType string

const (
	TicketTypeFree    TicketType = "free"
	TicketTypePaid    TicketType = "paid"
	TicketTypeEarly   TicketType = "early_bird"
)

type Ticket struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EventID      uint       `gorm:"not null" json:"event_id"`
	Name         string     `gorm:"not null" json:"name"`
	Type         TicketType `gorm:"default:free" json:"type"`
	Price        float64    `gorm:"default:0" json:"price"`
	Quantity     int        `gorm:"not null" json:"quantity"`
	SoldCount    int        `gorm:"default:0" json:"sold_count"`
	Description  string     `json:"description"`
	SaleStartAt  *time.Time `json:"sale_start_at"`
	SaleEndAt    *time.Time `json:"sale_end_at"`
	EarlyPrice   float64    `json:"early_price"`
	EarlyStartAt *time.Time `json:"early_start_at"`
	EarlyEndAt   *time.Time `json:"early_end_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"-"`

	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

type PromoCodeType string

const (
	PromoCodeTypePercent PromoCodeType = "percent"
	PromoCodeTypeFixed   PromoCodeType = "fixed"
)

type PromoCode struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	EventID     uint          `gorm:"not null" json:"event_id"`
	Code        string        `gorm:"not null;unique" json:"code"`
	Type        PromoCodeType `gorm:"not null" json:"type"`
	Value       float64       `gorm:"not null" json:"value"`
	MaxUsage    int           `gorm:"default:0" json:"max_usage"`
	UsedCount   int           `gorm:"default:0" json:"used_count"`
	ExpiresAt   *time.Time    `json:"expires_at"`
	IsActive    bool          `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	Event Event `gorm:"foreignKey:EventID" json:"event,omitempty"`
}

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID           uint        `gorm:"primaryKey" json:"id"`
	EventID      uint        `gorm:"not null" json:"event_id"`
	UserID       uint        `json:"user_id"`
	TicketID     uint        `gorm:"not null" json:"ticket_id"`
	OrderNumber  string      `gorm:"not null;unique" json:"order_number"`
	Quantity     int         `gorm:"default:1" json:"quantity"`
	UnitPrice    float64     `gorm:"not null" json:"unit_price"`
	Discount     float64     `gorm:"default:0" json:"discount"`
	TotalAmount  float64     `gorm:"not null" json:"total_amount"`
	Status       OrderStatus `gorm:"default:pending" json:"status"`
	PromoCodeID  *uint       `json:"promo_code_id"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt       *time.Time  `json:"paid_at"`
	InvoiceInfo  string      `json:"invoice_info"`
	RefundReason string      `json:"refund_reason"`
	RefundedAt   *time.Time  `json:"refunded_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt    *time.Time  `gorm:"index" json:"-"`

	Event      Event      `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Ticket     Ticket     `gorm:"foreignKey:TicketID" json:"ticket,omitempty"`
	PromoCode  *PromoCode `gorm:"foreignKey:PromoCodeID" json:"promo_code,omitempty"`
	UserIDVal  uint       `json:"-"`
}

type PaymentRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OrderID       uint      `gorm:"not null" json:"order_id"`
	Amount        float64   `gorm:"not null" json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	Remark        string    `json:"remark"`
	CreatedAt     time.Time `json:"created_at"`

	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}
