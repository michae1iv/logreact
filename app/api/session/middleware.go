package session

import (
	"context"
	"correlator/db"
	"net/http"
	"time"
)

type key int

const (
	UserCtxKey key = iota
)

func defineRights(endpoint string, g *db.Group) string {
	switch endpoint {
	case "list":
		if res, ok := g.Permissions["list"].(string); ok {
			return res
		}
	case "rule":
		if res, ok := g.Permissions["rule"].(string); ok {
			return res
		}
	default:
	}
	return "all"
}

func JwtMiddleware(endpoint string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := r.Cookie("auth")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			// Token check
			claims, err := ValidateJWT(tokenString.Value)
			if err != nil || claims.ExpiresAt < time.Now().Unix() {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			var user db.User
			if res := db.DB.Where(&db.User{Username: claims.Username}).First(&user); res.RowsAffected < 1 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			var group db.Group
			if res := db.DB.Where(&db.Group{ID: claims.UserGroup}).First(&group); res.RowsAffected < 1 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			claimsmap := map[string]interface{}{
				"User":    &user,
				"IsAdmin": claims.IsAdmin,
				"Group":   &group,
				"Rights":  defineRights(endpoint, &group),
			}

			ctx := context.WithValue(r.Context(), UserCtxKey, claimsmap)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
