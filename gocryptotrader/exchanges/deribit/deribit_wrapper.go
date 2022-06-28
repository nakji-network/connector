package deribit

import (
	"github.com/thrasher-corp/gocryptotrader/exchanges/trade"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/account"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/kline"
	"github.com/thrasher-corp/gocryptotrader/exchanges/order"
	"github.com/thrasher-corp/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-corp/gocryptotrader/exchanges/protocol"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/log"
	"github.com/thrasher-corp/gocryptotrader/portfolio/withdraw"
)

// GetDefaultConfig returns a default exchange config
func (de *Deribit) GetDefaultConfig() (*config.ExchangeConfig, error) {
	de.SetDefaults()
	exchCfg := new(config.ExchangeConfig)
	exchCfg.Name = de.Name
	exchCfg.HTTPTimeout = exchange.DefaultHTTPTimeout
	exchCfg.BaseCurrencies = de.BaseCurrencies

	de.SetupDefaults(exchCfg)

	if de.Features.Supports.RESTCapabilities.AutoPairUpdates {
		err := de.UpdateTradablePairs(true)
		if err != nil {
			return nil, err
		}
	}
	return exchCfg, nil
}

// SetDefaults sets the basic defaults for Deribit
func (de *Deribit) SetDefaults() {
	de.Name = "Deribit"
	de.Enabled = true
	de.Verbose = true
	de.API.CredentialsValidator.RequiresKey = true
	de.API.CredentialsValidator.RequiresSecret = true

	// If using only one pair format for request and configuration, across all
	// supported asset types either SPOT and FUTURES etc. You can use the
	// example below:

	// Request format denotes what the pair as a string will be, when you send
	// a request to an exchange.
	requestFmt := &currency.PairFormat{ /*Set pair request formatting details here for e.g.*/ Uppercase: true, Delimiter: ":"}
	// Config format denotes what the pair as a string will be, when saved to
	// the config.json file.
	configFmt := &currency.PairFormat{ /*Set pair request formatting details here*/ }
	err := de.SetGlobalPairsManager(requestFmt, configFmt /*multiple assets can be set here using the asset package ie asset.Spot*/)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}

	// If assets require multiple differences in formating for request and
	// configuration, another exchange method can be be used e.g. futures
	// contracts require a dash as a delimiter rather than an underscore. You
	// can use this example below:

	futsfmt := currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true, Delimiter: "-"},
		ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: "-"},
	}

	perpfmt := currency.PairStore{
		RequestFormat: &currency.PairFormat{Uppercase: true, Delimiter: "-"},
		ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: "-"},
	}

	//optfmt := currency.PairStore{
	//RequestFormat: &currency.PairFormat{Uppercase: true},
	//ConfigFormat:  &currency.PairFormat{Uppercase: true, Delimiter: ":"},
	//}

	err = de.StoreAssetPairFormat(asset.Futures, futsfmt)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	err = de.StoreAssetPairFormat(asset.PerpetualSwap, perpfmt)
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	//err = de.StoreAssetPairFormat(asset.Option, fmt2)
	//if err != nil {
	//log.Errorln(log.ExchangeSys, err)
	//}

	// Fill out the capabilities/features that the exchange supports
	de.Features = exchange.Features{
		Supports: exchange.FeaturesSupported{
			REST:      true,
			Websocket: true,
			RESTCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WebsocketCapabilities: protocol.Features{
				TickerFetching:    true,
				OrderbookFetching: true,
			},
			WithdrawPermissions: exchange.AutoWithdrawCrypto |
				exchange.AutoWithdrawFiat,
		},
		Enabled: exchange.FeaturesEnabled{
			AutoPairUpdates: true,
		},
	}
	// NOTE: SET THE EXCHANGES RATE LIMIT HERE
	de.Requester = request.New(de.Name,
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	de.API.Endpoints = de.NewEndpoints()
	err = de.API.Endpoints.SetDefaultEndpoints(map[exchange.URL]string{
		exchange.RestSpot: deribitAPIURL,
		exchange.WebsocketSpot: deribitWSURL,
	})
	if err != nil {
		log.Errorln(log.ExchangeSys, err)
	}
	de.Websocket = stream.New()
	de.WebsocketResponseMaxLimit = exchange.DefaultWebsocketResponseMaxLimit
	de.WebsocketResponseCheckTimeout = exchange.DefaultWebsocketResponseCheckTimeout
	de.WebsocketOrderbookBufferLimit = exchange.DefaultWebsocketOrderbookBufferLimit
}

// Setup takes in the supplied exchange configuration details and sets params
func (de *Deribit) Setup(exch *config.ExchangeConfig) error {
	if !exch.Enabled {
		de.SetEnabled(false)
		return nil
	}

	err := de.SetupDefaults(exch)
	if err != nil {
		return err
	}

	wsRunningURL, err := de.API.Endpoints.GetURL(exchange.WebsocketSpot)
	if err != nil {
		return err
	}

	// If websocket is supported, please fill out the following
	err = de.Websocket.Setup(&stream.WebsocketSetup{
		Enabled:                          exch.Features.Enabled.Websocket,
		Verbose:                          exch.Verbose,
		AuthenticatedWebsocketAPISupport: exch.API.AuthenticatedWebsocketSupport,
		WebsocketTimeout:                 exch.WebsocketTrafficTimeout,
		DefaultURL:                       deribitWSURL,
		ExchangeName:                     exch.Name,
		RunningURL:                       wsRunningURL,
		Connector:                        de.WsConnect,
		Subscriber:                       de.Subscribe,
		UnSubscriber:                     de.Unsubscribe,
		Features:                         &de.Features.Supports.WebsocketCapabilities,
	})
	if err != nil {
		return err
	}
	return de.Websocket.SetupNewConnection(stream.ConnectionSetup{
		ResponseCheckTimeout: exch.WebsocketResponseCheckTimeout,
		ResponseMaxLimit:     exch.WebsocketResponseMaxLimit,
	})
}

// Start starts the Deribit go routine
func (de *Deribit) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		de.Run()
		wg.Done()
	}()
}

// Run implements the Deribit wrapper
func (de *Deribit) Run() {
	if de.Verbose {
		log.Debugf(log.ExchangeSys,
			"%s Websocket: %s.",
			de.Name,
			common.IsEnabled(de.Websocket.IsEnabled()))
		de.PrintEnabledPairs()
	}

	if !de.GetEnabledFeatures().AutoPairUpdates {
		return
	}

	err := de.UpdateTradablePairs(false)
	if err != nil {
		log.Errorf(log.ExchangeSys,
			"%s failed to update tradable pairs. Err: %s",
			de.Name,
			err)
	}
}

// FetchTradablePairs returns a list of the exchanges tradable pairs
func (de *Deribit) FetchTradablePairs(asset asset.Item) ([]string, error) {
	// Implement fetching the exchange available pairs if supported
	return nil, nil
}

// UpdateTradablePairs updates the exchanges available pairs and stores
// them in the exchanges config
func (de *Deribit) UpdateTradablePairs(forceUpdate bool) error {
	pairs, err := de.FetchTradablePairs(asset.Spot)
	if err != nil {
		return err
	}

	p, err := currency.NewPairsFromStrings(pairs)
	if err != nil {
		return err
	}

	return de.UpdatePairs(p, asset.Spot, false, forceUpdate)
}

// UpdateTicker updates and returns the ticker for a currency pair
func (de *Deribit) UpdateTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	// NOTE: EXAMPLE FOR GETTING TICKER PRICE
	/*
		tickerPrice := new(ticker.Price)
		tick, err := de.GetTicker(p.String())
		if err != nil {
			return tickerPrice, err
		}
		tickerPrice = &ticker.Price{
			High:  tick.High,
			Low:   tick.Low,
			Bid:   tick.Bid,
			Ask:   tick.Ask,
			Open:  tick.Open,
			Close: tick.Close,
			Pair:  p,
		}
		err = ticker.ProcessTicker(de.Name, tickerPrice, assetType)
		if err != nil {
			return tickerPrice, err
		}
	*/
	return ticker.GetTicker(de.Name, p, assetType)
}

// FetchTicker returns the ticker for a currency pair
func (de *Deribit) FetchTicker(p currency.Pair, assetType asset.Item) (*ticker.Price, error) {
	tickerNew, err := ticker.GetTicker(de.Name, p, assetType)
	if err != nil {
		return de.UpdateTicker(p, assetType)
	}
	return tickerNew, nil
}

// FetchOrderbook returns orderbook base on the currency pair
func (de *Deribit) FetchOrderbook(currency currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	ob, err := orderbook.Get(de.Name, currency, assetType)
	if err != nil {
		return de.UpdateOrderbook(currency, assetType)
	}
	return ob, nil
}

// UpdateOrderbook updates and returns the orderbook for a currency pair
func (de *Deribit) UpdateOrderbook(p currency.Pair, assetType asset.Item) (*orderbook.Base, error) {
	orderBook := &orderbook.Base{
		Exchange:        de.Name,
		Pair:            p,
		Asset:           assetType,
		VerifyOrderbook: de.CanVerifyOrderbook,
	}
	// NOTE: UPDATE ORDERBOOK EXAMPLE
	/*
		orderbookNew, err := de.GetOrderBook(exchange.FormatExchangeCurrency(de.Name, p).String(), 1000)
		if err != nil {
			return orderBook, err
		}

		for x := range orderbookNew.Bids {
			orderBook.Bids = append(orderBook.Bids, orderbook.Item{
				Amount: orderbookNew.Bids[x].Quantity,
				Price: orderbookNew.Bids[x].Price,
			})
		}

		for x := range orderbookNew.Asks {
			orderBook.Asks = append(orderBook.Asks, orderbook.Item{
				Amount: orderBook.Asks[x].Quantity,
				Price: orderBook.Asks[x].Price,
			})
		}
	*/

	err := orderBook.Process()
	if err != nil {
		return orderBook, err
	}

	return orderbook.Get(de.Name, p, assetType)
}

// UpdateAccountInfo retrieves balances for all enabled currencies
func (de *Deribit) UpdateAccountInfo(assetType asset.Item) (account.Holdings, error) {
	return account.Holdings{}, common.ErrNotYetImplemented
}

// FetchAccountInfo retrieves balances for all enabled currencies
func (de *Deribit) FetchAccountInfo(assetType asset.Item) (account.Holdings, error) {
	return account.Holdings{}, common.ErrNotYetImplemented
}

// GetFundingHistory returns funding history, deposits and
// withdrawals
func (de *Deribit) GetFundingHistory() ([]exchange.FundHistory, error) {
	return nil, common.ErrNotYetImplemented
}

//// GetExchangeHistory returns historic trade data within the timeframe provided.
//func (de *Deribit) GetExchangeHistory(p currency.Pair, assetType asset.Item, timestampStart, timestampEnd time.Time) ([]exchange.TradeHistory, error) {
//	return nil, common.ErrNotYetImplemented
//}

// SubmitOrder submits a new order
func (de *Deribit) SubmitOrder(s *order.Submit) (order.SubmitResponse, error) {
	var submitOrderResponse order.SubmitResponse
	if err := s.Validate(); err != nil {
		return submitOrderResponse, err
	}
	return submitOrderResponse, common.ErrNotYetImplemented
}

// ModifyOrder will allow of changing orderbook placement and limit to
// market conversion
func (de *Deribit) ModifyOrder(action *order.Modify) (string, error) {
	return "", common.ErrNotYetImplemented
}

// CancelOrder cancels an order by its corresponding ID number
func (de *Deribit) CancelOrder(order *order.Cancel) error {
	return common.ErrNotYetImplemented
}

// CancelAllOrders cancels all orders associated with a currency pair
func (de *Deribit) CancelAllOrders(orderCancellation *order.Cancel) (order.CancelAllResponse, error) {
	return order.CancelAllResponse{}, common.ErrNotYetImplemented
}

// GetOrderInfo returns information on a current open order
func (de *Deribit) GetOrderInfo(orderID string, pair currency.Pair, assetType asset.Item) (order.Detail, error) {
	return order.Detail{}, common.ErrNotYetImplemented
}

// GetDepositAddress returns a deposit address for a specified currency
func (de *Deribit) GetDepositAddress(cryptocurrency currency.Code, accountID string) (string, error) {
	return "", common.ErrNotYetImplemented
}

// WithdrawCryptocurrencyFunds returns a withdrawal ID when a withdrawal is
// submitted
func (de *Deribit) WithdrawCryptocurrencyFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// WithdrawFiatFunds returns a withdrawal ID when a withdrawal is
// submitted
func (de *Deribit) WithdrawFiatFunds(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// WithdrawFiatFundsToInternationalBank returns a withdrawal ID when a withdrawal is
// submitted
func (de *Deribit) WithdrawFiatFundsToInternationalBank(withdrawRequest *withdraw.Request) (*withdraw.ExchangeResponse, error) {
	return nil, common.ErrNotYetImplemented
}

// GetActiveOrders retrieves any orders that are active/open
func (de *Deribit) GetActiveOrders(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	return nil, common.ErrNotYetImplemented
}

// GetOrderHistory retrieves account order information
// Can Limit response to specific order status
func (de *Deribit) GetOrderHistory(getOrdersRequest *order.GetOrdersRequest) ([]order.Detail, error) {
	return nil, common.ErrNotYetImplemented
}

// GetFeeByType returns an estimate of fee based on the type of transaction
func (de *Deribit) GetFeeByType(feeBuilder *exchange.FeeBuilder) (float64, error) {
	return 0, common.ErrNotYetImplemented
}

// ValidateCredentials validates current credentials used for wrapper
func (de *Deribit) ValidateCredentials(assetType asset.Item) error {
	_, err := de.UpdateAccountInfo(assetType)
	return de.CheckTransientError(err)
}

// GetHistoricCandles returns candles between a time period for a set time interval
func (de *Deribit) GetHistoricCandles(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return kline.Item{}, common.ErrNotYetImplemented
}

// GetHistoricCandlesExtended returns candles between a time period for a set time interval
func (de *Deribit) GetHistoricCandlesExtended(pair currency.Pair, a asset.Item, start, end time.Time, interval kline.Interval) (kline.Item, error) {
	return kline.Item{}, common.ErrNotYetImplemented
}

func (de *Deribit) GetRecentTrades(p currency.Pair, a asset.Item) ([]trade.Data, error) {
	panic("implement me")
}

func (de *Deribit) GetHistoricTrades(p currency.Pair, a asset.Item, startTime, endTime time.Time) ([]trade.Data, error) {
	panic("implement me")
}

func (de *Deribit) CancelBatchOrders(o []order.Cancel) (order.CancelBatchResponse, error) {
	panic("implement me")
}

func (de *Deribit) GetWithdrawalsHistory(code currency.Code) ([]exchange.WithdrawalHistory, error) {
	panic("implement me")
}
