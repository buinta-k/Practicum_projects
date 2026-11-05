package api

import (
	"net/http"
	"project/internal/db"
)

type Api struct {
	DB *db.Base
}

func Init(Base *db.Base) *Api {
	return &Api{
		DB: Base,
	}
}

func (s *Api) Init(mux *http.ServeMux) {
	mux.HandleFunc("POST /rooms", s.CreateRoom)
	mux.HandleFunc("POST /bookings", s.CreateBooking)
	mux.HandleFunc("GET /rooms", s.GetRooms)
	mux.HandleFunc("DELETE /bookings/{id}", s.DeleteBooking)
	mux.HandleFunc("GET /bookings", s.GetBookings)
	mux.HandleFunc("GET /rooms/available", s.GetAvailableRooms)
}
