package dbHelper

import (
	"example/ecommerce/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"time"
)

func CreateUser(tx *sqlx.Tx, user models.Users) (uuid.UUID, error) {
	SQL := `INSERT INTO users
            (name, email, password, number, address, zip_code, country)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING user_id`
	var userId uuid.UUID
	err := tx.QueryRowx(SQL, user.Name, user.Email, user.Password, user.PhoneNumber, user.Address, user.ZipCode, user.Country).Scan(&userId)
	return userId, err
}

func CreateUserRole(tx *sqlx.Tx, db *sqlx.DB, userId uuid.UUID, role string) error {
	SQL := `INSERT INTO user_role
            (role_user, user_id)
            VALUES ($1, $2)`
	if db == nil {
		_, err := tx.Exec(SQL, role, userId)
		return err
	} else {
		_, err := db.Exec(SQL, role, userId)
		return err
	}
}

func CheckEmail(db *sqlx.DB, email string) (bool, error) {
	SQL := `SELECT count(*) > 0 FROM users
            WHERE email = $1`

	var isEmailExists bool
	err := db.Get(&isEmailExists, SQL, email)
	return isEmailExists, err
}

func GetIdPassword(db *sqlx.DB, email string) (uuid.UUID, string, error) {
	SQL := `SELECT user_id, password
            FROM users
            WHERE email = $1`
	var emailDatabase string
	var userId uuid.UUID
	err := db.QueryRowx(SQL, email).Scan(&userId, &emailDatabase)
	return userId, emailDatabase, err
}

func CreateItem(db *sqlx.DB, item models.Item) error {
	SQL := `INSERT INTO items
            (name, description, features, price, type, stock_no)
            VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(SQL, item.Name, item.Description, item.Features, item.Price, item.Type, item.StockNo)
	return err
}

func CreateCart(tx *sqlx.Tx, userId uuid.UUID) error {
	SQL := `INSERT INTO carts
            (user_id)
            VALUES($1)`
	_, err := tx.Exec(SQL, userId)
	return err
}

func GetCartId(db *sqlx.DB, userId uuid.UUID) (uuid.UUID, error) {
	SQL := `SELECT cart_id
            FROM carts
            WHERE user_id = $1`
	var cartId uuid.UUID
	err := db.QueryRowx(SQL, userId).Scan(&cartId)
	return cartId, err
}

func AddToCart(db *sqlx.DB, cartItem models.CartItem) error {
	SQL := `INSERT INTO cart_items
			(cart_id, item_id, quantity)
			VALUES ($1, $2, $3)`
	_, err := db.Exec(SQL, cartItem.CartId, cartItem.ItemId, 1)
	return err
}

func GetQuantityCartItem(db *sqlx.DB, cartId, itemId uuid.UUID) (int, error) {
	SQL := `SELECT quantity
            FROM cart_items
            WHERE cart_id = $1 AND item_id = $2`
	var quantity int
	err := db.QueryRowx(SQL, cartId, itemId).Scan(&quantity)
	return quantity, err
}

func IncreaseQuantityCartItem(db *sqlx.DB, cartId, itemId uuid.UUID, quantity int) error {
	SQL := `UPDATE cart_items
            SET quantity = $1
            WHERE cart_id = $2 AND item_id = $3`
	_, err := db.Exec(SQL, quantity+1, cartId, itemId)
	return err
}

func DecreaseQuantityCartItem(db *sqlx.DB, cartId, itemId uuid.UUID, quantity int) error {
	SQL := `UPDATE cart_items
            SET quantity = $1
            WHERE cart_id = $2 AND item_id = $3`
	_, err := db.Exec(SQL, quantity-1, cartId, itemId)
	return err
}

func DeleteFromCart(db *sqlx.DB, cartId, itemId uuid.UUID) error {
	SQL := `DELETE FROM cart_items
           WHERE cart_id = $1 AND item_id = $2`
	_, err := db.Exec(SQL, cartId, itemId)
	return err
}

func GetAllCartItems(db *sqlx.DB, cartId uuid.UUID) ([]models.CartItemDisplay, error) {
	SQL := `SELECT i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no, c.quantity, array_agg(u.url) AS photos
            FROM items i LEFT JOIN cart_items c
            USING(item_id)
            LEFT JOIN images p 
            USING(item_id)
            LEFT JOIN uploads u 
            USING(upload_id)
            WHERE c.cart_id = $1
            GROUP BY i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no, c.quantity`
	var items []models.CartItemDisplay
	err := db.Select(&items, SQL, cartId)
	return items, err
}

func GetAllItemsByType(db *sqlx.DB, typee string) ([]models.Item, error) {
	SQL := `SELECT i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no, array_agg(u.url) AS photos
			FROM items i LEFT JOIN images p
			USING(item_id)
			LEFT JOIN uploads u 
			USING(upload_id)
			WHERE i.type = $1 AND i.archived IS NULL
			GROUP BY i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no`
	var items []models.Item
	err := db.Select(&items, SQL, typee)
	return items, err
}

func GetUserRoles(db *sqlx.DB, userID uuid.UUID) ([]string, error) {
	SQL := `SELECT role_user
            FROM user_role
            WHERE user_id = $1`
	var roles []string
	err := db.Select(&roles, SQL, userID)
	return roles, err
}

func DeleteUser(tx *sqlx.Tx, userId uuid.UUID) error {
	SQL := `UPDATE users
			SET archived = $1
			WHERE user_id = $2`
	_, err := tx.Exec(SQL, time.Now(), userId)
	return err
}

func DeleteUserRole(tx *sqlx.Tx, userId uuid.UUID) error {
	SQL := `DELETE FROM user_role
            WHERE user_id = $1`
	_, err := tx.Exec(SQL, userId)
	return err
}

func DeleteItem(tx *sqlx.Tx, itemId uuid.UUID) error {
	SQL := `UPDATE items
            SET archived = $1
            WHERE item_id = $2`
	_, err := tx.Exec(SQL, time.Now(), itemId)
	return err
}

func DeleteItemFromAllCarts(tx *sqlx.Tx, itemId uuid.UUID) error {
	SQL := `DELETE FROM cart_items
            WHERE item_id = $1`
	_, err := tx.Exec(SQL, itemId)
	return err
}

func DeleteFromImage(tx *sqlx.Tx, itemId uuid.UUID) error {
	SQL := `DELETE FROM images
            WHERE item_id = $1`
	_, err := tx.Exec(SQL, itemId)
	return err
}

func InsertUpload(db *sqlx.DB, upload models.Upload) (uuid.UUID, error) {
	SQL := `INSERT INTO uploads
            (path, name, url)
            VALUES ($1, $2, $3) 
            RETURNING upload_id`
	var uploadId uuid.UUID
	err := db.QueryRowx(SQL, upload.Path, upload.Name, upload.Url).Scan(&uploadId)
	return uploadId, err
}

func InsertImage(db *sqlx.DB, image models.Image) error {
	SQL := `INSERT INTO images
            (item_id, upload_id)
            VALUES ($1, $2)`
	_, err := db.Exec(SQL, image.ItemId, image.UploadId)
	return err
}

func GetAllUsers(db *sqlx.DB, limit, offset int) ([]models.Users, error) {
	SQL := `SELECT u.user_id, u.name, u.email, u.password, u.number AS phone_number, u.address, u.zip_code, u.country, r.role_user
            FROM users u INNER JOIN user_role r 
            USING(user_id)
            LIMIT $1
            OFFSET $2`
	var users []models.Users
	err := db.Select(&users, SQL, limit, offset)
	return users, err
}

func GetItemById(db *sqlx.DB, itemId uuid.UUID) (models.Item, error) {
	SQL := `SELECT i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no, array_agg(u.url) AS photos
			FROM items i LEFT JOIN images p 
			USING(item_id)
			LEFT JOIN uploads u
			USING(upload_id)
			WHERE i.item_id = $1 AND i.archived IS NULL
			GROUP BY i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no`
	var item models.Item
	err := db.Get(&item, SQL, itemId)
	return item, err
}

func GetAllItems(db *sqlx.DB, name string, limit, offset int) ([]models.Item, error) {
	search := "%" + name + "%"
	SQL := `SELECT i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no, array_agg(u.url) AS photos
			FROM items i LEFT JOIN images p 
			USING(item_id)
			LEFT JOIN uploads u
			USING(upload_id)
			WHERE i.archived IS NULL AND i.name ILIKE $1
			GROUP BY i.item_id, i.name, i.description, i.features, i.price, i.type, i.stock_no
			LIMIT $2
			OFFSET $3`
	var items []models.Item
	err := db.Select(&items, SQL, search, limit, offset)
	return items, err
}

func GetUsersCount(db *sqlx.DB) (int, error) {
	SQL := `SELECT COUNT(*)
            FROM users`
	var count int
	err := db.Get(&count, SQL)
	return count, err
}

func GetItemsCount(db *sqlx.DB) (int, error) {
	SQL := `SELECT COUNT(*)
            FROM items`
	var count int
	err := db.Get(&count, SQL)
	return count, err
}

func GetEmailNumber(db *sqlx.DB, userId uuid.UUID) (string, string, error) {
	SQL := `SELECT email, number
			FROM users
			WHERE user_id = $1`
	var (
		email  string
		number string
	)
	err := db.QueryRowx(SQL, userId).Scan(&email, &number)
	return email, number, err
}

func GetOtpNumber(db *sqlx.DB, email string, phone string) (string, error) {
	SQL := `SELECT otp_number
			FROM otp
			WHERE email = $1 AND phone = $2 AND expires_at > NOW()
			ORDER BY created_at DESC
			LIMIT 1`
	var otpNumber string
	err := db.Get(&otpNumber, SQL, email, phone)
	return otpNumber, err
}

func InsertInOtp(tx *sqlx.Tx, email string, phoneNumber string, otpNumber string) error {
	SQL := `INSERT INTO otp
            (email, phone, otp_number, created_at, expires_at)
            VALUES ($1, $2, $3, $4, $5)`
	currentTime := time.Now()
	fiveMinsLater := currentTime.Add(15 * time.Minute)
	_, err := tx.Exec(SQL, email, phoneNumber, otpNumber, currentTime, fiveMinsLater)
	return err
}

func UpdateVerification(db *sqlx.DB, userId uuid.UUID) error {
	SQL := `UPDATE users
            SET is_verified = $1
            WHERE user_id = $2`
	_, err := db.Exec(SQL, time.Now(), userId)
	return err
}

func IsVerified(db *sqlx.DB, userId uuid.UUID) (bool, error) {
	SQL := `SELECT COUNT(*) > 0
            FROM users
            WHERE user_id = $1 AND is_verified IS NOT NULL `
	var isVerified bool
	err := db.Get(&isVerified, SQL, userId)
	return isVerified, err
}
