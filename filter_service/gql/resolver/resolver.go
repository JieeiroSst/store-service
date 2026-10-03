package resolver

import "github.com/JIeeiroSst/filter-service/adapter"

type Resolver struct {
	Clients *adapter.Clients
}

func NewResolver(clients *adapter.Clients) *Resolver {
	return &Resolver{Clients: clients}
}
