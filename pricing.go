package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"ski-rental/internal/store"
)

type SkiPrice struct {
	SkiID       string `json:"skiId"`
	Model       string `json:"model"`
	HourlyCents int    `json:"hourlyCents"`
	DailyCents  int    `json:"dailyCents"`
}

func registerPricingAPI(skiStore *store.Store) func(string, string, int) (int, bool) {
	var mu sync.RWMutex
	customPrices := make(map[string]SkiPrice)

	getPrices := func() []SkiPrice {
		skis := skiStore.GetSkis()

		mu.RLock()
		defer mu.RUnlock()

		result := make([]SkiPrice, 0, len(skis))

		for _, ski := range skis {
			price := SkiPrice{
				SkiID:       ski.ID,
				Model:       ski.Model,
				HourlyCents: 500,
				DailyCents:  2000,
			}

			if custom, ok := customPrices[ski.ID]; ok {
				price.HourlyCents = custom.HourlyCents
				price.DailyCents = custom.DailyCents
			}

			result = append(result, price)
		}

		return result
	}

	quote := func(skiID, unit string, quantity int) (int, bool) {
		if quantity < 1 ||
			(unit == "hours" && quantity > 72) ||
			(unit == "days" && quantity > 365) ||
			(unit != "hours" && unit != "days") {
			return 0, false
		}

		for _, price := range getPrices() {
			if price.SkiID == skiID {
				total := price.DailyCents * quantity

				if unit == "hours" {
					total = price.HourlyCents * quantity
				}

				if unit == "days" && quantity >= 7 {
					total = total * 90 / 100
				}

				return total, true
			}
		}

		return 0, false
	}

	http.HandleFunc("/api/prices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(getPrices())
			return
		}

		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input SkiPrice

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if input.HourlyCents < 1 ||
			input.DailyCents < 1 ||
			input.HourlyCents > 1000000 ||
			input.DailyCents > 1000000 {
			http.Error(w, "Invalid prices", http.StatusBadRequest)
			return
		}

		found := false

		for _, ski := range skiStore.GetSkis() {
			if ski.ID == input.SkiID {
				input.Model = ski.Model
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Ski not found", http.StatusNotFound)
			return
		}

		mu.Lock()
		customPrices[input.SkiID] = input
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(input)
	})

	http.HandleFunc("/api/quote", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		quantity, err := strconv.Atoi(r.URL.Query().Get("quantity"))
		unit := r.URL.Query().Get("unit")

		if err != nil || quantity < 1 ||
			(unit == "hours" && quantity > 72) ||
			(unit == "days" && quantity > 365) ||
			(unit != "days" && unit != "hours") {
			http.Error(w, "Invalid duration", http.StatusBadRequest)
			return
		}

		skiID := r.URL.Query().Get("skiId")

		for _, price := range getPrices() {
			if price.SkiID == skiID {
				amount := price.DailyCents * quantity

				if unit == "hours" {
					amount = price.HourlyCents * quantity

				}

				if unit == "days" && quantity >= 7 {

					amount = amount * 90 / 100
				}

				w.Header().Set("Content-Type", "application/json")

				json.NewEncoder(w).Encode(map[string]any{

					"totalCents": amount,
					"currency":   "EUR",
				})
				return
			}
		}

		http.Error(w, "Ski not found", http.StatusNotFound)

	})
	return quote
}
