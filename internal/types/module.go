package types

import (
	"context"
)

// Source is an interface for modules that produce node data
type Source interface {
	Initialize(ctx context.Context) error
	Fetch(ctx context.Context) ([]Node, error)
	Watch(ctx context.Context) (<-chan []Node, <-chan error)
	Close() error
}

// Sink is an interface for modules that consume node data
type Sink interface {
	Initialize(ctx context.Context) error
	Process(ctx context.Context, nodes []Node) error
	Close() error
}
