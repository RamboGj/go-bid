package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID)
	api.Router.Use(middleware.Recoverer)
	api.Router.Use(middleware.Logger)
	api.Router.Use(api.Sessions.LoadAndSave)

	// csrfMiddleware := csrf.Protect(
	// 	[]byte(os.Getenv("GOBID_CSRF_KEY")),
	// 	csrf.Secure(false), // DEV ONLY
	// )

	// DEV ONLY: mark requests as plaintext HTTP so gorilla/csrf skips the
	// TLS-only Referer/Origin allow-listing checks. Without this it assumes
	// TLS and rejects requests with no Referer ("referer not supplied").
	// api.Router.Use(func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		next.ServeHTTP(w, csrf.PlaintextHTTPRequest(r))
	// 	})
	// })

	// api.Router.Use(csrfMiddleware)

	api.Router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			// r.Get("/csrftoken", api.HandleGetCSRFToken)
			r.Route("/users/", func(r chi.Router) {
				r.Post("/signup", api.handleSignupUser)
				r.Post("/login", api.handleLoginUser)

				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)
					r.Post("/", api.handleLogoutUser)
				})
			})

			r.Route("/products", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(api.AuthMiddleware)

					r.Post("/", api.handleCreateProduct)

					r.Get("/ws/subscribe/{product_id}", api.handleSubscribeUserToAuction)
				})
			})
		})

	})
}
