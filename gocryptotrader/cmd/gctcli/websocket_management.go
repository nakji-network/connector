package main

import (
	"context"
	"fmt"

	"github.com/thrasher-corp/gocryptotrader/gctrpc"
	"github.com/urfave/cli/v2"
)

var websocketManagerCommand = &cli.Command{
	Name:      "websocket",
	Usage:     "execute websocket management command",
	ArgsUsage: "<command> <args>",
	Subcommands: []*cli.Command{
		{
			Name:  "getinfo",
			Usage: "returns all exchange websocket information",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: getwebsocketInfo,
		},
		{
			Name:  "disable",
			Usage: "disables websocket connection for an exchange",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: enableDisableWebsocket,
		},
		{
			Name:  "enable",
			Usage: "enables websocket connection for an exchange",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				&cli.BoolFlag{
					Name:   "enable",
					Hidden: true,
				},
			},
			Action: enableDisableWebsocket,
		},
		{
			Name:  "getsubs",
			Usage: "returns current subscriptions for an exchange",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
			},
			Action: getSubscriptions,
		},
		{
			Name:  "setproxy",
			Usage: "sets exchange websocket proxy, flushes and reroutes connection",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				&cli.StringFlag{
					Name:  "proxy",
					Usage: "proxy address to change to, if proxy string is not set, this will stop the utilization of the prior set proxy.",
				},
			},
			Action: setProxy,
		},
		{
			Name:  "seturl",
			Usage: "sets exchange websocket endpoint URL and resets the websocket connection",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exchange",
					Usage: "the exchange to act on",
				},
				&cli.StringFlag{
					Name:  "url",
					Usage: "url string to change to, an empty string will set it back to the packaged defined default",
				},
			},
			Action: setURL,
		},
	},
}

func getwebsocketInfo(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return fmt.Errorf("[%s] is not a valid exchange", exchange)
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.WebsocketGetInfo(context.Background(),
		&gctrpc.WebsocketGetInfoRequest{Exchange: exchange})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func enableDisableWebsocket(c *cli.Context) error {
	enable := c.Bool("enable")
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return fmt.Errorf("[%s] is not a valid exchange", exchange)
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.WebsocketSetEnabled(context.Background(),
		&gctrpc.WebsocketSetEnabledRequest{Exchange: exchange, Enable: enable})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func getSubscriptions(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return fmt.Errorf("[%s] is not a valid exchange", exchange)
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.WebsocketGetSubscriptions(context.Background(),
		&gctrpc.WebsocketGetSubscriptionsRequest{Exchange: exchange})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func setProxy(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return fmt.Errorf("[%s] is not a valid exchange", exchange)
	}

	var proxy string
	if c.IsSet("proxy") {
		proxy = c.String("proxy")
	} else {
		proxy = c.Args().Get(1)
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.WebsocketSetProxy(context.Background(),
		&gctrpc.WebsocketSetProxyRequest{Exchange: exchange, Proxy: proxy})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}

func setURL(c *cli.Context) error {
	if c.NArg() == 0 && c.NumFlags() == 0 {
		return cli.ShowSubcommandHelp(c)
	}

	var exchange string
	if c.IsSet("exchange") {
		exchange = c.String("exchange")
	} else {
		exchange = c.Args().First()
	}

	if !validExchange(exchange) {
		return fmt.Errorf("[%s] is not a valid exchange", exchange)
	}

	var url string
	if c.IsSet("url") {
		url = c.String("url")
	} else {
		url = c.Args().Get(1)
	}

	conn, err := setupClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := gctrpc.NewGoCryptoTraderClient(conn)
	result, err := client.WebsocketSetURL(context.Background(),
		&gctrpc.WebsocketSetURLRequest{Exchange: exchange, Url: url})
	if err != nil {
		return err
	}
	jsonOutput(result)
	return nil
}
