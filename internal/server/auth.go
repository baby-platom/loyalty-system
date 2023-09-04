package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/baby-platom/loyalty-system/internal/auth"
	"github.com/baby-platom/loyalty-system/internal/database"
	"github.com/baby-platom/loyalty-system/internal/logger"
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

	if msg := checkIfOneStrcutFieldIsEmpty(userData); msg != "" {
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

	user := database.User{Login: userData.Login, PasswordHash: hash}
	res := database.DB.Create(&user)

	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			msg := fmt.Sprintf("user with login '%s' already exists", userData.Login)
			logger.Log.Error(msg)
			http.Error(w, msg, http.StatusNoContent)
			return
		}
		defaultReactionToInternalServerError(w, logger.Log, res.Error)
		return
	}

	balance := database.Balance{UserID: user.ID}
	res = database.DB.Create(&balance)
	if res.Error != nil {
		defaultReactionToInternalServerError(w, logger.Log, res.Error)
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

	var user database.User
	userFilter := database.User{Login: userData.Login}
	res := database.DB.Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
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
