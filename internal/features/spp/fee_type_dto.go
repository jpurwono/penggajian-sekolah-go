package spp

import "time"

type FeeTypeResponse struct {
	ID            int64     `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Frequency     string    `json:"frequency"`
	FrequencyLabel string   `json:"frequency_label"`
	DefaultAmount int64     `json:"default_amount"`
	Description   *string   `json:"description"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type SaveFeeTypeRequest struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Frequency     string  `json:"frequency"`
	DefaultAmount int64   `json:"default_amount"`
	Description   *string `json:"description"`
	IsActive      *bool   `json:"is_active"`
}

type GenerateYearlyRequest struct {
	Year           int                  `json:"year"`
	FeeTypeIDs     []int64              `json:"fee_type_ids"`
	CustomAmounts  map[int64]int64      `json:"custom_amounts,omitempty"`
	ClassName      *string              `json:"class_name,omitempty"`
	DueDate        *string              `json:"due_date,omitempty"`
}

type GenerateYearlyResult struct {
	TotalCreated int    `json:"total_created"`
	TotalSkipped int    `json:"total_skipped"`
	Details      []GenerateDetail `json:"details"`
	Message      string `json:"message"`
}

type GenerateDetail struct {
	FeeTypeCode string `json:"fee_type_code"`
	FeeTypeName string `json:"fee_type_name"`
	Created     int    `json:"created"`
	Skipped     int    `json:"skipped"`
	Amount      int64  `json:"amount"`
}

// Update BillResponse untuk include fee_type
