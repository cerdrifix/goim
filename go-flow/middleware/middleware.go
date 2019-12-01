package middleware

import (
	"net/http"
)

/** The middleware function **/
func authorize(h http.Handler) http.Handler {
	checkAuth := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		err := validateAuth(r.URL.Query().Get("auth"))
		if err != nil {
			http.Error(w, "bad authorization parameter", http.StatusUnauthorized)
		}

		h.ServeHTTP(w, r)
	})
	return checkAuth
}

func validateAuth(key string) error {
	return nil
}
