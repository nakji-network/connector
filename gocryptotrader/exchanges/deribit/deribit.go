package deribit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/common/convert"
	"github.com/thrasher-corp/gocryptotrader/common/crypto"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/log"
)

// Deribit is the overarching type across this package
type Deribit struct {
	exchange.Base
}

const (
	deribitAPIURL     = "https://www.deribit.com/ws/api/v2"
	deribitAPIVersion = "2.0.1"

	// Use this javascript to generate list of endpoints, at https://docs.deribit.com/
	/*
		let endpoints = [], results = []
		const snakeToCamel = (str) => str.replace(/([-_\/][a-z])/g,(group) => group.toUpperCase().replace('-', '').replace('_', '').replace('/', ''))
		for (let x = 5; x<15; x++) {
			endpoints = endpoints.concat(...$(`#toc > li:nth-child(${x}) > ul > li > a`).map((i, x)=>x.text))
		}
		results.push("\n// Public endpoints")
		results = results.concat(endpoints.filter(e => e.startsWith("/public")).map(x => `${snakeToCamel(x)} = "${x}"`))
		results.push("\n// Authenticated endpoints")
		results = results.concat(endpoints.filter(e => e.startsWith("/private")).map(x => `${snakeToCamel(x)} = "${x}"`))
		results.push("\n// Websocket channels")
		results = results.concat(endpoints.filter(e => !e.startsWith("/")).map(x => `//"${x}"`))
		console.log(results.join("\n"))
	*/

	// Public endpoints
	PublicAuth                             = "/public/auth"
	PublicExchangeToken                    = "/public/exchange_token"
	PublicForkToken                        = "/public/fork_token"
	PublicSetHeartbeat                     = "/public/set_heartbeat"
	PublicDisableHeartbeat                 = "/public/disable_heartbeat"
	PublicGetTime                          = "/public/get_time"
	PublicHello                            = "/public/hello"
	PublicTest                             = "/public/test"
	PublicSubscribe                        = "/public/subscribe"
	PublicUnsubscribe                      = "/public/unsubscribe"
	PublicGetAnnouncements                 = "/public/get_announcements"
	PublicGetBookSummaryByCurrency         = "/public/get_book_summary_by_currency"
	PublicGetBookSummaryByInstrument       = "/public/get_book_summary_by_instrument"
	PublicGetContractSize                  = "/public/get_contract_size"
	PublicGetCurrencies                    = "/public/get_currencies"
	PublicGetFundingChartData              = "/public/get_funding_chart_data"
	PublicGetFundingRateHistory            = "/public/get_funding_rate_history"
	PublicGetFundingRateValue              = "/public/get_funding_rate_value"
	PublicGetHistoricalVolatility          = "/public/get_historical_volatility"
	PublicGetIndex                         = "/public/get_index"
	PublicGetInstruments                   = "/public/get_instruments"
	PublicGetLastSettlementsByCurrency     = "/public/get_last_settlements_by_currency"
	PublicGetLastSettlementsByInstrument   = "/public/get_last_settlements_by_instrument"
	PublicGetLastTradesByCurrency          = "/public/get_last_trades_by_currency"
	PublicGetLastTradesByCurrencyAndTime   = "/public/get_last_trades_by_currency_and_time"
	PublicGetLastTradesByInstrument        = "/public/get_last_trades_by_instrument"
	PublicGetLastTradesByInstrumentAndTime = "/public/get_last_trades_by_instrument_and_time"
	PublicGetOrderBook                     = "/public/get_order_book"
	PublicGetTradeVolumes                  = "/public/get_trade_volumes"
	PublicGetTradingviewChartData          = "/public/get_tradingview_chart_data"
	PublicTicker                           = "/public/ticker"

	// Authenticated endpoints
	PrivateLogout                            = "/private/logout"
	PrivateEnableCancelOnDisconnect          = "/private/enable_cancel_on_disconnect"
	PrivateDisableCancelOnDisconnect         = "/private/disable_cancel_on_disconnect"
	PrivateGetCancelOnDisconnect             = "/private/get_cancel_on_disconnect"
	PrivateSubscribe                         = "/private/subscribe"
	PrivateUnsubscribe                       = "/private/unsubscribe"
	PrivateChangeApiKeyName                  = "/private/change_api_key_name"
	PrivateChangeScopeInApiKey               = "/private/change_scope_in_api_key"
	PrivateChangeSubaccountName              = "/private/change_subaccount_name"
	PrivateCreateApiKey                      = "/private/create_api_key"
	PrivateCreateSubaccount                  = "/private/create_subaccount"
	PrivateDisableApiKey                     = "/private/disable_api_key"
	PrivateDisableTfaForSubaccount           = "/private/disable_tfa_for_subaccount"
	PrivateEnableAffiliateProgram            = "/private/enable_affiliate_program"
	PrivateEnableApiKey                      = "/private/enable_api_key"
	PrivateGetAccountSummary                 = "/private/get_account_summary"
	PrivateGetAffiliateProgramInfo           = "/private/get_affiliate_program_info"
	PrivateGetEmailLanguage                  = "/private/get_email_language"
	PrivateGetNewAnnouncements               = "/private/get_new_announcements"
	PrivateGetPosition                       = "/private/get_position"
	PrivateGetPositions                      = "/private/get_positions"
	PrivateGetSubaccounts                    = "/private/get_subaccounts"
	PrivateListApiKeys                       = "/private/list_api_keys"
	PrivateRemoveApiKey                      = "/private/remove_api_key"
	PrivateRemoveSubaccount                  = "/private/remove_subaccount"
	PrivateResetApiKey                       = "/private/reset_api_key"
	PrivateSetAnnouncementAsRead             = "/private/set_announcement_as_read"
	PrivateSetApiKeyAsDefault                = "/private/set_api_key_as_default"
	PrivateSetEmailForSubaccount             = "/private/set_email_for_subaccount"
	PrivateSetEmailLanguage                  = "/private/set_email_language"
	PrivateSetPasswordForSubaccount          = "/private/set_password_for_subaccount"
	PrivateToggleNotificationsFromSubaccount = "/private/toggle_notifications_from_subaccount"
	PrivateToggleSubaccountLogin             = "/private/toggle_subaccount_login"
	PrivateExecuteBlockTrade                 = "/private/execute_block_trade"
	PrivateGetBlockTrade                     = "/private/get_block_trade"
	PrivateGetLastBlockTradesByCurrency      = "/private/get_last_block_trades_by_currency"
	PrivateInvalidateBlockTradeSignature     = "/private/invalidate_block_trade_signature"
	PrivateVerifyBlockTrade                  = "/private/verify_block_trade"
	PrivateBuy                               = "/private/buy"
	PrivateSell                              = "/private/sell"
	PrivateEdit                              = "/private/edit"
	PrivateCancel                            = "/private/cancel"
	PrivateCancelAll                         = "/private/cancel_all"
	PrivateCancelAllByCurrency               = "/private/cancel_all_by_currency"
	PrivateCancelAllByInstrument             = "/private/cancel_all_by_instrument"
	PrivateCancelByLabel                     = "/private/cancel_by_label"
	PrivateClosePosition                     = "/private/close_position"
	PrivateGetMargins                        = "/private/get_margins"
	PrivateGetMmpConfig                      = "/private/get_mmp_config"
	PrivateGetOpenOrdersByCurrency           = "/private/get_open_orders_by_currency"
	PrivateGetOpenOrdersByInstrument         = "/private/get_open_orders_by_instrument"
	PrivateGetOrderHistoryByCurrency         = "/private/get_order_history_by_currency"
	PrivateGetOrderHistoryByInstrument       = "/private/get_order_history_by_instrument"
	PrivateGetOrderMarginByIds               = "/private/get_order_margin_by_ids"
	PrivateGetOrderState                     = "/private/get_order_state"
	PrivateGetStopOrderHistory               = "/private/get_stop_order_history"
	PrivateGetUserTradesByCurrency           = "/private/get_user_trades_by_currency"
	PrivateGetUserTradesByCurrencyAndTime    = "/private/get_user_trades_by_currency_and_time"
	PrivateGetUserTradesByInstrument         = "/private/get_user_trades_by_instrument"
	PrivateGetUserTradesByInstrumentAndTime  = "/private/get_user_trades_by_instrument_and_time"
	PrivateGetUserTradesByOrder              = "/private/get_user_trades_by_order"
	PrivateResetMmp                          = "/private/reset_mmp"
	PrivateSetMmpConfig                      = "/private/set_mmp_config"
	PrivateGetSettlementHistoryByInstrument  = "/private/get_settlement_history_by_instrument"
	PrivateGetSettlementHistoryByCurrency    = "/private/get_settlement_history_by_currency"
	PrivateCancelTransferById                = "/private/cancel_transfer_by_id"
	PrivateCancelWithdrawal                  = "/private/cancel_withdrawal"
	PrivateCreateDepositAddress              = "/private/create_deposit_address"
	PrivateGetCurrentDepositAddress          = "/private/get_current_deposit_address"
	PrivateGetDeposits                       = "/private/get_deposits"
	PrivateGetTransfers                      = "/private/get_transfers"
	PrivateGetWithdrawals                    = "/private/get_withdrawals"
	PrivateSubmitTransferToSubaccount        = "/private/submit_transfer_to_subaccount"
	PrivateSubmitTransferToUser              = "/private/submit_transfer_to_user"
	PrivateWithdraw                          = "/private/withdraw"

	// Websocket channels
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

func (d *Deribit) GetInstruments(currency string, kind Kind, expired bool) ([]Instrument, error) {
	params := url.Values{}
	params.Add("currency", currency)
	if kind != NA {
		params.Add("kind", string(kind))
	}
	if expired == true {
		params.Add("expired", "true")
	}

	var instruments Instruments
	return instruments.Result, d.SendHTTPRequest(exchange.RestSpot, PublicGetInstruments+"?"+params.Encode(), &instruments)
}

// SendHTTPRequest sends an unauthenticated request
func (b *Deribit) SendHTTPRequest(ep exchange.URL, path string, result interface{}) error {
	endpoint, err := b.API.Endpoints.GetURL(ep)
	if err != nil {
		return err
	}
	var tempResp json.RawMessage
	var errCap errorCapture
	err = b.SendPayload(context.Background(), &request.Item{
		Method:        http.MethodGet,
		Path:          endpoint + path,
		Result:        &tempResp,
		Verbose:       b.Verbose,
		HTTPDebugging: b.HTTPDebugging,
		HTTPRecording: b.HTTPRecording,
	})
	if err != nil {
		return err
	}
	if err := json.Unmarshal(tempResp, &errCap); err == nil {
		if errCap.Code != 200 && errCap.ErrMsg != "" {
			return errors.New(errCap.ErrMsg)
		}
	}
	return json.Unmarshal(tempResp, result)
}

// TODO: SendAuthHTTPRequest sends an authenticated HTTP request
func (b *Deribit) SendAuthHTTPRequest(method, path string, params url.Values, f request.EndpointLimit, result interface{}) error {
	if !b.AllowAuthenticatedRequest() {
		return fmt.Errorf("%s %w", b.Name, exchange.ErrAuthenticatedRequestWithoutCredentialsSet)
	}

	if params == nil {
		params = url.Values{}
	}
	recvWindow := 5 * time.Second
	params.Set("recvWindow", strconv.FormatInt(convert.RecvWindow(recvWindow), 10))
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix()*1000, 10))

	signature := params.Encode()
	hmacSigned := crypto.GetHMAC(crypto.HashSHA256, []byte(signature), []byte(b.API.Credentials.Secret))
	hmacSignedStr := crypto.HexEncodeToString(hmacSigned)

	headers := make(map[string]string)
	headers["X-MBX-APIKEY"] = b.API.Credentials.Key

	if b.Verbose {
		log.Debugf(log.ExchangeSys, "sent path: %s", path)
	}

	path = common.EncodeURLValues(path, params)
	path += "&signature=" + hmacSignedStr

	interim := json.RawMessage{}

	errCap := struct {
		Success bool   `json:"success"`
		Message string `json:"msg"`
	}{}

	ctx, cancel := context.WithTimeout(context.Background(), recvWindow)
	defer cancel()
	err := b.SendPayload(ctx, &request.Item{
		Method:        method,
		Path:          path,
		Headers:       headers,
		Body:          bytes.NewBuffer(nil),
		Result:        &interim,
		AuthRequest:   true,
		Verbose:       b.Verbose,
		HTTPDebugging: b.HTTPDebugging,
		HTTPRecording: b.HTTPRecording,
		Endpoint:      f})
	if err != nil {
		return err
	}

	if err := json.Unmarshal(interim, &errCap); err == nil {
		if !errCap.Success && errCap.Message != "" {
			return errors.New(errCap.Message)
		}
	}

	return json.Unmarshal(interim, result)
}
