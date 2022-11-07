<div align="center">
<a href="https://nakji.network"><img alt="Nakji" src="https://github.com/nakji-network/landing/raw/master/src/images/logo.svg" width="300" /></a>
<br/>
<strong></strong>
<h1>Nakji Connector SDK (Golang)</h1>
</div>
<p align="center">
<a href="https://github.com/nakji-network/connector/actions/workflows/go.yml"><img alt="Github Workflow" src="https://github.com/nakji-network/connector/actions/workflows/go.yml/badge.svg" /></a>
<a href="https://godoc.org/github.com/nakji-network/connector"><img alt="GoDoc" src="https://godoc.org/github.com/nakji-network/connector?status.svg" /></a>
<a href="https://goreportcard.com/report/github.com/nakji-network/connector"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/nakji-network/connector" /></a>
</p>

This is a library to help anyone integrate a new data source (aka connector) with [Nakji Network](https://nakji.network).

The library handles: 

- Connector boilerplate
- Getting configs
- Publishing and Subscribing to the Nakji message queue
- Initializing monitoring
- [Healthcheck support](https://pkg.go.dev/github.com/heptiolabs/healthcheck)

Connector examples are in [examples/](examples).

You can read more documentation on this library in the [Nakji Documentation](https://docs.nakji.network).
