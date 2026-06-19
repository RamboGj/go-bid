package api

import (
	"go-bid/internal/jsonutils"
	"go-bid/internal/usecase/product"
	"net/http"

	"github.com/google/uuid"
)

func (api *Api) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[product.CreateProductRequest](r)

	if err != nil {
		if problems != nil {
			jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"error": "invalid or malformed request body",
		})
		return
	}

	userID, ok := api.Sessions.Get(r.Context(), "AuthenticatedUserId").(uuid.UUID)

	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "unexpected server error",
		})
		return
	}

	id, err := api.ProductService.CreateProduct(
		r.Context(),
		userID,
		data.ProductName,
		data.Description,
		data.BasePriceCents,
		data.AuctionEnd,
	)

	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"error": "failed to create product auction",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"message":    "product successfully created",
		"product_id": id,
	})
}
