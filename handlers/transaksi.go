package handlers

import (
	dto "Backend/dto/result"
	transaksidto "Backend/dto/transaksi"
	"Backend/models"
	"Backend/repositories"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
	"gopkg.in/gomail.v2"
)

var path_file = "http://localhost:5000/uploads/"

var c = coreapi.Client{
	ServerKey: os.Getenv("SERVER_KEY"),
	ClientKey: os.Getenv("CLIENT_KEY"),
}

type handleTransaksi struct {
	TransaksiRepository repositories.TransaksiRepository
}

func HandlerTransaksi(TransaksiRepository repositories.TransaksiRepository) *handleTransaksi {
	return &handleTransaksi{TransaksiRepository}
}

func (h *handleTransaksi) FindTransaksi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	transaksi, err := h.TransaksiRepository.FindTransaksi()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}
	for i, p := range transaksi {
		transaksi[i].Image = path_file + p.Image
	}

	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResult{Code: http.StatusOK, Data: transaksi}
	json.NewEncoder(w).Encode(response)

}

func (h *handleTransaksi) GetTransaksi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	transaksi, err := h.TransaksiRepository.GetTransaksi(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResult{Code: http.StatusOK, Data: transaksi}
	json.NewEncoder(w).Encode(response)
}

func (h *handleTransaksi) AddTransaksi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// get data user token
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	userId := int(userInfo["id"].(float64))

	// dataContex := r.Context().Value("dataFile") // add this code
	// filename := dataContex.(string)             // add this code

	counterqty, _ := strconv.Atoi(r.FormValue("counter_qty"))
	total, _ := strconv.Atoi(r.FormValue("total"))
	tripid, _ := strconv.Atoi(r.FormValue("trip_id"))
	request := transaksidto.CreateTransaksiRequest{
		CounterQTY: counterqty,
		Total:      total,
		// Image:      filename,
		Trip_id: tripid,
		UserID:  userId,
	}
	validation := validator.New()
	err := validation.Struct(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := dto.ErrorResult{Code: http.StatusBadRequest, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	var TransIdIsMatch = false
	var TransactionId int
	for !TransIdIsMatch {
		TransactionId = request.Trip_id + rand.Intn(10000) - rand.Intn(100)
		transactionData, _ := h.TransaksiRepository.GetTransaksi(TransactionId)
		if transactionData.ID == 0 {
			TransIdIsMatch = true
		}
	}

	transaksi := models.Transaction{
		ID:         TransactionId,
		CounterQTY: request.CounterQTY,
		Total:      request.Total,
		Status:     "pending",
		Trip_id:    request.Trip_id,
		UserID:     request.UserID,
	}

	data, err := h.TransaksiRepository.AddTransaksi(transaksi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := dto.ErrorResult{Code: http.StatusInternalServerError, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
	}

	test, err := h.TransaksiRepository.GetTransaksi(data.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := dto.ErrorResult{Code: http.StatusInternalServerError, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
	}

	var s = snap.Client{}
	s.New("SB-Mid-server-6I3E0inWOb5vmbkn7BZULp4Q", midtrans.Sandbox)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  strconv.Itoa(test.ID),
			GrossAmt: int64(test.Total),
		},
		CreditCard: &snap.CreditCardDetails{
			Secure: true,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: test.User.Fullname,
			Email: test.User.Email,
		},
	}

	snapResp, _ := s.CreateTransaction(req)
	fmt.Println(snapResp)

	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResult{Code: http.StatusOK, Data: snapResp}
	json.NewEncoder(w).Encode(response)
}

func SendEmail(status string, transaksi models.Transaction) {

	var CONFIG_SMTP_HOST = "smtp.gmail.com"
	var CONFIG_SMTP_PORT = 587
	var CONFIG_SENDER_NAME = "NetokidsMerch <dionovalino@gmail.com>"
	var CONFIG_AUTH_EMAIL = os.Getenv("SYSTEM_EMAIL")
	var CONFIG_AUTH_PASSWORD = os.Getenv("SYSTEM_PASSWORD")

	var TripName = transaksi.User.Fullname
	var price = strconv.Itoa(transaksi.Total)

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", CONFIG_SENDER_NAME)
	mailer.SetHeader("To", transaksi.User.Email)
	mailer.SetHeader("Subject", "Status Transaksi")
	mailer.SetBody("text/html", fmt.Sprintf(`<!DOCTYPE html>
    <html lang="en">
      <head>
      <meta charset="UTF-8" />
      <meta http-equiv="X-UA-Compatible" content="IE=edge" />
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <title>Document</title>
      <style>
        h1 {
        color: brown;
        }
      </style>
      </head>
      <body>
      <h2>Product payment :</h2>
      <ul style="list-style-type:none;">
        <li>Name : %s</li>
        <li>Total Harga: Rp.%s</li>
        <li>Status : <b>%s</b></li>
        <li>Iklan : <b>%s</b></li>
      </ul>
      </body>
    </html>`, TripName, price, status, "Gausah Beli Lagi"))

	dialer := gomail.NewDialer(
		CONFIG_SMTP_HOST,
		CONFIG_SMTP_PORT,
		CONFIG_AUTH_EMAIL,
		CONFIG_AUTH_PASSWORD,
	)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	err := dialer.DialAndSend(mailer)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (h *handleTransaksi) Notification(w http.ResponseWriter, r *http.Request) {
	var notificationPayload map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&notificationPayload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := dto.ErrorResult{Code: http.StatusBadRequest, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	transactionStatus := notificationPayload["transaction_status"].(string)
	fraudStatus := notificationPayload["fraud_status"].(string)
	orderId := notificationPayload["order_id"].(string)

	orderIdInt, _ := strconv.Atoi(orderId)

	// Get One Transaction from repository GetOneTransaction using orderId parameter here ...
	transaction, _ := h.TransaksiRepository.GetTransaksi(orderIdInt)

	if transactionStatus == "capture" {
		if fraudStatus == "challenge" {
			SendEmail("GAGAL", transaction)
			h.TransaksiRepository.UpdateTransaksi("pending", transaction.ID)
		} else if fraudStatus == "accept" {
			SendEmail("OKE", transaction)
			h.TransaksiRepository.UpdateTransaksi("success", transaction.ID)
		}
	} else if transactionStatus == "settlement" {
		SendEmail("OKE", transaction)
		h.TransaksiRepository.UpdateTransaksi("success", transaction.ID)
	} else if transactionStatus == "deny" {
		SendEmail("GAGAL", transaction)
		h.TransaksiRepository.UpdateTransaksi("failed", transaction.ID)
	} else if transactionStatus == "cancel" || transactionStatus == "expire" {
		SendEmail("GAGAL", transaction)
		h.TransaksiRepository.UpdateTransaksi("failed", transaction.ID)
	} else if transactionStatus == "pending" {
		SendEmail("TEST", transaction)
		h.TransaksiRepository.UpdateTransaksi("pending", transaction.ID)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *handleTransaksi) DeleteTransaksi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	transaksi, err := h.TransaksiRepository.GetTransaksi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := dto.ErrorResult{Code: http.StatusBadRequest, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	data, err := h.TransaksiRepository.DeleteTransaksi(transaksi)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := dto.ErrorResult{Code: http.StatusInternalServerError, Message: err.Error()}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.WriteHeader(http.StatusOK)
	response := dto.SuccessResult{Code: http.StatusOK, Data: convertResponseTransaksi(data)}
	json.NewEncoder(w).Encode(response)

}

func convertResponseTransaksi(u models.Transaction) transaksidto.TransaksiResponse {
	return transaksidto.TransaksiResponse{
		ID: u.ID,
	}
}
