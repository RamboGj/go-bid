package api

import (
	"errors"
	"go-bid/internal/jsonutils"
	"go-bid/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (api *Api) handleSubscribeUserToAuction(w http.ResponseWriter, r *http.Request) {
	rawProductId := chi.URLParam(r, "product_id")

	productId, err := uuid.Parse(rawProductId)

	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error": "invalid product id, must be a valid uuid",
		})
		return
	}

	_, err = api.ProductService.GetProductById(r.Context(), productId)

	if err != nil {
		if errors.Is(err, services.ErrProductNotFound) {
			jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
				"message": "product not found",
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "unexpected server internal error",
		})
		return
	}

	userId, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "unexpected server internal error",
		})
		return
	}

	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "could not upgrade connection to a websocket protocol",
		})
		return
	}

	api.AuctionLobby.Lock()
	room, ok := api.AuctionLobby.Rooms[productId]
	api.AuctionLobby.Unlock()

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "the auction has ended",
		})
		return
	}

	conn, err := api.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "could not upgrade connection to a websocket protocol",
		})
		return
	}

	// The read/write goroutines own the connection's lifecycle and each
	// close it in their own defer, so we must NOT close it here — doing so
	// would tear down the socket the instant this handler returns.
	client := services.NewClient(room, conn, userId)

	room.Register <- client
	go client.ReadEventLoop()
	go client.WriteEventLoop()
}
