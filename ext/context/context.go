package context

import (
	"context"

	"github.com/lugvitc/steve/core/sql"
	"github.com/lugvitc/steve/logger"

	"go.mau.fi/whatsmeow/types/events"
)

func stringPtr(s string) *string {
	return &s
}

type Context struct {
	Message *Message
	Logger  *logger.Logger
	context.Context
}

func New(ctx context.Context, evt interface{}) *Context {
	switch v := evt.(type) {
	case *events.Message:
		// buf, _ := json.MarshalIndent(v, "", " ") // For debugging purposes
		// fmt.Println(string(buf))
		sql.AddMessage(v.Info.ID, v.Info.PushName, v.Info.Sender.User, v.Info.Timestamp.Unix())
		return &Context{
			Message: &Message{
				Message: v,
				ctx:     ctx,
			},
			Context: ctx,
		}
	}
	return &Context{Context: ctx}
}
