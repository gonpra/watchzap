package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/radovskyb/watcher"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mau.fi/whatsmeow/types"

	"github.com/watchzap/internal/api"
	"github.com/watchzap/internal/parser"
	"github.com/watchzap/internal/prompt"
	"github.com/watchzap/internal/static"
)

// Msa stands for shortcut for map[string]any
type msa map[string]any

var (
	debug        bool
	removeOnSend bool
	folder       string
	port         string
)

type MessageRequest struct {
	Jid  types.JID
	Flag bool
}

// Parses messages based on their content type
func parse(ext string, body []byte) (*[]parser.Message, error) {
	if strings.HasSuffix(ext, "json") {
		return parser.JsonParser(body)
	}

	if strings.HasSuffix(ext, "yaml") {
		return parser.YamlParser(body)
	}

	return nil, errors.New(static.NO_PARSER_FOUND)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flag.BoolVar(&debug, "debug", false, "enables the debug mode for WhatsApp API")
	flag.BoolVar(&removeOnSend, "removeOnSend", false, "deletes the file after sending the message")
	flag.Parse()

	db, err := sql.Open("sqlite3", "file:zap.db?_foreign_keys=on")
	if err != nil {
		log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
	}

	whatsapp, err := api.NewWhatsapp(debug)
	if err != nil {
		log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
	}

	err = whatsapp.Login()
	if err != nil {
		log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
	}

	runResult := prompt.Select(
		"Select ",
		[]string{"Watch Folder", "Enable HTTP Server", "Both", "Logout"},
	)
	switch runResult {
	case "Watch Folder":
		folder = prompt.Input("What folder will you watch", nil)
		watch(whatsapp)
	case "Enable HTTP Server":
		port = prompt.Input("What port will you listen", nil)
		httpServe(whatsapp)
	case "Both":
		folder = prompt.Input("What folder will you watch", nil)
		port = prompt.Input("What port will you listen", nil)
		go watch(whatsapp)
		httpServe(whatsapp)
	case "Logout":
		whatsapp.Client.Logout()
		db.Exec(static.WIPE_DB)
		restart()
	}
}

// Sets up an HTTP server for receiving message requests
func httpServe(whatsapp *api.Whatsapp) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Error().Err(err).Msg("WZ: Error reading request body")

			jsonR, _ := json.Marshal(msa{"status": "error", "error": err.Error()})
			w.Write(jsonR)
			return
		}
		defer r.Body.Close()

		messages, err := parse(r.Header.Get("Content-Type"), body)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			log.Error().Err(err).Msg("WZ: Error parsing file")

			jsonR, _ := json.Marshal(msa{"status": "error", "error": err.Error()})
			w.Write(jsonR)
			return
		}

		err = sendMessages(messages, whatsapp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error().Err(err).Msg("WZ: Error sending messages")

			jsonR, _ := json.Marshal(msa{"status": "error", "error": err.Error()})
			w.Write(jsonR)
			return
		}

		jsonR, _ := json.Marshal(msa{"status": "ok", "amount": len(*messages)})

		w.WriteHeader(http.StatusCreated)
		w.Write(jsonR)
	})
	log.Info().Str("function", "http").Msg("WZ: Serving HTTP server at " + port)
	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal().Err(err).Msg("WZ: Failed to serve HTTP server")
	}
}

// Sets up a file watcher to monitor changes in a directory
func watch(whatsapp *api.Whatsapp) {
	time.Sleep(time.Millisecond * 100)

	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Move, watcher.Write, watcher.Rename)
	go func() {
		for {
			select {
			case event := <-w.Event:
				go doEvent(event, whatsapp)
			case err := <-w.Error:
				log.Fatal().Err(err).Str("function", "watch").Msg("WZ: Failed getting folder event")
			case <-w.Closed:
				return
			}
		}
	}()
	if err := w.Add(folder); err != nil {
		log.Fatal().Err(err).Str("function", "watch").Msg("WZ: Error adding folder to watch")
	}
	log.Info().Str("folder", folder).Msg("WZ: Watching folder every 100ms")
	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatal().
			Err(err).
			Str("function", "watch").
			Msg(fmt.Sprintf("Failed to watch folder %s", folder))
	}
}

// / Processes file events triggered by the watch function
func doEvent(w watcher.Event, whatsapp *api.Whatsapp) {
	if w.IsDir() {
		return
	}

	f, err := os.Open(w.Path)
	if err != nil {
		log.Error().Err(err).Msg("WZ: Failed opening file")
	}

	stat, err := f.Stat()
	if err != nil {
		log.Error().Err(err).Msg("WZ: Failed getting stat of file")
		return
	}
	if stat.Size() == 0 {
		log.Warn().Str("file", w.Name()).Str("path", w.Path).Msg("WZ: File content is empty")
		return
	}

	body := make([]byte, stat.Size())
	_, err = f.Read(body)
	if err != nil {
		log.Error().Err(err).Str("parser", "json").Msg("WZ: Error reading file")
		return
	}

	messages, err := parse(filepath.Ext(w.Path), body)
	if err != nil {
		log.Error().Err(err).Msg("WZ: Error parsing messages")
		return
	}

	sendMessages(messages, whatsapp)

	err = f.Close()
	if err != nil {
		log.Error().Err(err).Msg("WZ: Could not close file")
		return
	}

	if removeOnSend {
		err := os.Remove(w.Path)
		if err != nil {
			log.Warn().Err(err).Msg("WZ: Could not delete the file upon send")
			return
		}
	}
}

// Sends messages to recipients based on parsed messages
func sendMessages(messages *[]parser.Message, whatsapp *api.Whatsapp) error {
	for _, m := range *messages {
		var req MessageRequest

		contacts, err := whatsapp.Client.Store.Contacts.GetAllContacts()
		if err != nil {
			log.Error().Err(err).Msg("WZ: Failed getting contacts")
			return err
		}
		groups, err := whatsapp.Client.GetJoinedGroups()
		if err != nil {
			log.Error().Err(err).Msg("WZ: Failed to get joined groups")
			return err
		}

		for _, g := range groups {
			if m.Recipient == g.GroupName.Name {
				req.Jid = g.JID
				req.Flag = true
			}
		}
		for j, c := range contacts {
			if m.Recipient == c.PushName || m.Recipient == c.FullName {
				req.Jid = j
				req.Flag = true
			}
		}

		if req.Flag {
			sendMessage, err := whatsapp.GenerateMessage(m)
			if err != nil {
				return err
			}

			_, err = whatsapp.Client.SendMessage(context.Background(), req.Jid, sendMessage)
			if err != nil {
				log.Error().Err(err).Msg("WZ: Error sending message to recipient")
				return err
			}
			log.Info().
				Str("recipient", m.Recipient).
				Str("content", m.Content).
				Msg("WZ: Sent message successfully")
		} else {
			log.Info().Msg("WZ: Recipient was not found")
		}
	}

	return nil
}

// Restart go program execution
func restart() {
	self, err := os.Executable()
	if err != nil {
		log.Fatal().Err(err)
	}
	args := os.Args
	env := os.Environ()

	// Windows doesn't support exec syscalls :(
	if runtime.GOOS == "windows" {
		cmd := exec.Command(self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env

		err := cmd.Run()
		if err != nil {
			log.Fatal().Err(err)
		}

		os.Exit(0)
	}

	syscall.Exec(self, args, env)
}
