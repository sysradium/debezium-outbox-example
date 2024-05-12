package ports

import (
	"encoding/json"
	"net/http"

	"github.com/sysradium/debezium-outbox-example/users-service/internal/app"
	"github.com/sysradium/debezium-outbox-example/users-service/internal/app/commands"
)

type UserCreationRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewHTTP(a app.Application) *Server {
	return &Server{
		a: a,
	}
}

type Server struct {
	a app.Application
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	var request UserCreationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	u, err := s.a.Commands.CreateUser.Handle(ctx, commands.CreateUser{
		Username:  request.Username,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
