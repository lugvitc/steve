package core

import (
	"bytes"
	"encoding/ascii85"
	"encoding/base64"
	"strings"

	"github.com/lugvitc/steve/ext/context"
)

func init() {
	NewCommand("b64", b64Encode).SetDescription("Encode text to Base64")
	NewCommand("b64d", b64Decode).SetDescription("Decode Base64 text")
	NewCommand("b85", b85Encode).SetDescription("Encode text to Base85")
	NewCommand("b85d", b85Decode).SetDescription("Decode Base85 text")
}

func b64Encode(ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b64 <text>")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return ctx.Reply("🔒 Base64:\n" + encoded)
}

func b64Decode(ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b64d <base64>")
	}
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return ctx.Reply("❌ Invalid Base64 input.")
	}
	return ctx.Reply("🔓 Decoded:\n" + string(decoded))
}

func b85Encode(ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b85 <text>")
	}
	buf := make([]byte, ascii85.MaxEncodedLen(len(text)))
	n := ascii85.Encode(buf, []byte(text))
	return ctx.Reply("🔒 Base85:\n" + string(buf[:n]))
}

func b85Decode(ctx *context.Context) error {
	text := strings.TrimSpace(ctx.Args)
	if text == "" {
		return ctx.Reply("Usage: .b85d <base85>")
	}
	decoded := make([]byte, len(text))
	n, _, err := ascii85.Decode(decoded, []byte(text), true)
	if err != nil {
		return ctx.Reply("❌ Invalid Base85 input.")
	}
	return ctx.Reply("🔓 Decoded:\n" + string(decoded[:n]))
}
