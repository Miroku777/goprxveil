package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Miroku77/goprxveil/internal/storage"
)

// GET /api/analytics/{id} - accepts id url, return all products {"seller_name":"", "name":"", "rating":"", "description":"", "options":"", "current_price":"", "history":""}
func GetAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	idUrl, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/analytics/"))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	productId, analitics, err := storage.FindProductsByUrlid(db, idUrl)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	currentPrice, history, err := storage.FindPriceHistoryByProductId(db, productId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	analitics.CurrentPrice = currentPrice
	analitics.History = history
	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(analitics); err != nil {
		log.Println("Error response JSON:", err)
	}
}

// GET /api/analytics/group/{id} - accepts id group from path for group urls analotics, return{ "name": "Ноутбук MacBook", "price": "1623" },
func GetGroupComparisonHandler(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/analytics/group/"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	data, err := storage.GetGroupPrices(db, groupId)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// serves the analytics and charts page (analytics.html)
func AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/analytics/analytics.html")
}
