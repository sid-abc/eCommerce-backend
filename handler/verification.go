package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"example/ecommerce/database"
	"example/ecommerce/database/dbHelper"
	"example/ecommerce/models"
	"example/ecommerce/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/rest/api/v2010"
	"math/big"

	"net/http"

	"os"
)

const (
	otpLength = 6
)

func GenerateOTP() (string, error) {
	charSet := "0123456789"

	otpBytes := make([]byte, otpLength)

	for i := 0; i < otpLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
		if err != nil {
			return "", err
		}
		otpBytes[i] = charSet[randomIndex.Int64()]
	}

	otp := string(otpBytes)

	return otp, nil
}

func SendVerificationEmail(otpNumber, email string) error {
	sender := "siddhant.nigam@outlook.com"
	recipient := email
	subject := "Email Verification"
	htmlBody := "<h1>Amazon SES Test Email (AWS SDK for Go)</h1><p>This email was sent with " +
		"<a href='https://aws.amazon.com/ses/'>Amazon SES</a> using the " +
		"<a href='https://aws.amazon.com/sdk-for-go/'>AWS SDK for Go</a>. Your OTP is " + otpNumber + "</p>"
	textBody := "eghfnehf;hghlnf"
	charSet := "UTF-8"

	awsConfig := utils.AWSConfig{
		AccessKeyID:     os.Getenv("AccessKeyID"),
		AccessKeySecret: os.Getenv("AccessKeySecret"),
		Region:          os.Getenv("Region"),
	}

	// creating aws session
	sess := utils.CreateSession(awsConfig)

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(htmlBody),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(textBody),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	_, err := svc.SendEmail(input)
	return err
}

func SendVerificationEmailHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserID

	isVerified, err := dbHelper.IsVerified(database.Todo, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isVerified {
		w.WriteHeader(http.StatusConflict)
		return
	}

	otpNumber, err := GenerateOTP()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	email, _, err := dbHelper.GetEmailNumber(database.Todo, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dbHelper.MakeOtpInvalid(tx, userId, models.OtpType1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = dbHelper.InsertInOtp(tx, userId, models.OtpType1, otpNumber)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = SendVerificationEmail(otpNumber, email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tx.Commit()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SendVerificationSmsHandler(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserID

	isVerified, err := dbHelper.IsVerified(database.Todo, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isVerified {
		w.WriteHeader(http.StatusConflict)
		return
	}

	otpNumber, err := GenerateOTP()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, number, err := dbHelper.GetEmailNumber(database.Todo, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	client := twilio.NewRestClient()
	params := &openapi.CreateMessageParams{}
	params.SetTo("+91" + number)
	params.SetFrom(os.Getenv("TWILIO_PHONE_NUMBER"))
	params.SetBody("Your OTP is " + otpNumber)

	tx, err := database.Todo.Beginx()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = dbHelper.MakeOtpInvalid(tx, userId, models.OtpType2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	err = dbHelper.InsertInOtp(tx, userId, models.OtpType2, otpNumber)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		tx.Rollback()
		return
	}

	_, err = client.Api.CreateMessage(params)
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

func VerifyOtpHandler(w http.ResponseWriter, r *http.Request) {
	var userOtp models.Otp
	err := json.NewDecoder(r.Body).Decode(&userOtp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userOtp.Type != models.OtpType1 && userOtp.Type != models.OtpType2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := r.Context().Value("claims").(*models.Claims)
	userId := claims.UserID

	databaseOtp, err := dbHelper.GetOtpNumber(database.Todo, userId, userOtp.Type)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if databaseOtp != userOtp.OtpNumber {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if userOtp.Type == models.OtpType1 {
		err = dbHelper.UpdateEmailVerification(database.Todo, userId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		err = dbHelper.UpdatePhoneVerification(database.Todo, userId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
