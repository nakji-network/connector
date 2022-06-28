package deribit

import "encoding/json"

type Kind string

const (
	Future Kind = "future"
	Option      = "option"
	NA          = ""
)

type errorCapture struct {
	Status    string `json:"status"`
	Code      int64  `json:"err_code"`
	ErrMsg    string `json:"err_msg"`
	Timestamp int64  `json:"ts"`
}

// remove 'kind' for options and futures and perp etc
// /public/get_instruments
//{"method":"public/get_instruments","params":{"currency":"BTC","kind":"future","expired":false},"jsonrpc":"2.0","id":0}
//{"method":"public/get_instruments","params":{"currency":"ETH","kind":"future","expired":false},"jsonrpc":"2.0","id":0}
type Instruments struct {
	Jsonrpc string       `json:"jsonrpc"`
	ID      int          `json:"id"`
	Result  []Instrument `json:"result"`
	//UsIn    int64 `json:"usIn"`
	//UsOut   int64 `json:"usOut"`
	//UsDiff  int   `json:"usDiff"`
	//Testnet bool  `json:"testnet"`
}

type Instrument struct {
	TickSize             float64 `json:"tick_size"`
	TakerCommission      float64 `json:"taker_commission"`
	Strike               float64 `json:"strike"`
	SettlementPeriod     string  `json:"settlement_period"`
	QuoteCurrency        string  `json:"quote_currency"`
	OptionType           string  `json:"option_type"`
	MinTradeAmount       float64 `json:"min_trade_amount"`
	MakerCommission      float64 `json:"maker_commission"`
	Kind                 string  `json:"kind"`
	IsActive             bool    `json:"is_active"`
	InstrumentName       string  `json:"instrument_name"`
	ExpirationTimestamp  int64   `json:"expiration_timestamp"`
	CreationTimestamp    int64   `json:"creation_timestamp"`
	ContractSize         float64 `json:"contract_size"`
	BlockTradeCommission float64 `json:"block_trade_commission"`
	BaseCurrency         string  `json:"base_currency"`
}

type SubscribeWrapper struct {
	Method string `json:"method"`
	Params struct {
		Channel string          `json:"channel"`
		Data    json.RawMessage `json:"data"`
	} `json:"params"`
	Error struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// /public/ticker //get ticker for an instrument
// subscribe ticker.{instrument_name}.{interval} Key info about the instrument
type Ticker struct {
	UnderlyingPrice float64 `json:"underlying_price"`
	UnderlyingIndex string  `json:"underlying_index"`
	Timestamp       int64   `json:"timestamp"`
	Stats           struct {
		Volume      float64 `json:"volume"`
		PriceChange float64 `json:"price_change"`
		Low         float64 `json:"low"`
		High        float64 `json:"high"`
	} `json:"stats"`
	State           string  `json:"state"`
	SettlementPrice float64 `json:"settlement_price"`
	OpenInterest    float64 `json:"open_interest"`
	MinPrice        float64 `json:"min_price"`
	MaxPrice        float64 `json:"max_price"`
	MarkPrice       float64 `json:"mark_price"`
	MarkIv          float64 `json:"mark_iv"`
	LastPrice       float64 `json:"last_price"`
	InterestRate    float64 `json:"interest_rate"`
	InstrumentName  string  `json:"instrument_name"`
	IndexPrice      float64 `json:"index_price"`
	Greeks          struct {
		Vega  float64 `json:"vega"`
		Theta float64 `json:"theta"`
		Rho   float64 `json:"rho"`
		Gamma float64 `json:"gamma"`
		Delta float64 `json:"delta"`
	} `json:"greeks"`
	EstimatedDeliveryPrice float64 `json:"estimated_delivery_price"`
	BidIv                  float64 `json:"bid_iv"`
	BestBidPrice           float64 `json:"best_bid_price"`
	BestBidAmount          float64 `json:"best_bid_amount"`
	BestAskPrice           float64 `json:"best_ask_price"`
	BestAskAmount          float64 `json:"best_ask_amount"`
	AskIv                  float64 `json:"ask_iv"`
}

// WsSub stores subscription data
type WsSub struct {
	Method string `json:"method"`
	Params struct {
		Channels []string `json:"channels"`
	} `json:"params"`
}
