package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/koffee/invitations/pkg/auth"
)

// Authenticate saves the request
func Authenticate(next httprouter.Handle, tokenService auth.Token) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		usr, err := tokenService.VerifyToken(r.Header.Get("Authorization"))
		if err != nil {
			b, _ := json.Marshal(map[string]interface{}{"status": 400, "message": "Unauthorized", "data": err})
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(b)
			return
		}
		email, _, id, _ := usr.Information()
		ctx := context.WithValue(context.Background(), "user", map[string]interface{}{"email": email, "id": id})
		rcloned := r.Clone(ctx)
		next(w, rcloned, p)
	}
}

// UserID returns the loged user
func UserID(r *http.Request) (uint32, string) {
	value := r.Context().Value("user")
	val := value.(map[string]interface{})
	return val["id"].(uint32), val["email"].(string)
}
