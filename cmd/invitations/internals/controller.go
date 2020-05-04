package internals

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/koffee/invitations/pkg/auth"
	"github.com/koffee/invitations/pkg/middleware"
	"github.com/koffee/invitations/pkg/rabbitmq"
	"github.com/streadway/amqp"
)

var listener = rabbitmq.Initialize()

var senderOfNewCollaborator = listener.NewSender("update_collaborators")
var msgListener = listener

type album struct {
	ID            uint32   `json:"id"`
	AlbumID       uint32   `json:"album_id"`
	UserID        uint32   `json:"user_id"`
	Collaborators []string `json:"collaborators"`
	Name          string   `json:"name"`

	EmailCreator string   `json:"email_creator"`
	Collaborator string   `json:"collaborator"`
	Artists      []string `json:"artists"`
}

type profile struct {
	UserID uint32 `json:"userid"`
	Name   string `json:"username"`
}

type handler struct {
	db Repository

	rabbit rabbitmq.MessageListener
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
	router.POST("/api/invitations", middleware.Authenticate(h.postInvitation, paseto))
	router.PUT("/api/invitations", middleware.Authenticate(h.acceptInvitation, paseto))

	msgListener.OnMessage("new_profile", func(msg *amqp.Delivery) {
		var data profile
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			log.Println(err)
		}
		log.Println(data)
		profile := h.db.CreateProfile(data.UserID, data.Name)
		log.Println("New profile", profile)
	})

	msgListener.OnMessage("new_album", func(msg *amqp.Delivery) {
		var data album
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			log.Println(err)
		}
		log.Println(data)
		if data.Artists == nil || len(data.Artists) == 0 {
			data.Artists = []string{}
		}
		album := h.db.CreateAlbum(data.ID, data.Artists, data.EmailCreator, data.Name)
		log.Println("New album", album)
	})
	return router
}

func (h *handler) getInvitations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := middleware.UserID(r)
	invitations := h.db.GetInvitations(id)
	res := toResponse(map[string]interface{}{"invitations": invitations}, 200, "Success!")
	w.WriteHeader(200)
	w.Write(res)
}

func (h *handler) postInvitation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var album album
	_, email := middleware.UserID(r)
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		w.WriteHeader(400)
		w.Write(toResponse(nil, 400, "Error parsing body"))
		return
	}
	err := h.db.NewInvitation(album.UserID, album.AlbumID, email)
	if err != nil {
		w.WriteHeader(400)
		w.Write(toResponse(map[string]string{"error": err.Error()}, 400, "Error creating invitation"))
		return
	}
	w.WriteHeader(200)
	w.Write(toResponse(map[string]bool{"invitation": true}, 200, "Success"))
}

// Works
func (h *handler) acceptInvitation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id, _ := middleware.UserID(r)
	var album album
	if err := json.NewDecoder(r.Body).Decode(&album); err != nil {
		w.WriteHeader(400)
		w.Write(toResponse(nil, 400, "Error parsing body"))
		return
	}
	newArtists, err := h.db.AcceptInvitation(id, album.AlbumID)
	if err != nil {
		w.WriteHeader(400)
		w.Write(toResponse(err, 400, "Error accepting invitation"))
		return
	}
	w.WriteHeader(200)
	w.Write(toResponse(map[string]interface{}{"new_artists": newArtists, "album_id": album.AlbumID}, 200, "Success"))
	bytes, _ := json.Marshal(map[string]interface{}{"artists": newArtists, "id": album.AlbumID})
	senderOfNewCollaborator.Send(bytes, "application/json")
}
