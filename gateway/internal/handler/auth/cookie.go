package auth

import (
	"net/http"
	"time"

	"github.com/uwu-octane/antBackend/gateway/internal/handler/constvar"
)

func SetAuthCookies(w http.ResponseWriter, sid string, refresh string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     constvar.CookieSidName,
		Value:    sid,
		Path:     constvar.CookiePath,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     constvar.CookieRefreshName,
		Value:    refresh,
		Path:     constvar.CookiePath,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearAuthCookies(w http.ResponseWriter, secure bool) {
	expired := time.Unix(0, 0)
	http.SetCookie(w, &http.Cookie{
		Name:     constvar.CookieSidName,
		Value:    "",
		Path:     constvar.CookiePath,
		Expires:  expired,
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode, //if samesite=none, secure must be true
	})
	http.SetCookie(w, &http.Cookie{
		Name:     constvar.CookieRefreshName,
		Value:    "",
		Path:     constvar.CookiePath,
		Expires:  expired,
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
