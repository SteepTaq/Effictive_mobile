package model

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          int64      `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int64      `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}
