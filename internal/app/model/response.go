package model

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type OrderData struct {
	Name	string `json:"name"`
	Recipe	string `json:"recipe"`
	PhoneNumber string `json:"phone_number"`
	Timestamp   string `json:"timestamp"`
}
