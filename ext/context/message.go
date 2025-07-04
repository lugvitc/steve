package context

import (
	"context"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Message struct {
	ctx context.Context
	*events.Message
}

func (m *Message) ArgsN(n int) []string {
	if m.Message.Message.Conversation != nil {
		return strings.SplitN(*m.Message.Message.Conversation, " ", n)
	} else if m.Message.Message.ExtendedTextMessage != nil {
		return strings.SplitN(*m.Message.Message.ExtendedTextMessage.Text, " ", n)
	}
	return []string{}
}

func (m *Message) Args() []string {
	if m.Message.Message.Conversation != nil {
		return strings.Fields(*m.Message.Message.Conversation)
	} else if m.Message.Message.ExtendedTextMessage != nil {
		return strings.Fields(*m.Message.Message.ExtendedTextMessage.Text)
	}
	return []string{}
}

func (m *Message) GetText() string {
	if m.Message.Message.Conversation != nil {
		return m.Message.Message.GetConversation()
	} else if m.Message.Message.ExtendedTextMessage != nil {
		return m.Message.Message.ExtendedTextMessage.GetText()
	}
	return ""
}

func (m *Message) Send(client *whatsmeow.Client, to types.JID, text string) (resp whatsmeow.SendResponse, err error) {
	return client.SendMessage(m.ctx, to, &waE2E.Message{
		Conversation: &text,
	})
}

func (m *Message) Reply(client *whatsmeow.Client, text string) (whatsmeow.SendResponse, error) {
	return client.SendMessage(m.ctx, m.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: &text,
			ContextInfo: &waE2E.ContextInfo{
				StanzaID:      &m.Info.ID,
				Participant:   stringPtr(m.Info.Sender.String()),
				QuotedMessage: m.Message.Message,
			},
		},
	})
}

func (m *Message) ReplyCustom(client *whatsmeow.Client, text string, custom *waE2E.ContextInfo) (whatsmeow.SendResponse, error) {
	return client.SendMessage(m.ctx, m.Info.Chat, &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        &text,
			ContextInfo: custom,
		},
	})
}

func (m *Message) Edit(client *whatsmeow.Client, text string) (whatsmeow.SendResponse, error) {
	return client.SendMessage(m.ctx, m.Info.Chat, client.BuildEdit(m.Info.Chat, m.Info.ID, &waE2E.Message{
		Conversation: &text,
	}))
}
