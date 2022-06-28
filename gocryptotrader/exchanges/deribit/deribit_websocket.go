package deribit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/asset"
	"github.com/thrasher-corp/gocryptotrader/exchanges/stream"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	deribitWSURL = "wss://www.deribit.com/ws/api/v2"

	pingDelay = 1 * time.Minute // TODO: check this

	//"announcements"
	//"book.{instrument_name}.{group}.{depth}.{interval}"
	//"book.{instrument_name}.{interval}"
	//"chart.trades.{instrument_name}.{resolution}"
	//"deribit_price_index.{index_name}"
	//"deribit_price_ranking.{index_name}"
	//"estimated_expiration_price.{index_name}"
	//"markprice.options.{index_name}"
	//"perpetual.{instrument_name}.{interval}"
	//"platform_state"
	//"quote.{instrument_name}"
	//"ticker.{instrument_name}.{interval}"
	//"trades.{instrument_name}.{interval}"
	//"trades.{kind}.{currency}.{interval}"
	//"user.changes.{instrument_name}.{interval}"
	//"user.changes.{kind}.{currency}.{interval}"
	//"user.mmp_trigger.{currency}"
	//"user.orders.{instrument_name}.raw"
	//"user.orders.{instrument_name}.{interval}"
	//"user.orders.{kind}.{currency}.raw"
	//"user.orders.{kind}.{currency}.{interval}"
	//"user.portfolio.{currency}"
	//"user.trades.{instrument_name}.{interval}"
	//"user.trades.{kind}.{currency}.{interval}"
)

type StreamParams struct {
	InstrumentName string
	Group          string
	Depth          string
	Interval       string
	Resolution     string
	IndexName      string
	Kind           string
	Currency       string
}

var streamTemplateStrings map[string]*template.Template = map[string]*template.Template{
	"announcements":              template.Must(template.New("announcements").Parse("announcements")),
	"bookGroup":                  template.Must(template.New("bookGroup").Parse("book.{{.InstrumentName}}.{{.Group}}.{{.Depth}}.{{.Interval}}")),
	"book":                       template.Must(template.New("book").Parse("book.{{.InstrumentName}}.{{.Interval}}")),
	"chart.trades":               template.Must(template.New("chart.trades").Parse("chart.trades.{{.InstrumentName}}.{{.Resolution}}")),
	"deribit_price_index":        template.Must(template.New("deribit_price_index").Parse("deribit_price_index.{{.IndexName}}")),
	"deribit_price_ranking":      template.Must(template.New("deribit_price_ranking").Parse("deribit_price_ranking.{{.IndexName}}")),
	"estimated_expiration_price": template.Must(template.New("estimated_expiration_price").Parse("estimated_expiration_price.{{.IndexName}}")),
	"markprice.options":          template.Must(template.New("markprice.options").Parse("markprice.options.{{.IndexName}}")),
	"perpetual":                  template.Must(template.New("perpetual").Parse("perpetual.{{.InstrumentName}}.{{.Interval}}")),
	"platform_state":             template.Must(template.New("platform_state").Parse("platform_state")),
	"quote":                      template.Must(template.New("quote").Parse("quote.{{.InstrumentName}}")),
	"ticker":                     template.Must(template.New("ticker").Parse("ticker.{{.InstrumentName}}.{{.Interval}}")),
	"trades":                     template.Must(template.New("trades").Parse("trades.{{.InstrumentName}}.{{.Interval}}")),
	"tradesKind":                 template.Must(template.New("tradesKind").Parse("trades.{{.Kind}}.{{.Currency}}.{{.Interval}}")),
	"user.changes":               template.Must(template.New("user.changes").Parse("user.changes.{{.InstrumentName}}.{{.Interval}}")),
	"user.changesKind":           template.Must(template.New("user.changesKind").Parse("user.changes.{{.Kind}}.{{.Currency}}.{{.Interval}}")),
	"user.mmp_trigger":           template.Must(template.New("user.mmp_trigger").Parse("user.mmp_trigger.{{.Currency}}")),
	"user.orders.raw":            template.Must(template.New("user.orders.raw").Parse("user.orders.{{.InstrumentName}}.raw")),
	"user.orders":                template.Must(template.New("user.orders").Parse("user.orders.{{.InstrumentName}}.{{.Interval}}")),
	"user.ordersKind.raw":        template.Must(template.New("user.ordersKind.raw").Parse("user.orders.{{.Kind}}.{{.Currency}}.raw")),
	"user.ordersKind":            template.Must(template.New("user.ordersKind").Parse("user.orders.{{.Kind}}.{{.Currency}}.{{.Interval}}")),
	"user.portfolio":             template.Must(template.New("user.portfolio").Parse("user.portfolio.{{.Currency}}")),
	"user.trades":                template.Must(template.New("user.trades").Parse("user.trades.{{.InstrumentName}}.{{.Interval}}")),
	"user.tradesKind":            template.Must(template.New("user.tradesKind").Parse("user.trades.{{.Kind}}.{{.Currency}}.{{.Interval}}")),
}

func StreamName(name string, params StreamParams) (string, error) {
	var sb strings.Builder

	tmpl := streamTemplateStrings[name]
	err := tmpl.Execute(&sb, params)
	if err != nil {
		return "", err
	}
	streamName := sb.String()
	if strings.Contains(streamName, "..") || strings.HasSuffix(streamName, ".") {
		return "", errors.New("streamname has invalid args")
	}
	return streamName, nil
}

func (de *Deribit) WsConnect() error {
	if !de.Websocket.IsEnabled() || !de.IsEnabled() {
		return errors.New(stream.WebsocketNotEnabled)
	}
	var dialer websocket.Dialer
	err := de.Websocket.Conn.Dial(&dialer, http.Header{})
	if err != nil {
		return err
	}
	de.Websocket.Conn.SetupPingHandler(stream.PingHandler{
		MessageType: websocket.PingMessage,
		Delay:       pingDelay,
	})
	if de.Verbose {
		log.Debugf(log.ExchangeSys, "%s Connected to Websocket.\n", de.Name)
	}

	go de.wsReadData()
	if de.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
		err = de.WsAuth()
		if err != nil {
			de.Websocket.DataHandler <- err
			de.Websocket.SetCanUseAuthenticatedEndpoints(false)
		}
	}

	subs, err := de.GenerateDefaultSubscriptions()
	if err != nil {
		return err
	}
	return de.Websocket.SubscribeToChannels(subs)
}

// WsAuth sends an authentication message to receive auth data
func (de *Deribit) WsAuth() error {
	//intNonce := time.Now().UnixNano() / 1000000
	//strNonce := strconv.FormatInt(intNonce, 10)
	//hmac := crypto.GetHMAC(
	//crypto.HashSHA256,
	//[]byte(strNonce+"websocket_login"),
	//[]byte(de.API.Credentials.Secret),
	//)
	//sign := crypto.HexEncodeToString(hmac)
	//req := Authenticate{Operation: "login",
	//Args: AuthenticationData{
	//Key:  de.API.Credentials.Key,
	//Sign: sign,
	//Time: intNonce,
	//},
	//}
	//return de.Websocket.Conn.SendJSONMessage(req)
	return nil
}

// Subscribe sends a websocket message to receive data from the channel
func (de *Deribit) Subscribe(channelsToSubscribe []stream.ChannelSubscription) error {
	var errs common.Errors
	var sub WsSub
	sub.Method = PublicSubscribe

channels:
	for i := range channelsToSubscribe {
		switch ch := channelsToSubscribe[i].Channel; ch {
		default:
			a, err := de.GetPairAssetType(channelsToSubscribe[i].Currency)
			if err != nil {
				errs = append(errs, err)
				continue channels
			}

			formattedPair, err := de.FormatExchangeCurrency(channelsToSubscribe[i].Currency, a)
			if err != nil {
				errs = append(errs, err)
				continue channels
			}
			streamName, err := StreamName(ch, StreamParams{InstrumentName: formattedPair.String(), Interval: "raw"})
			if err != nil {
				errs = append(errs, err)
				continue channels
			}
			sub.Params.Channels = append(sub.Params.Channels, streamName)
		}
	}
	err := de.Websocket.Conn.SendJSONMessage(sub)
	if err != nil {
		errs = append(errs, err)
	}
	de.Websocket.AddSuccessfulSubscriptions(channelsToSubscribe...)
	return nil
}

// Unsubscribe sends a websocket message to stop receiving data from the channel
func (de *Deribit) Unsubscribe(channelsToUnsubscribe []stream.ChannelSubscription) error {
	var errs common.Errors
	var sub WsSub
	sub.Method = PublicUnsubscribe

channels:
	for i := range channelsToUnsubscribe {
		switch ch := channelsToUnsubscribe[i].Channel; ch {
		default:
			a, err := de.GetPairAssetType(channelsToUnsubscribe[i].Currency)
			if err != nil {
				errs = append(errs, err)
				continue channels
			}

			formattedPair, err := de.FormatExchangeCurrency(channelsToUnsubscribe[i].Currency, a)
			if err != nil {
				errs = append(errs, err)
				continue channels
			}
			streamName, err := StreamName(ch, StreamParams{InstrumentName: formattedPair.String(), Interval: "raw"})
			if err != nil {
				errs = append(errs, err)
				continue channels
			}
			sub.Params.Channels = append(sub.Params.Channels, streamName)
		}
	}
	err := de.Websocket.Conn.SendJSONMessage(sub)
	if err != nil {
		errs = append(errs, err)
	}
	de.Websocket.AddSuccessfulSubscriptions(channelsToUnsubscribe...)
	return nil
}

// GenerateDefaultSubscriptions generates default subscription
func (de *Deribit) GenerateDefaultSubscriptions() ([]stream.ChannelSubscription, error) {
	var subscriptions []stream.ChannelSubscription
	//subscriptions = append(subscriptions, stream.ChannelSubscription{
	//Channel: wsMarkets,
	//})
	var channels = []string{"ticker"} //, wsTrades, wsOrderbook}
	assets := de.GetAssetTypes(true)
	for a := range assets {
		pairs, err := de.GetEnabledPairs(assets[a])
		if err != nil {
			return nil, err
		}
		for z := range pairs {
			newPair := currency.NewPairWithDelimiter(pairs[z].Base.String(),
				pairs[z].Quote.String(),
				pairs[z].Delimiter)
			for _, c := range channels {
				subscriptions = append(subscriptions,
					stream.ChannelSubscription{
						Channel:  c,
						Currency: newPair,
						Asset:    assets[a],
					})
			}
		}
	}
	//if de.GetAuthenticatedAPISupport(exchange.WebsocketAuthentication) {
	//var authchan = []string{wsOrders, wsFills}
	//for x := range authchan {
	//subscriptions = append(subscriptions, stream.ChannelSubscription{
	//Channel: authchan[x],
	//})
	//}
	//}
	return subscriptions, nil
}

// wsReadData gets and passes on websocket messages for processing
func (de *Deribit) wsReadData() {
	de.Websocket.Wg.Add(1)
	defer de.Websocket.Wg.Done()

	for {
		select {
		case <-de.Websocket.ShutdownC:
			return
		default:
			resp := de.Websocket.Conn.ReadMessage()
			if resp.Raw == nil {
				return
			}

			err := de.wsHandleData(resp.Raw)
			if err != nil {
				de.Websocket.DataHandler <- err
			}
		}
	}
}

func timestampFromFloat64(ts float64) time.Time {
	secs := int64(ts)
	nsecs := int64((ts - float64(secs)) * 1e9)
	return time.Unix(secs, nsecs).UTC()
}

func (de *Deribit) wsHandleData(respRaw []byte) error {
	var result SubscribeWrapper
	err := json.Unmarshal(respRaw, &result)
	if err != nil {
		return err
	}

	switch ch := result.Params.Channel; {
	case strings.HasPrefix(ch, "ticker"):
		var data Ticker
		json.Unmarshal(result.Params.Data, &data)

		var p currency.Pair
		p, err = currency.NewPairFromString(data.InstrumentName)
		if err != nil {
			return err
		}

		var a asset.Item
		a, err = de.GetPairAssetType(p)
		if err != nil {
			return err
		}

		de.Websocket.DataHandler <- &ticker.Ticker{
			Price: ticker.Price{
				Last:   data.LastPrice,
				High:   data.Stats.High,
				Low:    data.Stats.Low,
				Bid:    data.BestBidPrice,
				Ask:    data.BestAskPrice,
				Volume: data.Stats.Volume,
				//QuoteVolume
				PriceATH: data.MaxPrice,
				//Open
				//Close
				Pair:         p,
				ExchangeName: de.Name,
				AssetType:    a,
				LastUpdated:  time.Unix(0, data.Timestamp*int64(time.Millisecond)),
				OpenInterest: data.OpenInterest,
			},
			DerivStatus: ticker.DerivStatus{
				DerivPrice:     data.LastPrice,
				SpotPrice:      data.UnderlyingPrice,
				CurrentFunding: data.InterestRate,
				MarkPrice:      data.MarkPrice,
				OpenInterest:   data.OpenInterest,
			},
		}
	}
	//case wsUpdate:
	//var p currency.Pair
	//var a asset.Item
	//market, ok := result["market"]
	//if ok {
	//p, err = currency.NewPairFromString(market.(string))
	//if err != nil {
	//return err
	//}
	//a, err = de.GetPairAssetType(p)
	//if err != nil {
	//return err
	//}
	//}
	//switch result["channel"] {
	//case wsTicker:
	//var resultData WsTickerDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//de.Websocket.DataHandler <- &ticker.Price{
	//ExchangeName: de.Name,
	//Bid:          resultData.Ticker.Bid,
	//Ask:          resultData.Ticker.Ask,
	//Last:         resultData.Ticker.Last,
	//LastUpdated:  timestampFromFloat64(resultData.Ticker.Time),
	//Pair:         p,
	//AssetType:    a,
	//}
	//case wsOrderbook:
	//var resultData WsOrderbookDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//if len(resultData.OBData.Asks) == 0 && len(resultData.OBData.Bids) == 0 {
	//return nil
	//}
	//err = de.WsProcessUpdateOB(&resultData.OBData, p, a)
	//if err != nil {
	//err2 := de.wsResubToOB(p)
	//if err2 != nil {
	//de.Websocket.DataHandler <- err2
	//}
	//return err
	//}
	//case wsTrades:
	//var resultData WsTradeDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//for z := range resultData.TradeData {
	//var oSide order.Side
	//oSide, err = order.StringToOrderSide(resultData.TradeData[z].Side)
	//if err != nil {
	//de.Websocket.DataHandler <- order.ClassificationError{
	//Exchange: de.Name,
	//Err:      err,
	//}
	//}
	//de.Websocket.DataHandler <- stream.TradeData{
	//Timestamp:    resultData.TradeData[z].Time,
	//CurrencyPair: p,
	//AssetType:    a,
	//Exchange:     de.Name,
	//Price:        resultData.TradeData[z].Price,
	//Amount:       resultData.TradeData[z].Size,
	//Side:         oSide,
	//}
	//}
	//case wsOrders:
	//var resultData WsOrderDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//var pair currency.Pair
	//pair, err = currency.NewPairFromString(resultData.OrderData.Market)
	//if err != nil {
	//return err
	//}
	//var assetType asset.Item
	//assetType, err = de.GetPairAssetType(pair)
	//if err != nil {
	//return err
	//}
	//var oSide order.Side
	//oSide, err = order.StringToOrderSide(resultData.OrderData.Side)
	//if err != nil {
	//de.Websocket.DataHandler <- order.ClassificationError{
	//Exchange: de.Name,
	//Err:      err,
	//}
	//}
	//var resp order.Detail
	//resp.Side = oSide
	//resp.Amount = resultData.OrderData.Size
	//resp.AssetType = assetType
	//resp.ClientOrderID = resultData.OrderData.ClientID
	//resp.Exchange = de.Name
	//resp.ExecutedAmount = resultData.OrderData.FilledSize
	//resp.ID = strconv.FormatInt(resultData.OrderData.ID, 10)
	//resp.Pair = pair
	//resp.RemainingAmount = resultData.OrderData.Size - resultData.OrderData.FilledSize
	//var orderVars OrderVars
	//orderVars, err = de.compatibleOrderVars(resultData.OrderData.Side,
	//resultData.OrderData.Status,
	//resultData.OrderData.OrderType,
	//resultData.OrderData.FilledSize,
	//resultData.OrderData.Size,
	//resultData.OrderData.AvgFillPrice)
	//if err != nil {
	//return err
	//}
	//resp.Status = orderVars.Status
	//resp.Side = orderVars.Side
	//resp.Type = orderVars.OrderType
	//resp.Fee = orderVars.Fee
	//de.Websocket.DataHandler <- &resp
	//case wsFills:
	//var resultData WsFillsDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//de.Websocket.DataHandler <- resultData.FillsData
	//default:
	//de.Websocket.DataHandler <- stream.UnhandledMessageWarning{Message: de.Name + stream.UnhandledMessage + string(respRaw)}
	//}
	//case wsPartial:
	//switch result["channel"] {
	//case "orderbook":
	//var p currency.Pair
	//var a asset.Item
	//market, ok := result["market"]
	//if ok {
	//p, err = currency.NewPairFromString(market.(string))
	//if err != nil {
	//return err
	//}
	//a, err = de.GetPairAssetType(p)
	//if err != nil {
	//return err
	//}
	//}
	//var resultData WsOrderbookDataStore
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//err = de.WsProcessPartialOB(&resultData.OBData, p, a)
	//if err != nil {
	//err2 := de.wsResubToOB(p)
	//if err2 != nil {
	//de.Websocket.DataHandler <- err2
	//}
	//return err
	//}
	//// reset obchecksum failure blockage for pair
	//delete(obSuccess, p)
	//case wsMarkets:
	//var resultData WSMarkets
	//err = json.Unmarshal(respRaw, &resultData)
	//if err != nil {
	//return err
	//}
	//de.Websocket.DataHandler <- resultData.Data
	//}
	//case "error":
	//de.Websocket.DataHandler <- stream.UnhandledMessageWarning{
	//Message: de.Name + stream.UnhandledMessage + string(respRaw),
	//}
	//}
	return nil
}

// WsProcessUpdateOB processes an update on the orderbook
//func (de *Deribit) WsProcessUpdateOB(data *WsOrderbookData, p currency.Pair, a asset.Item) error {
//update := buffer.Update{
//Asset:      a,
//Pair:       p,
//UpdateTime: timestampFromFloat64(data.Time),
//}

//var err error
//for x := range data.Bids {
//update.Bids = append(update.Bids, orderbook.Item{
//Price:  data.Bids[x][0],
//Amount: data.Bids[x][1],
//})
//}
//for x := range data.Asks {
//update.Asks = append(update.Asks, orderbook.Item{
//Price:  data.Asks[x][0],
//Amount: data.Asks[x][1],
//})
//}

//err = de.Websocket.Orderbook.Update(&update)
//if err != nil {
//return err
//}

//updatedOb := de.Websocket.Orderbook.GetOrderbook(p, a)
//checksum := de.CalcUpdateOBChecksum(updatedOb)

//if checksum != data.Checksum {
//log.Warnf(log.ExchangeSys, "%s checksum failure for item %s",
//de.Name,
//p)
//return errors.New("checksum failed")
//}
//return nil
//}

//func (de *Deribit) wsResubToOB(p currency.Pair) error {
//if ok := obSuccess[p]; ok {
//return nil
//}

//obSuccess[p] = true

//channelToResubscribe := &stream.ChannelSubscription{
//Channel:  wsOrderbook,
//Currency: p,
//}
//err := de.Websocket.ResubscribeToChannel(channelToResubscribe)
//if err != nil {
//return fmt.Errorf("%s resubscribe to orderbook failure %s", de.Name, err)
//}
//return nil
//}

//// WsProcessPartialOB creates an OB from websocket data
//func (de *Deribit) WsProcessPartialOB(data *WsOrderbookData, p currency.Pair, a asset.Item) error {
//signedChecksum := de.CalcPartialOBChecksum(data)
//if signedChecksum != data.Checksum {
//return fmt.Errorf("%s channel: %s. Orderbook partial for %v checksum invalid",
//de.Name,
//a,
//p)
//}
//var bids, asks []orderbook.Item
//for x := range data.Bids {
//bids = append(bids, orderbook.Item{
//Price:  data.Bids[x][0],
//Amount: data.Bids[x][1],
//})
//}
//for x := range data.Asks {
//asks = append(asks, orderbook.Item{
//Price:  data.Asks[x][0],
//Amount: data.Asks[x][1],
//})
//}

//newOrderBook := orderbook.Base{
//Asks:         asks,
//Bids:         bids,
//AssetType:    a,
//LastUpdated:  timestampFromFloat64(data.Time),
//Pair:         p,
//ExchangeName: de.Name,
//}
//return de.Websocket.Orderbook.LoadSnapshot(&newOrderBook)
//}

//// CalcPartialOBChecksum calculates checksum of partial OB data received from WS
//func (de *Deribit) CalcPartialOBChecksum(data *WsOrderbookData) int64 {
//var checksum strings.Builder
//var price, amount string
//for i := 0; i < 100; i++ {
//if len(data.Bids)-1 >= i {
//price = checksumParseNumber(data.Bids[i][0])
//amount = checksumParseNumber(data.Bids[i][1])
//checksum.WriteString(price + ":" + amount + ":")
//}
//if len(data.Asks)-1 >= i {
//price = checksumParseNumber(data.Asks[i][0])
//amount = checksumParseNumber(data.Asks[i][1])
//checksum.WriteString(price + ":" + amount + ":")
//}
//}
//checksumStr := strings.TrimSuffix(checksum.String(), ":")
//return int64(crc32.ChecksumIEEE([]byte(checksumStr)))
//}

//// CalcUpdateOBChecksum calculates checksum of update OB data received from WS
//func (de *Deribit) CalcUpdateOBChecksum(data *orderbook.Base) int64 {
//var checksum strings.Builder
//var price, amount string
//for i := 0; i < 100; i++ {
//if len(data.Bids)-1 >= i {
//price = checksumParseNumber(data.Bids[i].Price)
//amount = checksumParseNumber(data.Bids[i].Amount)
//checksum.WriteString(price + ":" + amount + ":")
//}
//if len(data.Asks)-1 >= i {
//price = checksumParseNumber(data.Asks[i].Price)
//amount = checksumParseNumber(data.Asks[i].Amount)
//checksum.WriteString(price + ":" + amount + ":")
//}
//}
//checksumStr := strings.TrimSuffix(checksum.String(), ":")
//return int64(crc32.ChecksumIEEE([]byte(checksumStr)))
//}

//func checksumParseNumber(num float64) string {
//modifier := byte('f')
//if num < 0.0001 {
//modifier = 'e'
//}
//r := strconv.FormatFloat(num, modifier, -1, 64)
//if strings.IndexByte(r, '.') == -1 && modifier != 'e' {
//r += ".0"
//}
//return r
//}
