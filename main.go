package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"ski-rental/internal/store"
)

type DashboardData struct {
	AvailableSkis     int `json:"availableSkis"`
	RentedSkis        int `json:"rentedSkis"`
	TotalSkis         int `json:"totalSkis"`
	RegisteredPersons int `json:"registeredPersons"`
}

func main() {
	skiStore := store.NewStore()

	getQuote := registerPricingAPI(skiStore)

	http.HandleFunc("/api/dashboard", func(w http.ResponseWriter, r *http.Request) {

		available, rented, total := skiStore.SkiCounts()

		data := DashboardData{
			AvailableSkis:     available,
			RentedSkis:        rented,
			TotalSkis:         total,
			RegisteredPersons: skiStore.PersonCount(),
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(data); err != nil {
			fmt.Println("JSON error:", err)
		}
	})

	http.HandleFunc("/api/skis", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(skiStore.GetSkis())
			return
		}

		if r.Method == http.MethodPost {
			var input struct {
				Model string `json:"model"`
			}

			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			ski, ok := skiStore.AddSki(input.Model)
			if !ok {
				http.Error(w, "Model is required", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ski)
			return
		}

		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/persons", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(skiStore.GetPersons())
			return
		}

		if r.Method == http.MethodPost {
			var input struct {
				FirstName    string `json:"firstName"`
				LastName     string `json:"lastName"`
				DocumentType string `json:"documentType"`
				DocumentNr   string `json:"documentNr"`
			}

			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}

			person, ok := skiStore.AddPerson(
				input.FirstName,
				input.LastName,
				input.DocumentType,
				input.DocumentNr,
			)

			if !ok {
				http.Error(w, "All fields are required", http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(person)
			return
		}

		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/rentals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(skiStore.GetRentals())
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input struct {
			PersonID  string `json:"personId"`
			SkiID     string `json:"skiId"`
			StartDate string `json:"startDate"`
			Quantity  int    `json:"quantity"`
			Unit      string `json:"unit"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		start, err := time.ParseInLocation(
			"2006-01-02T15:04",
			input.StartDate,
			time.Local,
		)

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		totalCents, ok := getQuote(input.SkiID, input.Unit, input.Quantity)

		if !ok {
			http.Error(w, "Invalid price or duration", http.StatusBadRequest)
			return
		}

		rental, ok := skiStore.CreatePricedRental(
			input.PersonID,
			input.SkiID,
			start,
			input.Quantity,
			input.Unit,
			totalCents,
		)

		if !ok {
			http.Error(w, "Rental could not be created", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rental)
	})

	http.HandleFunc("/api/rentals/return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input struct {
			RentalID string `json:"rentalId"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if !skiStore.ReturnRental(input.RentalID) {
			http.Error(w, "Rental not found or already returned", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Rental returned successfully",
		})
	})

	http.HandleFunc("/api/skis/available", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		start, err := time.ParseInLocation(
			"2006-01-02T15:04",
			r.URL.Query().Get("startDate"),
			time.Local,
		)

		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		quantity, err := strconv.Atoi(r.URL.Query().Get("quantity"))
		unit := r.URL.Query().Get("unit")

		if err != nil || quantity < 1 ||
			(unit != "hours" && unit != "days") {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}

		skis := skiStore.AvailableSkisFor(start, quantity, unit)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skis)
	})

	http.HandleFunc("/api/persons/document", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 3<<20)

		if err := r.ParseMultipartForm(2 << 20); err != nil {
			http.Error(w, "Invalid upload", http.StatusBadRequest)
			return
		}

		personID := r.FormValue("personId")

		file, _, err := r.FormFile("document")
		if err != nil {
			http.Error(w, "Document is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		data, err := io.ReadAll(io.LimitReader(file, 2*1024*1024+1))

		if err != nil || len(data) == 0 || len(data) > 2*1024*1024 {
			http.Error(w, "Invalid file size", http.StatusBadRequest)
			return
		}

		contentType := http.DetectContentType(data)

		if contentType != "application/pdf" &&
			contentType != "image/jpeg" &&
			contentType != "image/png" {
			http.Error(w, "Only PDF, JPG and PNG are allowed", http.StatusBadRequest)
			return
		}

		if !skiStore.SaveTestDocument(personID, data) {
			http.Error(w, "Person not found or invalid file", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]string{
			"message": "Document uploaded successfully",
		})
	})

	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	fmt.Println("Ski Rental started on http://localhost:8080")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
