package internals

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/koffee/invitations/pkg/auth"
	"github.com/koffee/invitations/pkg/middleware"
	"github.com/koffee/invitations/pkg/rabbitmq"
)

var listener = rabbitmq.Initialize()

var senderOfNewCollaborator = listener.NewSender("new_collaborator")
var senderOfDeletionOfCollaborator = listener.NewSender("delete_collaborator")

type album struct {
	ID            uint32   `json:"id"`
	Collaborators []string `json:"collaborators"`

	Collaborator string `json:"collaborator"`
}

type profile struct {
	UserID uint32 `json:"userid"`
	Name   string `json:"name"`
}

type handler struct {
	db Repository
}

func toResponse(i interface{}, status uint32, message string) []byte {
	data := map[string]interface{}{"status": status, "message": message, "data": i}
	res, _ := json.Marshal(data)
	return res
}

// InitController of the app
func InitController(db Repository) *httprouter.Router {
	h := handler{db: db}
	router := httprouter.New()
	paseto := auth.NewPaseto()
	router.GET("/api/invitations", middleware.Authenticate(h.getInvitations, paseto))
	return router
}

func (h *handler) getInvitations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := middleware.UserID(r)
	invitations := h.db.GetInvitations(id)
	res := toResponse(map[string]interface{}{"invitations": invitations}, 200, "Success!")
	w.WriteHeader(200)
	w.Write(res)
}
