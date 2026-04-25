package token

import (
	"net/http"
)

func Extract(r *http.Request) string {
	cookie, err := r.Cookie("access_token")
	if err != nil || cookie.Value == "" {
		return ""
	}
	return cookie.Value
}
