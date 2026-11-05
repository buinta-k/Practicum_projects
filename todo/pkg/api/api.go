package api

import (
	"net/http"
	"project/pkg/db"
)

const dateLayout = "20060102"

type Api struct {
	DB *db.Base
}

func NewApi(Base *db.Base) *Api {
	Api := &Api{
		DB: Base,
	}

	return Api
}

type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

type password struct {
	Password string `json:"password"`
}

func (s *Api) Init(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/nextdate", s.NextDayHandler)
	mux.HandleFunc("POST /api/task", auth(s.TaskHandler))
	mux.HandleFunc("GET /api/tasks", auth(s.TasksHandler))
	mux.HandleFunc("GET /api/task", auth(s.GetHandler))
	mux.HandleFunc("PUT /api/task", auth(s.PutHandler))
	mux.HandleFunc("POST /api/task/done", auth(s.ReadyHandler))
	mux.HandleFunc("DELETE /api/task", auth(s.DeleteHandler))
	mux.HandleFunc("POST /api/signin", s.SignHandler)
}
