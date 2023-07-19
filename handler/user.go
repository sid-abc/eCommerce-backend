package handler

import (
	"database/sql"
	"encoding/json"
	"example/ecommerce/database"
	"example/ecommerce/database/dbHelper"
	"example/ecommerce/models"
	"example/ecommerce/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"strconv"
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

	validate := validator.New()
	err = validate.Struct(user)
	if err != nil {
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
		tx.Rollback()
		return
	}

	err = dbHelper.CreateUserRole(tx, nil, userId, models.Role2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = dbHelper.CreateCart(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
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
	var user models.UserLogin
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, passwordDatabase, err := dbHelper.GetIdPassword(database.Todo, user.Email)
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

	validate := validator.New()
	err = validate.Struct(item)
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
		w.WriteHeader(http.StatusBadRequest)
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

	awsConfig := utils.AWSConfig{
		AccessKeyID:     os.Getenv("AccessKeyID"),
		AccessKeySecret: os.Getenv("AccessKeySecret"),
		Region:          os.Getenv("Region"),
		BucketName:      os.Getenv("BucketName"),
	}

	// creating aws session
	sess := utils.CreateSession(awsConfig)

	upload := models.Upload{}
	upload.Name = handler.Filename + time.Now().String()

	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsConfig.BucketName),
		Key:    aws.String(upload.Name),
		Body:   file,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// creating s3 session
	svc := s3.New(sess)
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(awsConfig.BucketName),
		Key:    aws.String(upload.Name),
	})
	url, err := req.Presign(20 * time.Minute)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	upload.Url = url
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
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("size"))

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNumber - 1) * pageSize
	users, err := dbHelper.GetAllUsers(database.Todo, pageSize, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userCount, err := dbHelper.GetUsersCount(database.Todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":      users,
		"totalItems": userCount,
	})
}

// admin
func DeleteUserByAdminHandler(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
		tx.Rollback()
		return
	}

	err = dbHelper.DeleteUserRole(tx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
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
	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("size"))

	if pageNumber <= 0 {
		pageNumber = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (pageNumber - 1) * pageSize

	items, err := dbHelper.GetAllItems(database.Todo, name, pageSize, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	itemCount, err := dbHelper.GetItemsCount(database.Todo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items":      items,
		"totalItems": itemCount,
	})
}

func AddUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	userIdString := chi.URLParam(r, "userId")
	userId, err := uuid.Parse(userIdString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userRole := r.URL.Query().Get("role")
	roles, err := dbHelper.GetUserRoles(database.Todo, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, x := range roles {
		if x == userRole {
			w.WriteHeader(http.StatusConflict)
			return
		}
	}

	err = dbHelper.CreateUserRole(nil, database.Todo, userId, userRole)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Checkout Success",
	})
}
