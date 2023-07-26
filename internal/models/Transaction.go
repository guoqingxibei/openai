package models

type Transaction struct {
	OutTradeNo  string `json:"out_trade_no"`
	OpenId      string `json:"openid"`
	UncleOpenId string `json:"uncle_openid"`
	PrepayId    string `json:"prepay_id"`
	PriceInFen  int    `json:"price_in_fen"`
	Times       int    `json:"times"`
	Description string `json:"description"`
	TradeState  string `json:"trade_state"`
	Redeemed    bool   `json:"redeemed"`
	Payload     string `json:"payload"`
	CreatedTime int64  `json:"created_time"`
	UpdatedTime int64  `json:"updated_time"`
}
