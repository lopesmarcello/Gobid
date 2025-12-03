package api

import (
	"errors"
	"net/http"

	"github.com/lopesmarcello/gobid/internal/jsonutils"
	"github.com/lopesmarcello/gobid/internal/services"
	"github.com/lopesmarcello/gobid/internal/usecase/user"
)

func (api *API) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJSON[user.CreateUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJSON(w, r, http.StatusUnprocessableEntity, problems)
	}

	id, err := api.UserService.CreateUser(
		r.Context(),
		data.UserName,
		data.Password,
		data.Bio,
		data.Email,
	)
	if err != nil {
		if errors.Is(err, services.ErrDuplicatedEmailOrUsername) {
			_ = jsonutils.EncodeJSON(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": "duplicated email or invalid password",
			})
			return
		}
	}

	_ = jsonutils.EncodeJSON(w, r, http.StatusCreated, map[string]any{
		"user_id": id,
	})
}

func (api *API) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJSON[user.LoginUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJSON(w, r, http.StatusUnprocessableEntity, problems)
	}

	id, err := api.UserService.AuthenticateUser(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			jsonutils.EncodeJSON(w, r, http.StatusBadRequest, jsonutils.JSONmsg("error", "invalid email or password"))
			return
		}
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("error", "unexpected internal server error while authenticating user via user service"))
		return
	}

	err = api.Session.RenewToken(r.Context())
	if err != nil {
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("error", "unexpected internal server error while renewing token"))
		return
	}

	api.Session.Put(r.Context(), "AuthenticatedUserId", id)

	jsonutils.EncodeJSON(w, r, http.StatusOK, jsonutils.JSONmsg("message", "logged in, succesfully"))
}

func (api *API) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	err := api.Session.RenewToken(r.Context())
	if err != nil {
		jsonutils.EncodeJSON(w, r, http.StatusInternalServerError, jsonutils.JSONmsg("error", "unexpected internal server error"))
		return
	}

	api.Session.Remove(r.Context(), "AuthenticatedUserId")
	jsonutils.EncodeJSON(w, r, http.StatusOK, jsonutils.JSONmsg("message", "logged out succesfully"))
}
