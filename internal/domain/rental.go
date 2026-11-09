package domain

import "time"

type Rental struct {
	ID           string
	PersonID     string
	SkiID        string
	StartDate    time.Time
	DurationDays int
	Quantity     int
	Unit         string
	TotalCents   int
	Returned     bool
}
