package store

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"ski-rental/internal/domain"
)

type Store struct {
	mu        sync.RWMutex
	skis      []domain.Ski
	persons   []domain.Person
	rentals   []domain.Rental
	documents map[string][]byte
}

func NewStore() *Store {
	return &Store{
		skis: []domain.Ski{
			{ID: "S001", Model: "Atomic Redster", Available: true},
			{ID: "S002", Model: "Head Supershape", Available: true},
			{ID: "S003", Model: "Fischer RC4", Available: true},
		},
	}
}

func (s *Store) SkiCounts() (int, int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.skis)
	available := 0

	for _, ski := range s.skis {
		if !s.isBusyToday(ski.ID) {
			available++
		}
	}

	rented := total - available

	return available, rented, total
}

func (s *Store) GetSkis() []domain.Ski {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Ski, len(s.skis))
	copy(result, s.skis)

	for i := range result {
		result[i].Available = !s.isBusyToday(result[i].ID)
	}

	return result
}

func (s *Store) AddSki(model string) (domain.Ski, bool) {
	model = strings.TrimSpace(model)

	if model == "" {
		return domain.Ski{}, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	ski := domain.Ski{
		ID:        fmt.Sprintf("S%03d", len(s.skis)+1),
		Model:     model,
		Available: true,
	}

	s.skis = append(s.skis, ski)

	return ski, true
}

func (s *Store) AddPerson(firstName, lastName, documentType, documentNr string) (domain.Person, bool) {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	documentType = strings.TrimSpace(documentType)
	documentNr = strings.TrimSpace(documentNr)

	if firstName == "" || lastName == "" || documentType == "" || documentNr == "" {
		return domain.Person{}, false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	person := domain.Person{
		ID:           fmt.Sprintf("P%03d", len(s.persons)+1),
		FirstName:    firstName,
		LastName:     lastName,
		DocumentType: documentType,
		DocumentNr:   documentNr,
	}

	s.persons = append(s.persons, person)

	return person, true
}

func (s *Store) GetPersons() []domain.Person {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Person, len(s.persons))
	copy(result, s.persons)

	return result
}

func (s *Store) PersonCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.persons)
}

func (s *Store) CreateRental(personID, skiID string, startDate time.Time, durationDays int) (domain.Rental, bool) {
	today := time.Now().Format("2006-01-02")

	if startDate.IsZero() ||
		durationDays <= 0 ||
		durationDays > 365 ||
		startDate.Format("2006-01-02") < today {
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

	if !skiExists {
		return domain.Rental{}, false
	}

	if s.hasPeriodConflict(skiID, startDate, durationDays) {
		return domain.Rental{}, false
	}

	rental := domain.Rental{
		ID:           fmt.Sprintf("R%03d", len(s.rentals)+1),
		PersonID:     personID,
		SkiID:        skiID,
		StartDate:    startDate,
		DurationDays: durationDays,
		Returned:     false,
	}

	s.rentals = append(s.rentals, rental)

	return rental, true
}

func (s *Store) GetRentals() []domain.Rental {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Rental, len(s.rentals))
	copy(result, s.rentals)

	return result
}

func (s *Store) ReturnRental(rentalID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.rentals {
		if s.rentals[i].ID == rentalID {
			if s.rentals[i].Returned {
				return false
			}

			for j := range s.skis {
				if s.skis[j].ID == s.rentals[i].SkiID {
					s.rentals[i].Returned = true
					s.skis[j].Available = true
					return true
				}
			}

			return false
		}
	}

	return false
}

func (s *Store) isBusyToday(skiID string) bool {
	now := time.Now()

	for _, rental := range s.rentals {
		if rental.SkiID == skiID &&
			!rental.Returned &&
			!rental.StartDate.After(now) {
			return true
		}
	}

	return false
}

func (s *Store) hasPeriodConflict(skiID string, start time.Time, days int) bool {

	return s.overlaps(skiID, start, start.AddDate(0, 0, days))
}

func (s *Store) AvailableSkis(start time.Time, days int) []domain.Ski {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.Ski, 0)

	if start.IsZero() || days <= 0 || days > 365 {
		return result
	}

	for _, ski := range s.skis {
		if !s.hasPeriodConflict(ski.ID, start, days) {
			ski.Available = true
			result = append(result, ski)
		}
	}

	return result
}

func (s *Store) SaveTestDocument(personID string, data []byte) bool {
	if len(data) == 0 || len(data) > 2*1024*1024 {
		return false
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
		return false
	}

	if s.documents == nil {
		s.documents = make(map[string][]byte)
	}

	s.documents[personID] = append([]byte(nil), data...)

	return true
}
