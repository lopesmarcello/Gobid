package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/lopesmarcello/gobid/internal/jsonutils"
	"github.com/lopesmarcello/gobid/internal/services"
	"github.com/lopesmarcello/gobid/internal/usecase/product"
)

func (api *API) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJSON[product.CreateProductReq](r)
	if err != nil {
		jsonutils.EncodeJSON(w, r, http.StatusUnprocessableEntity, problems)
		return
	}

	userID, ok := api.Session.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)
	if !ok {

		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("error", "unexpected internal server error"))
		return
	}

	productID, err := api.ProductService.CreateProduct(r.Context(), userID, data.ProductName, data.Description, data.Baseprice, data.AuctionEnd)
	if err != nil {
		fmt.Println("erro:")
		fmt.Println(err)
		jsonutils.EncodeJSON(w, r, http.StatusUnprocessableEntity, jsonutils.JSONmsg("error", "failed to create product auction"))
		return
	}

	ctx, _ := context.WithDeadline(context.Background(), data.AuctionEnd)

	auctionRoom := services.NewAuctionRoom(ctx, productID, api.BidsService)

	go auctionRoom.Run()

	api.AuctionLobby.Lock()
	api.AuctionLobby.Rooms[productID] = auctionRoom
	api.AuctionLobby.Unlock()

	jsonutils.EncodeJSON(w, r, http.StatusCreated, jsonutils.JSONmsg("product_id", productID))
}
