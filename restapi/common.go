package restapi

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return nil
}

func writeError(w http.ResponseWriter, err error, httpStatus int) {
	http.Error(w, err.Error(), httpStatus)
}
