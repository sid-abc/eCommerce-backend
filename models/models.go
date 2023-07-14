package models

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

const (
	Role1 = "admin"
	Role2 = "user"
	Type1 = "headphone"
	Type2 = "speaker"
	Type3 = "earphone"
)

type Users struct {
	UserId      uuid.UUID `db:"user_id"`
	Name        string    `json:"name" db:"name" validate:"required"`
	Email       string    `json:"email" db:"email" validate:"required,email"`
	Password    string    `json:"password" db:"password" validate:"required,min=6"`
	PhoneNumber int       `json:"phoneNumber" db:"phone_number" validate:"required"`
	Address     string    `json:"address" db:"address" validate:"required"`
	ZipCode     int       `json:"zipCode" db:"zip_code" validate:"required"`
	Country     string    `json:"country" db:"country" validate:"required"`
	Archived    time.Time `db:"archived"`
	RoleUser    string    `db:"role_user"`
}

type UserLogin struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type UserRole struct {
	Id       uuid.UUID
	RoleUser string    `json:"roleUser" db:"role_user"`
	UserId   uuid.UUID `json:"userId" db:"user_id"`
}

type Item struct {
	ItemId      uuid.UUID `db:"item_id"`
	Name        string    `json:"name" db:"name" validate:"required"`
	Description string    `json:"description" db:"description" validate:"required"`
	Features    string    `json:"features" db:"features"`
	Price       int       `json:"price" db:"price"`
	Type        string    `json:"type" db:"type" validate:"required"`
	StockNo     int       `json:"stockNo" db:"stock_no"`
	Photos      []byte    `json:"photos" db:"photos"`
}

type Cart struct {
	CartId uuid.UUID
	UserId uuid.UUID `json:"userId" db:"user_id"`
}

type CartItem struct {
	CartId   uuid.UUID `json:"cartId" db:"cart_id"`
	ItemId   uuid.UUID `json:"itemId" db:"item_id"`
	Quantity int       `json:"quantity" db:"quantity"`
}

type CartItemDisplay struct {
	ItemId      uuid.UUID `json:"itemId" db:"item_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Features    string    `json:"features" db:"features"`
	Price       int       `json:"price" db:"price"`
	Type        string    `json:"type" db:"type"`
	StockNo     int       `json:"stockNo" db:"stock_no"`
	Quantity    int       `json:"quantity" db:"quantity"`
	Photos      []byte    `json:"photos" db:"photos"`
}

type Image struct {
	ItemId   uuid.UUID `json:"itemId" db:"item_id"`
	UploadId uuid.UUID `json:"uploadId" db:"upload_id"`
}

type Upload struct {
	UploadId uuid.UUID `json:"uploadId"`
	Path     string    `json:"path" db:"path"`
	Name     string    `json:"name" db:"name"`
	Url      string    `json:"url" db:"url"`
}

type Claims struct {
	UserID uuid.UUID `json:"userID"`
	jwt.StandardClaims
}

//r.HandleFunc("/users/cart/{cartId}/products/{itemId}/", handler.GetAllRestaurantsHandler).Methods("GET")

// "/users/products/{itemId}"
