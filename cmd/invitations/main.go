package main

import (
	"net/http"

	"github.com/koffee/invitations/cmd/invitations/internals"
	"github.com/koffee/invitations/pkg/config"
)

func main() {
	// rabbit := rabbitmq.Initialize()
	db := config.InitConfig()
	repo := internals.Initialize(db)
	// repo.CreateProfile(2, "gabivlj")
	r := internals.InitController(repo)
	if http.ListenAndServe(":4333", r) != nil {
		panic("Error, probably port already assigned")
	}
}
