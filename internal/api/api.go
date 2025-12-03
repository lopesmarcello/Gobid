// Package api corresponds to everything api-related
package api

import (
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/lopesmarcello/gobid/internal/services"
)

type API struct {
	Router         *chi.Mux
	UserService    services.UserService
	ProductService services.ProductService
	Session        *scs.SessionManager
	WsUpgrader     websocket.Upgrader
	AuctionLobby   services.AuctionLobby
	BidsService    services.BidsService
}
