package server

import (
	"example/ecommerce/handler"
	"example/ecommerce/middlewares"
	"github.com/go-chi/chi"
)

func SetUpRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", handler.RegisterHandler)
	r.Post("/login", handler.LoginHandler)
	r.Get("/all-items", handler.GetAllItemsHandler)
	r.Get("/all-items-type", handler.GetAllItemsByTypeHandler)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.JWTMiddleware)

		r.Route("/user", func(r chi.Router) {
			r.Use(middlewares.MiddlewareUser)
			r.Get("/", handler.GetUserDetailsHandler)
			r.Get("/send-email", handler.SendVerificationEmailHandler)
			r.Get("/send-sms", handler.SendVerificationSmsHandler)
			r.Get("/verify-otp", handler.VerifyOtpHandler)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.VerificationMiddleware)
				r.Delete("/delete-account", handler.DeleteAccountHandler)
				r.Route("/items", func(r chi.Router) {
					r.Get("/{itemId}", handler.GetItemByIdHandler)
					r.Post("/{itemId}/add-to-cart", handler.AddToCartHandler)
				})
				r.Route("/cart", func(r chi.Router) {
					r.Get("/{cartId}/all-items", handler.GetAllCartItemsHandler)
					r.Delete("/{cartId}/items/{itemId}", handler.DeleteFromCartHandler)
					r.Post("/checkout", handler.CheckoutHandler)
				})
			})
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(middlewares.MiddlewareAdmin)

			r.Get("/all-users", handler.GetAllUsersHandler)
			r.Route("/users", func(r chi.Router) {
				r.Get("/{userId}", handler.GetUserDetailsHandler)
				r.Delete("/{userId}", handler.DeleteUserByAdminHandler)
				r.Put("/{userId}/add-role", handler.AddUserRoleHandler)
			})
			r.Post("/upload-image", handler.UploadImageHandler)
			r.Route("/items", func(r chi.Router) {
				r.Post("/add-item", handler.CreateItemHandler)
				r.Post("/{itemId}/add-image", handler.InsertImageHandler)
				r.Delete("/{itemId}", handler.DeleteItemHandler)
			})
		})
	})

	return r
}
