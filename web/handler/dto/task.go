package dto

import "time"

type TaskInfoResponse struct {
	ID           int        `json:"id"`
	ResourceID   int        `json:"resource_id,omitempty"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Attempts     int        `json:"attempts"`
	MaxAttempts  int        `json:"max_attempts"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
