package api

import (
	"net/http"

	"github.com/lopesmarcello/gobid/internal/jsonutils"
)

func (api *Api) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Session.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, jsonutils.JsonMsg("message", "must be logged in"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
