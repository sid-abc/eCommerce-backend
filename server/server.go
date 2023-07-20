package server

import (
	"example/ecommerce/handler"
	"example/ecommerce/middlewares"
	"github.com/go-chi/chi"
)

func SetUpRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", handler.RegisterHandler)
	r.Get("/login", handler.LoginHandler)
	r.Get("/all-items", handler.GetAllItemsHandler)
	r.Get("/all-items-type", handler.GetAllItemsByTypeHandler)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware)

		r.Route("/user", func(r chi.Router) {
			r.Use(middlewares.MiddlewareUser)
			r.Delete("/delete-account", handler.DeleteAccountHandler)
			r.Put("/verify-email", handler.SendVerificationEmailHandler)
			r.Put("/verify-sms", handler.SendVerificationSmsHandler)
			r.Put("/verify-otp", handler.VerifyOtpHandler)
			r.Route("/items", func(r chi.Router) {
				r.Get("/{itemId}", handler.GetItemByIdHandler)
				r.Post("/{itemId}/add-to-cart", handler.AddToCartHandler)
			})
			r.Route("/cart", func(r chi.Router) {
				r.Get("/{cartId}/all-items", handler.GetAllCartItemsHandler)
				r.Delete("/{cartId}/items/{itemId}", handler.DeleteFromCartHandler)
				r.Group(func(r chi.Router) {
					r.Use(middlewares.VerificationMiddleware)
					r.Put("/checkout", handler.CheckoutHandler)
				})
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middlewares.MiddlewareAdmin)

			r.Get("/all-users", handler.GetAllUsersHandler)
			r.Route("/users", func(r chi.Router) {
				r.Delete("/{userId}/delete-user", handler.DeleteUserByAdminHandler)
				r.Put("/{userId}/add-role", handler.AddUserRoleHandler)
			})
			r.Post("/upload-image", handler.UploadImageHandler)
			r.Route("/items", func(r chi.Router) {
				r.Post("/add-item", handler.CreateItemHandler)
				r.Post("/{itemId}/add-image", handler.InsertImageHandler)
				r.Delete("/{itemId}/delete-item", handler.DeleteItemHandler)
			})
		})
	})

	return r
}
