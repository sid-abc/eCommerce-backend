package handler

import (
	"database/sql"
	"encoding/json"
	"example/ecommerce/database"
	"example/ecommerce/database/dbHelper"
	"example/ecommerce/models"
	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
	"os"
	"time"
)

// user
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.Users
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if user.Password == "" || user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	isEmailExists, err := dbHelper.CheckEmail(database.Todo, user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if isEmailExists {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	user.Password = string(hash)

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId, err := dbHelper.CreateUser(tx, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = dbHelper.CreateUserRole(tx, userId, models.Role2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = dbHelper.CreateCart(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.Users
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, passwordDatabase, err := dbHelper.GetPassword(database.Todo, user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(passwordDatabase), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims := &models.Claims{
		UserID: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"token": tokenString,
	}
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// admin
func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = dbHelper.CreateItem(database.Todo, item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// user
func AddToCartHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserID
	cartId, err := dbHelper.GetCartId(database.Todo, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	itemIdString := chi.URLParam(r, "itemId")
	itemId, err := uuid.Parse(itemIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	quantity, err := dbHelper.GetQuantityCartItem(database.Todo, cartId, itemId)

	if err != nil && err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if quantity == 0 {
		cartItem := models.CartItem{CartId: cartId, ItemId: itemId}
		err2 := dbHelper.AddToCart(database.Todo, cartItem)
		if err2 != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		return
	}

	err = dbHelper.IncreaseQuantityCartItem(database.Todo, cartId, itemId, quantity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// user
func DeleteFromCartHandler(w http.ResponseWriter, r *http.Request) {
	cartIdString := chi.URLParam(r, "cartId")
	cartId, err := uuid.Parse(cartIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	itemIdString := chi.URLParam(r, "itemId")
	itemId, err := uuid.Parse(itemIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	quantity, err := dbHelper.GetQuantityCartItem(database.Todo, cartId, itemId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if quantity == 1 {
		err = dbHelper.DeleteFromCart(database.Todo, cartId, itemId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}

	if quantity == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = dbHelper.DecreaseQuantityCartItem(database.Todo, cartId, itemId, quantity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// user
func GetAllCartItemsHandler(w http.ResponseWriter, r *http.Request) {
	cartIdString := chi.URLParam(r, "cartId")
	cartId, err := uuid.Parse(cartIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cartItems, err := dbHelper.GetAllCartItems(database.Todo, cartId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cartItems)
}

func GetAllItemsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	itemType := r.URL.Query().Get("itemType")
	if itemType != models.Type1 && itemType != models.Type2 && itemType != models.Type3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	items, err := dbHelper.GetAllItemsByType(database.Todo, itemType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

func GetItemByIdHandler(w http.ResponseWriter, r *http.Request) {
	itemIdString := chi.URLParam(r, "itemId")
	itemId, err := uuid.Parse(itemIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	item, err := dbHelper.GetItemById(database.Todo, itemId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}

// admin
func DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	itemIdString := chi.URLParam(r, "itemId")
	itemId, err := uuid.Parse(itemIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dbHelper.DeleteItem(tx, itemId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}
	err = dbHelper.DeleteItemFromAllCarts(tx, itemId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = dbHelper.DeleteFromImage(tx, itemId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// admin
func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	upload := models.Upload{}
	upload.Name = handler.Filename + time.Now().String()
	upload.Path = "./uploads/" + upload.Name
	upload.Url = "https://dfstudio-d420.kxcdn.com/wordpress/wp-content/uploads/2019/06/digital_camera_photo-1080x675.jpg"
	f, err := os.OpenFile(upload.Path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	uploadId, err := dbHelper.InsertUpload(database.Todo, upload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := map[string]string{
		"uploadId": uploadId.String(),
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// admin
func InsertImageHandler(w http.ResponseWriter, r *http.Request) {
	var upload models.Upload
	err := json.NewDecoder(r.Body).Decode(&upload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	itemIdString := chi.URLParam(r, "itemId")
	itemId, err := uuid.Parse(itemIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	image := models.Image{ItemId: itemId, UploadId: upload.UploadId}
	err = dbHelper.InsertImage(database.Todo, image)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// admin
func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := dbHelper.GetAllUsers(database.Todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// admin
func DeleteUserByAdminHandler(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dbHelper.DeleteUser(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = dbHelper.DeleteUserRole(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// user
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserID

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dbHelper.DeleteUser(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = dbHelper.DeleteUserRole(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func GetAllItemsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	items, err := dbHelper.GetAllItems(database.Todo, name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}
