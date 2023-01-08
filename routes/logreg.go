package routes

import (
	"Backend/handlers"
	"Backend/pkg/middleware"
	"Backend/pkg/mysql"
	"Backend/repositories"

	"github.com/gorilla/mux"
)

func LogRegRoutes(r *mux.Router) {
	userRepository := repositories.RepositoryUser(mysql.DB)
	h := handlers.HandlerLogReg(userRepository)

	r.HandleFunc("/register", h.Register).Methods("POST")
	r.HandleFunc("/Login", h.Login).Methods("POST")
	r.HandleFunc("/check-auth", middleware.Auth(h.CheckAuth)).Methods("get")
}
