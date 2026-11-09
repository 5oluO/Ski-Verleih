package store

import (
	"fmt"
	"time"

	"ski-rental/internal/domain"
)

func periodEnd(start time.Time, unit string, quantity int) (time.Time, bool) {
	if unit == "days" && quantity >= 1 && quantity <= 365 {
		return start.AddDate(0, 0, quantity), true
	}

	if unit == "hours" && quantity >= 1 && quantity <= 72 {
		return start.Add(time.Duration(quantity) * time.Hour), true
	}

	return time.Time{}, false
}

func (s *Store) overlaps(skiID string, start, end time.Time) bool {
	now := time.Now()

	for _, old := range s.rentals {
		if old.SkiID != skiID || old.Returned {
			continue
		}

		// Already issued skis remain unavailable until returned.
		if !old.StartDate.After(now) {
			return true
		}

		unit := old.Unit
		quantity := old.Quantity

		if unit == "" {
			unit = "days"
			quantity = old.DurationDays
		}

		oldEnd, ok := periodEnd(old.StartDate, unit, quantity)

		if ok && start.Before(oldEnd) && old.StartDate.Before(end) {
			return true
		}
	}

	return false
}

func (s *Store) AvailableSkisFor(start time.Time, quantity int, unit string) []domain.Ski {
	end, ok := periodEnd(start, unit, quantity)

	if !ok || start.Before(time.Now().Add(-time.Minute)) {
		return []domain.Ski{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Ski, 0)

	for _, ski := range s.skis {
		if !s.overlaps(ski.ID, start, end) {
			ski.Available = true
			result = append(result, ski)
		}
	}

	return result
}

func (s *Store) CreatePricedRental(
	personID, skiID string,
	start time.Time,
	quantity int,
	unit string,
	totalCents int,
) (domain.Rental, bool) {

	end, ok := periodEnd(start, unit, quantity)

	if !ok || totalCents <= 0 ||
		start.Before(time.Now().Add(-time.Minute)) {
		return domain.Rental{}, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	personExists := false

	for _, person := range s.persons {
		if person.ID == personID {
			personExists = true
			break
		}
	}

	if !personExists {
		return domain.Rental{}, false
	}

	skiExists := false

	for _, ski := range s.skis {
		if ski.ID == skiID {
			skiExists = true
			break
		}
	}

	if !skiExists || s.overlaps(skiID, start, end) {
		return domain.Rental{}, false
	}

	days := 0
	if unit == "days" {
		days = quantity
	}

	rental := domain.Rental{
		ID:           fmt.Sprintf("R%03d", len(s.rentals)+1),
		PersonID:     personID,
		SkiID:        skiID,
		StartDate:    start,
		DurationDays: days,
		Quantity:     quantity,
		Unit:         unit,
		TotalCents:   totalCents,
		Returned:     false,
	}

	s.rentals = append(s.rentals, rental)

	return rental, true
}
