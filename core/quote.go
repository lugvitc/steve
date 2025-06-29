package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/fogleman/gg"
	"github.com/lugvitc/steve/core/sql"
	"github.com/lugvitc/steve/ext"
	sve_context "github.com/lugvitc/steve/ext/context"
	"github.com/lugvitc/steve/ext/handlers"
	waLogger "github.com/lugvitc/steve/logger"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

var (
	darkBgColor     = color.RGBA{R: 11, G: 20, B: 26, A: 255}    // #0B141A
	darkBubbleColor = color.RGBA{R: 29, G: 39, B: 42, A: 255}    // #202C33
	darkNameColor   = color.RGBA{R: 83, G: 189, B: 235, A: 255}  // #53BDEB
	darkTextColor   = color.RGBA{R: 233, G: 237, B: 239, A: 255} // #E9EDEF
	darkTimeColor   = color.RGBA{R: 134, G: 150, B: 160, A: 255} // #8696A0
)

func createQuoteImage(pfp image.Image, name, text string, timestamp time.Time) ([]byte, error) {
	const (
		minWidth       = 400.0
		maxWidth       = 514.0
		padding        = 15.0
		pfpSize        = 60.0
		pfpRightMargin = 10.0
		bubbleRadius   = 15.0
	)

	// font location change
	if _, err := gg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 20); err != nil {
		return nil, fmt.Errorf("could not load bold font. Make sure a valid font file exists")
	}
	if _, err := gg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 22); err != nil {
		return nil, fmt.Errorf("could not load regular font. Make sure a valid font file exists")
	}

	tempDC := gg.NewContext(0, 0)
	// this too
	regularFont, err := gg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 20)
	if err != nil {
		return nil, err
	}

	regularFontSmall, err := gg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 18)
	if err != nil {
		return nil, err
	}

	tempDC.SetFontFace(regularFont)

	lineSpacing := 1.4

	wrappedLines := tempDC.WordWrap(text, maxWidth-(padding*2))
	// this too
	boldFont, err := gg.LoadFontFace("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 20)
	if err != nil {
		return nil, err
	}
	tempDC.SetFontFace(boldFont)
	nameWidth, _ := tempDC.MeasureString(name)

	tempDC.SetFontFace(regularFont)
	maxTextWidth := 0.0
	for _, line := range wrappedLines {
		lineWidth, _ := tempDC.MeasureString(line)
		if lineWidth > maxTextWidth {
			maxTextWidth = lineWidth
		}
	}

	contentWidth := math.Max(nameWidth, maxTextWidth)
	finalWidth := minWidth
	if contentWidth > minWidth-(padding*2) {
		finalWidth = contentWidth + (padding * 2)
	}
	if finalWidth > maxWidth {
		finalWidth = maxWidth
	}

	textBlockHeight := float64(len(wrappedLines)) * tempDC.FontHeight() * lineSpacing
	headerHeight := 30.0
	footerHeight := 30.0
	totalHeight := headerHeight + textBlockHeight + footerHeight + (padding * 2) + 10

	dc := gg.NewContext(int(finalWidth), int(totalHeight))

	// todo: remove this
	// dc.SetColor(color.RGBA{R: 255, G: 255, B: 255, A: 255}) // white background

	dc.Clear()

	dc.SetColor(darkBubbleColor)

	// the tail (bulge)
	tailX := pfpSize + 2
	tailY := 15.0
	dc.NewSubPath()
	dc.MoveTo(tailX, tailY)
	dc.LineTo(tailX+30, tailY)
	dc.LineTo(tailX+15, tailY+18)
	dc.ClosePath()
	dc.Fill()

	// rectangle for the bubble
	dc.DrawRoundedRectangle(padding+pfpSize, padding, finalWidth-(padding*2)-pfpSize, totalHeight-(padding*2), bubbleRadius)
	dc.Fill()

	// pfp
	pfp = resizeImage(pfp, pfpSize, pfpSize)
	pfpX := padding - pfpRightMargin + pfpSize/2
	pfpY := padding + 30.0
	dc.DrawCircle(pfpX, pfpY, pfpSize/2)
	dc.Clip()
	dc.DrawImage(pfp, int(pfpX-pfpSize/2), int(pfpY-pfpSize/2))
	dc.ResetClip()

	contentX := padding + pfpSize + pfpRightMargin*2
	contentY := padding + 30.0

	dc.SetFontFace(boldFont)
	dc.SetColor(darkNameColor)
	dc.DrawString(name, contentX, contentY)
	contentY += 13.0

	dc.SetFontFace(regularFont)
	dc.SetColor(darkTextColor)
	dc.SetLineWidth(lineSpacing)
	dc.DrawStringWrapped(text, contentX, contentY, 0, 0, (finalWidth - contentX - padding*2), 1, gg.AlignLeft)

	dc.SetFontFace(regularFontSmall)

	timeStr := timestamp.Format("3:04 PM")
	timeWidth, _ := dc.MeasureString(timeStr)
	timeX := finalWidth - padding - timeWidth - 13.0
	timeY := totalHeight - padding - 10.0
	dc.SetColor(darkTimeColor)
	dc.DrawString(timeStr, timeX, timeY)

	dim := 512.0
	scaleFactor := float64(dim) / float64(finalWidth)
	scaledHeight := float64(totalHeight) * scaleFactor
	offsetY := (float64(dim) - scaledHeight) / 2
	offsetX := 0.0
	if scaledHeight > dim {
		dim = scaledHeight
		offsetY = 0
		offsetX = (dim - finalWidth) / 2
	}

	dc1 := gg.NewContext(int(dim), int(dim))
	dc1.Push()
	dc1.Translate(offsetX, offsetY)
	dc1.Scale(scaleFactor, scaleFactor)
	dc1.DrawImage(dc.Image(), 0, 0)
	dc1.Pop()

	var buf bytes.Buffer
	if err := webp.Encode(&buf, dc1.Image(), &webp.Options{Lossless: true}); err != nil {
		return nil, fmt.Errorf("failed to encode to webp: %w", err)
	}

	return buf.Bytes(), nil

	// png:

	// var buf bytes.Buffer
	// err = dc.EncodePNG(&buf)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to encode image to PNG: %w", err)
	// }

	// return buf.Bytes(), nil
}

func createDefaultPFP() image.Image {
	dc := gg.NewContext(60, 60)
	dc.SetHexColor("#344047")
	dc.DrawCircle(30, 30, 30)
	dc.Fill()
	return dc.Image()
}

func resizeImage(img image.Image, width, height float64) image.Image {
	origBounds := img.Bounds()
	origWidth := float64(origBounds.Dx())
	origHeight := float64(origBounds.Dy())

	if origWidth == 0 || origHeight == 0 {
		return createDefaultPFP()
	}

	resizedCtx := gg.NewContext(int(width), int(height))
	resizedCtx.Scale(width/origWidth, height/origHeight)
	resizedCtx.DrawImage(img, 0, 0)

	return resizedCtx.Image()
}

func (*Module) LoadQuote(dispatcher *ext.Dispatcher) {
	qLogger := LOGGER.Create("quote")
	defer qLogger.Println("Loaded Quote module")
	dispatcher.AddHandler(
		handlers.NewCommand("q", quote, qLogger.Create("q-cmd").
			ChangeLevel(waLogger.LevelInfo),
		).AddDescription("Reply to a message to create a sticker quote of it."),
	)
}

func quote(client *whatsmeow.Client, ctx *sve_context.Context) error {
	msg := ctx.Message.Message.Message
	if msg.GetExtendedTextMessage() == nil || msg.ExtendedTextMessage.GetContextInfo() == nil || msg.ExtendedTextMessage.ContextInfo.QuotedMessage == nil {
		_, err := reply(client, ctx.Message, "Please reply to a message to quote it.")
		return err
	}

	contextInfo := msg.ExtendedTextMessage.ContextInfo
	quotedMsg := contextInfo.QuotedMessage

	var msgText string
	switch {
	case quotedMsg.Conversation != nil:
		msgText = quotedMsg.GetConversation()
	case quotedMsg.ExtendedTextMessage != nil:
		msgText = quotedMsg.ExtendedTextMessage.GetText()
	}

	if msgText == "" {
		_, err := reply(client, ctx.Message, "Cannot quote an empty message or non-text content.")
		return err
	}

	senderJID, err := types.ParseJID(contextInfo.GetParticipant())
	if err != nil {
		senderJID = ctx.Message.Info.Sender
	}
	var senderName string
	savedMessage := sql.GetMessage(contextInfo.GetStanzaID())
	if savedMessage.PushName == "" {
		senderInfo, err := client.GetUserInfo([]types.JID{senderJID})
		if err == nil {
			senderName = *senderInfo[senderJID].VerifiedName.Details.VerifiedName
		}
	} else {
		senderName = savedMessage.PushName
	}

	if senderName == "" {
		senderName = strings.Split(senderJID.String(), "@")[0]
	}

	pfpImage, err := getProfilePicture(client, senderJID)
	if err != nil {
		// Don't return an error, just use the default PFP
		pfpImage = createDefaultPFP()
	}
	if savedMessage.Timestamp == 0 {
		savedMessage.Timestamp = time.Now().Unix()
	}

	msgTimestamp := time.Unix(savedMessage.Timestamp, 0)

	imageData, err := createQuoteImage(pfpImage, senderName, msgText, msgTimestamp)
	if err != nil {
		log.Printf("Failed to create quote image: %v", err)
		_, err := reply(client, ctx.Message, "Sorry, something went wrong while creating the quote.")
		return err
	}

	resp, err := client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		log.Printf("Failed to upload sticker: %v", err)
		_, err := reply(client, ctx.Message, "Failed to upload the sticker.")
		return err
	}
	// imageMsg := &waE2E.ImageMessage{
	// 	Mimetype: proto.String("image/png"), // replace this with the actual mime type
	// 	// you can also optionally add other fields like ContextInfo and JpegThumbnail here

	// 	URL:           &resp.URL,
	// 	DirectPath:    &resp.DirectPath,
	// 	MediaKey:      resp.MediaKey,
	// 	FileEncSHA256: resp.FileEncSHA256,
	// 	FileSHA256:    resp.FileSHA256,
	// 	FileLength:    &resp.FileLength,
	// 	ContextInfo: &waE2E.ContextInfo{
	// 		StanzaID:      &ctx.Message.Info.ID,
	// 		Participant:   proto.String(ctx.Message.Info.Sender.String()),
	// 		QuotedMessage: ctx.Message.Message.Message,
	// 	},
	// }

	stickerMsg := &waE2E.StickerMessage{
		Mimetype: proto.String("image/webp"), // replace this with the actual mime type
		// you can also optionally add other fields like ContextInfo and JpegThumbnail here

		URL:           &resp.URL,
		DirectPath:    &resp.DirectPath,
		MediaKey:      resp.MediaKey,
		FileEncSHA256: resp.FileEncSHA256,
		FileSHA256:    resp.FileSHA256,
		FileLength:    &resp.FileLength,
		ContextInfo: &waE2E.ContextInfo{
			StanzaID:      &ctx.Message.Info.ID,
			Participant:   proto.String(ctx.Message.Info.Sender.String()),
			QuotedMessage: ctx.Message.Message.Message,
		},
	}

	_, err = client.SendMessage(ctx, ctx.Message.Info.Chat, &waE2E.Message{
		StickerMessage: stickerMsg,
	})
	if err != nil {
		log.Printf("Failed to send sticker: %v", err)
	}
	return nil
}

func getProfilePicture(client *whatsmeow.Client, jid types.JID) (image.Image, error) {
	pfpInfo, err := client.GetProfilePictureInfo(jid, &whatsmeow.GetProfilePictureParams{Preview: true})
	if err != nil {
		return nil, fmt.Errorf("could not get PFP info: %w", err)
	}
	if pfpInfo == nil {
		return nil, errors.New("PFP info is nil, user may not have a picture")
	}

	resp, err := http.Get(pfpInfo.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download PFP: %w", err)
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	return img, err
}
