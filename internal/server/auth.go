package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
	"github.com/baby-platom/loyalty-system/internal/reflect"
	"gorm.io/gorm"
)

type userDataStruct struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func returnAuthCookie(w http.ResponseWriter, id uint) {
	newAuthToken, err := auth.BuildJWTString(id)
	if err == nil {
		cookie := &http.Cookie{
			Name:  "auth",
			Value: newAuthToken,
			Path:  "/",
		}
		http.SetCookie(w, cookie)
	} else {
		logger.Log.Errorw(
			"Error occured while building new JWT auth token",
			"error", err,
		)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func parseUserData(w http.ResponseWriter, r *http.Request) (userDataStruct, bool) {
	var userData userDataStruct
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return userData, false
	}

	if msg := reflect.CheckIfOneStrcutFieldIsEmpty(userData); msg != "" {
		logger.Log.Error(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return userData, false
	}
	return userData, true
}

func UserRegisterAPIHandler(w http.ResponseWriter, r *http.Request) {
	userData, ok := parseUserData(w, r)
	if !ok {
		return
	}

	hash, err := auth.HashPassword(userData.Password)
	if err != nil {
		msg := "password can't be hashed"
		logger.Log.Error(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	var user database.User
	err = database.DB.WithinTransaction(
		r.Context(),
		func(ctx context.Context) error {
			user, err = database.CreateUserWithBalance(r.Context(), userData.Login, hash)
			return err
		},
	)

	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			msg := fmt.Sprintf("user with login '%s' already exists", userData.Login)
			logger.Log.Error(msg)
			http.Error(w, msg, http.StatusNoContent)
			return
		}
		defaultReactionToInternalServerError(w, logger.Log, err)
		return
	}

	logger.Log.Info("Creating new auth cookie")
	returnAuthCookie(w, user.ID)
}

func reactToIncorrectLoginCredentials(w http.ResponseWriter) {
	msg := "incorrect combination of login and password"
	logger.Log.Error(msg)
	http.Error(w, msg, http.StatusUnauthorized)
}

func UserLoginAPIHandler(w http.ResponseWriter, r *http.Request) {
	userData, ok := parseUserData(w, r)
	if !ok {
		return
	}

	user, err := database.GetUserByLogin(r.Context(), userData.Login)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		reactToIncorrectLoginCredentials(w)
		return
	}

	if !auth.CheckPasswordHash(userData.Password, user.PasswordHash) {
		reactToIncorrectLoginCredentials(w)
		return
	}

	logger.Log.Info("Creating new auth cookie")
	returnAuthCookie(w, user.ID)
}
