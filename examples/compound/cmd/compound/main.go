package main

import (
	"github.com/nakji-network/connector/examples/compound"
)

func main() {
	cc := compound.NewConnector()
	cc.Start()
}
