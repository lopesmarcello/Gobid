package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lopesmarcello/gobid/internal/jsonutils"
	"github.com/lopesmarcello/gobid/internal/services"
)

func (api *API) HandleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Subscribing user to auction")
	rawProductID := chi.URLParam(r, "product_id")

	productID, err := uuid.Parse(rawProductID)
	if err != nil {
		jsonutils.EncodeJSON(w, r, http.StatusBadRequest, jsonutils.JSONmsg("message:", "invalid product id - must be a valid uuid"))
		return
	}

	_, err = api.ProductService.GetProductByID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJSON(w, r, http.StatusNotFound, jsonutils.JSONmsg("error:", "product not found"))
			return
		}
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("message:", "unexpected server error"))
		return
	}

	userID, ok := api.Session.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("message:", "failed to recover session ID"))
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productID]
	api.AuctionLobby.Unlock()

	if !ok {
		jsonutils.EncodeJSON(w, r, http.StatusBadRequest, jsonutils.JSONmsg("message:", "the auction has ended"))
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("message:", "Could not upgrade connection to a websocket protocol"))
		return
	}

	client := services.NewClient(room, conn, userID)

	room.Register <- client
	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
