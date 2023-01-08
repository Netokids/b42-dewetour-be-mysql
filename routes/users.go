package routes

import (
	"Backend/handlers"
	"Backend/pkg/mysql"
	"Backend/repositories"

	"github.com/gorilla/mux"
)

func UserRoute(r *mux.Router) {
	userRepository := repositories.RepositoryUser(mysql.DB)
	h := handlers.HandlerUser(userRepository)

	r.HandleFunc("/users", h.FindUser).Methods("GET")
	r.HandleFunc("/users/{id}", h.GetUser).Methods("GET")
	r.HandleFunc("/users", h.AddUser).Methods("POST")
	r.HandleFunc("/users/{id}", h.UpdateUser).Methods("PATCH")
	r.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")
}
