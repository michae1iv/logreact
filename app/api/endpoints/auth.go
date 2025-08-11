package endpoints

import (
	"correlator/api/session"
	"correlator/db"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Function which handles LoginPage
func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var response map[string]string

	var credentials map[string]string
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil || credentials["username"] == "" || credentials["password"] == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	user := db.User{}

	tx := db.DB.Begin()
	if tx.Error != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	res := tx.Where(&db.User{Username: credentials["username"]}).First(&user)
	if res.Error != nil {
		tx.Rollback()
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if !user.CheckPasswordHash(credentials["password"]) {
		tx.Rollback()
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	tx.Commit() // if success, then complete transaction

	if !user.IsActive {
		response = map[string]string{"status": "failed", "operation": "login", "message": fmt.Sprintf("User %s isn't active", user.Username)}
		json.NewEncoder(w).Encode(response)
		return
	}

	res = db.DB.Preload("Group").Where("id = ?", user.ID).First(&user)
	if res.Error != nil {
		http.Error(w, http.StatusText(http.StatusFailedDependency), http.StatusFailedDependency)
		return
	}

	// Generating and setting JWT
	jwtToken, err := session.GenerateJWT(user.Username, user.Group.ID, user.Group.Permissions["admin"].(bool))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	body := strings.Split(jwtToken, ".")[1]
	exp := time.Now().Add(24 * time.Hour)
	//? This cookie stores only jwt payload,
	//? so frontend could parse it and select params from it
	http.SetCookie(w, &http.Cookie{
		Name:     "sess",
		Value:    body,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  exp,
		HttpOnly: false,
		Secure:   false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    jwtToken,
		SameSite: http.SameSiteLaxMode,
		Path:     "/api/",
		Expires:  exp,
		HttpOnly: true,
		Secure:   false,
		Domain:   "localhost",
	})

	response = map[string]string{"status": "success", "operation": "login", "message": fmt.Sprintf("User %s successfully logged in", user.Username)}
	json.NewEncoder(w).Encode(response)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Delete token from cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/api/",
		MaxAge:   -1,
		Secure:   false,
		Domain:   "localhost",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "sess",
		Value:    "",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   false,
	})
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "success", "operation": "logout", "message": "You've logged out"}
	json.NewEncoder(w).Encode(response)
}
