package model

import "time"

type Transaction struct {
	OutTradeNo  string    `json:"out_trade_no"`
	OpenId      string    `json:"openid"`
	UncleOpenId string    `json:"uncle_openid"`
	PrepayId    string    `json:"prepay_id"`
	PriceInFen  int       `json:"price_in_fen"`
	Times       int       `json:"times"`
	Description string    `json:"description"`
	TradeState  string    `json:"trade_state"`
	Redeemed    bool      `json:"redeemed"`
	Payload     string    `json:"payload"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
