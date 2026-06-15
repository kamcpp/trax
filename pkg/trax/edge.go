package trax

import (
	"context"
)

type Edge interface {
	Start(ctx context.Context) error
}

type defaultEdge struct {
	mqClient MQClient
	store    Store
}

func NewEdge(mqClient MQClient, store Store) Edge {
	return &defaultEdge{
		mqClient: mqClient,
		store:    store,
	}
}

func (c *defaultEdge) Start(ctx context.Context) error {

	return nil
}
