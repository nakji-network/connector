package chain

type Clients struct {
	rpcMap map[string]RPCs
}

type RPCs struct {
	Full []string
	//Archival []string
}

func NewClients(rpcMap map[string]RPCs) *Clients {
	return &Clients{rpcMap}
}
