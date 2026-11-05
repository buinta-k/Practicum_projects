package api

import (
	"encoding/json"
	"net/http"
)

func writeJson(res http.ResponseWriter, data any) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(res).Encode(data)
}

func writeError(res http.ResponseWriter, message string, status int) {
	res.Header().Set("Content-Type", "application/json; charset=UTF-8")
	res.WriteHeader(status)
	_ = json.NewEncoder(res).Encode(map[string]string{
		"error": message,
	})
}
