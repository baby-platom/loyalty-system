package local_accrual

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func OrderInfoAPIHandler(w http.ResponseWriter, r *http.Request) {
	order := mux.Vars(r)["number"]
	status := getRandomHttpStatus()
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}

	orderInfo := getRandomOrderInfo(order)
	dataEncoded, _ := json.Marshal(orderInfo)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dataEncoded)
}
