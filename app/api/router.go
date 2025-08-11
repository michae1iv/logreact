package api

import (
	point "correlator/api/endpoints"
	"correlator/api/session"
	"correlator/config"
	"correlator/logger"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func InitRouter(cfg config.ServerConfig, secret string) error {
	session.JwtSecret = []byte(secret)
	router := chi.NewRouter()
	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins, // Frontend
		AllowedMethods:   []string{"GET", "POST", "DELETE", "PUT"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,                           // for getting and setting tokens
		MaxAge:           int((1 * time.Hour).Seconds()), // cash
	}))

	// auth api
	router.Post("/api/auth/login", point.Login)
	router.Get("/api/auth/logout", point.Logout)

	// rules api
	router.Route("/api/rules", func(router chi.Router) {
		router.Use(session.JwtMiddleware("rule"))
		router.Get("/", point.ShowRules)
		router.Get("/{id}", point.ShowSingleRule)
		router.Patch("/{id}", point.EditRule)
		router.Delete("/{id}", point.DeleteRule)
		router.Post("/create", point.CreateRule)
	})

	// lists api
	router.Route("/api/lists", func(router chi.Router) {
		router.Use(session.JwtMiddleware("list"))
		router.Get("/", point.ShowLists)
		router.Get("/{id}", point.GetList)
		router.Patch("/{id}", point.UpdateList)
		router.Delete("/{id}", point.DeleteList)
		router.Post("/create", point.CreateList)
	})

	// stastics api
	router.Route("/api/stats", func(router chi.Router) {
		router.Use(session.JwtMiddleware("list"))
		router.Get("/alerts", point.GetAlerts)
	})

	router.Mount("/api/admin", adminRouter())

	logger.InfoLogger.Printf("Server started on port %v\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), router)

	return nil
}

// A completely separate router for administrator routes
func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(adminOnly)
	r.Get("/stats", point.AdminStats)
	//r.Get("/actions", point.LatestActions)
	// api for working with users
	r.Get("/users", point.GetAllUsers)
	r.Get("/users/{id}", point.GetUser)
	r.Patch("/users/{id}", point.UpdateUser)
	r.Delete("/users/{id}", point.DeleteUser)
	r.Post("/users/create", point.CreateUser)
	// api for working with groups
	r.Get("/groups", point.GetAllGroups)
	r.Get("/groups/{id}", point.GetGroup)
	r.Patch("/groups/{id}", point.UpdateGroup)
	r.Delete("/groups/{id}", point.DeleteGroup)
	r.Post("/groups/create", point.CreateGroup)

	return r
}

func adminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := r.Cookie("auth")
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// Token check
		claims, err := session.ValidateJWT(tokenString.Value)
		if err != nil || claims.ExpiresAt < time.Now().Unix() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		// checking if user is admin
		if !claims.IsAdmin {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
