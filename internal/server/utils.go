package server

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/baby-platom/loyalty-system/internal/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func jsonContentTypeMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			h.ServeHTTP(w, r)
		},
	)
}

func checkIfOneStrcutFieldIsEmpty(s any) string {
	v := reflect.ValueOf(s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field, typeOfField := v.Field(i), t.Field(i)
		if field.IsZero() {
			return fmt.Sprintf("field %s is empty", typeOfField.Name)
		}
	}
	return ""
}

func defaultReactionToInternalServerError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error(err)
	http.Error(w, "", http.StatusInternalServerError)
}

func defaultReactionToEncodingResponseError(w http.ResponseWriter, logger *zap.SugaredLogger, err error) {
	logger.Error("Error encoding response", zap.Error(err))
	http.Error(w, "Error encoding response", http.StatusInternalServerError)
}

func fillUserByLogin(w http.ResponseWriter, logger *zap.SugaredLogger, user *database.User, userLogin string) {
	userFilter := database.User{Login: userLogin}
	res := database.DB.Where(&userFilter).First(&user)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		defaultReactionToInternalServerError(w, logger, res.Error)
		return
	}
}

// Luhn validates the provided data using the Luhn algorithm.
func Luhn(s []byte) bool {
	n := len(s)
	number := 0
	result := 0
	for i := 0; i < n; i++ {
		number = int(s[i]) - '0'
		if i%2 != 0 {
			result += number
			continue
		}
		number *= 2
		if number > 9 {
			number -= 9
		}
		result += number
	}
	return result%10 == 0
}
