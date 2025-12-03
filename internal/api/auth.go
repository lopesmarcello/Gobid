package api

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/lopesmarcello/gobid/internal/jsonutils"
)

func (api *API) HandleGetCSRFtoken(w http.ResponseWriter, r *http.Request) {
	token := csrf.Token(r)
	jsonutils.EncodeJSON(w, r, http.StatusOK, jsonutils.JSONmsg("csrf_token", token))
}

func (api *API) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !api.Session.Exists(r.Context(), "AuthenticatedUserId") {
			jsonutils.EncodeJSON(w, r, http.StatusUnauthorized, jsonutils.JSONmsg("message", "must be logged in"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
