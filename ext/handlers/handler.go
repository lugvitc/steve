package handlers

import (
	"github.com/lugvitc/steve/ext/context"

	"go.mau.fi/whatsmeow"
)

type Handler interface {
	CheckUpdate(ctx *context.Context) bool
	HandleUpdate(client *whatsmeow.Client, ctx *context.Context) error
	AddDescription(desc string) Handler
	GetDescription() string
	GetName() string
}

type Response func(client *whatsmeow.Client, ctx *context.Context) error
