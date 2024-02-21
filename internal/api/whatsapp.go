package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/watchzap/internal/parser"
	"github.com/watchzap/internal/static"
)

type Whatsapp struct {
	Container *sqlstore.Container
	Client    *whatsmeow.Client
	Debug     bool
}

// Creates new Whatsapp struct and initializes the container, devices stores and the client.
// Returns a new instance of Whatsapp struct
func NewWhatsapp(debug bool) (*Whatsapp, error) {
	var dbLog waLog.Logger
	var clientLog waLog.Logger
	if debug {
		dbLog = waLog.Stdout("Database", "DEBUG", true)
		clientLog = waLog.Stdout("Client", "DEBUG", true)
	} else {
		dbLog = waLog.Stdout("Database", "INFO", true)
		clientLog = waLog.Stdout("Client", "INFO", true)
	}

	container, err := sqlstore.New("sqlite3", "file:zap.db?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, err
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, err
	}

	client := whatsmeow.NewClient(deviceStore, clientLog)

	return &Whatsapp{
		Container: container,
		Client:    client,
		Debug:     debug,
	}, nil
}

// Does the login procedure in Whatsapp
// First if there is no device stored in the database it shows a QR code
// Scan it the same way you would to access whatsapp web
// After that if there is no error the user is successfully logged in
func (w *Whatsapp) Login() error {
	if w.Client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := w.Client.GetQRChannel(context.Background())
		err := w.Client.Connect()
		if err != nil {
			return err
		}

		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				log.Info().Msg("WZ: Please connect to whatsapp by scanning this QR code")

				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				log.Info().Msg(fmt.Sprintf("WZ: Login event: %s", evt.Event))
			}
		}
	} else {
		// Already logged in
		err := w.Client.Connect()
		if err != nil {
			return err
		}

		log.Info().Str("user", w.Client.Store.ID.User).Msg("WZ: Connected to WhatsApp")
	}

	return nil
}

func (w *Whatsapp) GenerateMessage(
	m parser.Message,
) (*waProto.Message, error) {
	if m.Attachment == "" {
		return &waProto.Message{Conversation: proto.String(m.Content)}, nil
	}

	decodedAttachment, err := static.DecodeBase64(m.Attachment)
	if err != nil {
		return nil, err
	}

	mimeType := mimetype.Detect([]byte(decodedAttachment)).String()

	uploadRes, err := w.Client.Upload(
		context.Background(),
		[]byte(decodedAttachment),
		w.GetMediaType(mimeType),
	)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.Contains(mimeType, "image"):
		return &waProto.Message{
			ImageMessage: &waProto.ImageMessage{
				Caption:       proto.String(m.Content),
				Mimetype:      proto.String(mimeType),
				Url:           &uploadRes.URL,
				DirectPath:    &uploadRes.DirectPath,
				MediaKey:      uploadRes.MediaKey,
				FileEncSha256: uploadRes.FileEncSHA256,
				FileSha256:    uploadRes.FileSHA256,
				FileLength:    &uploadRes.FileLength,
			},
		}, nil
	case strings.Contains(mimeType, "audio"):
		return &waProto.Message{
			AudioMessage: &waProto.AudioMessage{
				Mimetype:      proto.String(mimeType),
				Url:           &uploadRes.URL,
				DirectPath:    &uploadRes.DirectPath,
				MediaKey:      uploadRes.MediaKey,
				FileEncSha256: uploadRes.FileEncSHA256,
				FileSha256:    uploadRes.FileSHA256,
				FileLength:    &uploadRes.FileLength,
			},
		}, nil
	case strings.Contains(mimeType, "video"):
		return &waProto.Message{
			VideoMessage: &waProto.VideoMessage{
				Caption:       proto.String(m.Content),
				Mimetype:      proto.String(mimeType),
				Url:           &uploadRes.URL,
				DirectPath:    &uploadRes.DirectPath,
				MediaKey:      uploadRes.MediaKey,
				FileEncSha256: uploadRes.FileEncSHA256,
				FileSha256:    uploadRes.FileSHA256,
				FileLength:    &uploadRes.FileLength,
			},
		}, nil
	case strings.Contains(mimeType, "application"):
		return &waProto.Message{
			DocumentMessage: &waProto.DocumentMessage{
				Mimetype:      proto.String(mimeType),
				Url:           &uploadRes.URL,
				DirectPath:    &uploadRes.DirectPath,
				MediaKey:      uploadRes.MediaKey,
				FileEncSha256: uploadRes.FileEncSHA256,
				FileSha256:    uploadRes.FileSHA256,
				FileLength:    &uploadRes.FileLength,
			},
		}, nil
	}

	return nil, errors.New(static.INTERNAL_SERVER_ERROR)
}

// Gets whatsmeow.MediaType based on the mimeType
func (w *Whatsapp) GetMediaType(mimeType string) whatsmeow.MediaType {
	switch {
	case strings.Contains(mimeType, "image"):
		return whatsmeow.MediaImage
	case strings.Contains(mimeType, "audio"):
		return whatsmeow.MediaAudio
	case strings.Contains(mimeType, "video"):
		return whatsmeow.MediaVideo
	case strings.Contains(mimeType, "application"):
		return whatsmeow.MediaDocument
	}

	return whatsmeow.MediaDocument
}
