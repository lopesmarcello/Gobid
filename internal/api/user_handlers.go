package api

import (
	"errors"
	"net/http"

	"github.com/lopesmarcello/gobid/internal/jsonutils"
	"github.com/lopesmarcello/gobid/internal/services"
	"github.com/lopesmarcello/gobid/internal/usecase/user"
)

func (api *Api) handleSignupUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.CreateUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
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
			_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
				"error": "duplicated email or invalid password",
			})
			return
		}
	}

	_ = jsonutils.EncodeJson(w, r, http.StatusCreated, map[string]any{
		"user_id": id,
	})
}

func (api *Api) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	data, problems, err := jsonutils.DecodeValidJson[user.LoginUserReq](r)
	if err != nil {
		_ = jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, problems)
	}

	id, err := api.UserService.AuthenticateUser(r.Context(), data.Email, data.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, jsonutils.JsonMsg("error", "invalid email or password"))
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, jsonutils.JsonMsg("error", "unexpected internal server error while authenticating user via user service"))
		return
	}

	err = api.Session.RenewToken(r.Context())
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, jsonutils.JsonMsg("error", "unexpected internal server error while renewing token"))
		return
	}

	api.Session.Put(r.Context(), "AuthenticatedUserId", id)

	jsonutils.EncodeJson(w, r, http.StatusOK, jsonutils.JsonMsg("message", "logged in, succesfully"))
}

func (api *Api) handleLogoutUser(w http.ResponseWriter, r *http.Request) {
	err := api.Session.RenewToken(r.Context())
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, jsonutils.JsonMsg("error", "unexpected internal server error"))
		return
	}

	api.Session.Remove(r.Context(), "AuthenticatedUserId")
	jsonutils.EncodeJson(w, r, http.StatusOK, jsonutils.JsonMsg("message", "logged out succesfully"))
}
