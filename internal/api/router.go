package api

import (
	"log"
	"net/http"

	"github.com/Miroku77/goprxveil/internal/handlers"
	"github.com/Miroku77/goprxveil/internal/middleware"
	"github.com/gorilla/mux"
)

func StartServer() {
	r := mux.NewRouter()

	r.HandleFunc("/login", handlers.LoginPageHandler).Methods("GET")
	r.HandleFunc("/", handlers.DashboardHandler).Methods("GET")
	r.HandleFunc("/analytics", handlers.AnalyticsHandler).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("templates"))))

	r.HandleFunc("/api/register", handlers.RegisterHandler).Methods("POST")
	r.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST")

	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.LoginMiddleware)

	// page dashboard: group and link
	protected.HandleFunc("/api/groups", handlers.ManipulateGroupHandler).Methods("POST", "GET")
	protected.HandleFunc("/api/groups/{id}", handlers.DeleteByIdGroupHandler).Methods("DELETE")
	protected.HandleFunc("/api/urls", handlers.ManipulateURLHandler).Methods("POST", "GET")
	protected.HandleFunc("/api/urls/{id}", handlers.DeleteURLHandler).Methods("DELETE")
	protected.HandleFunc("/api/urls/{id}/active", handlers.UpdateActiveUrlHandler).Methods("PATCH")

	// page  analytics
	protected.HandleFunc("/api/analytics/{id}", handlers.GetAnalyticsHandler).Methods("GET")
	// +  GET /api/urls*
	protected.HandleFunc("/api/analytics/group/{id}", handlers.GetGroupComparisonHandler).Methods("GET")

	log.Println("Сервер запущен на http://localhost:8888")
	if err := http.ListenAndServe(":8888", r); err != nil {
		log.Println(err)
		return
	}
}
