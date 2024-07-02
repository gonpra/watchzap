package parser

import (
    "bytes"
    "encoding/json"
    "errors"

    "github.com/rs/zerolog/log"

    "github.com/watchzap/internal/static"
)

func JsonParser(body []byte) (*[]Message, error) {
    var messages []Message

    decodedBody, err := DecodeUTF16(body)
    if err != nil {
        log.Error().Err(err).Msg("WZ: Could not decode body")
        return nil, err
    }

    r := bytes.NewReader(decodedBody)
    decoder := json.NewDecoder(r)
    err = decoder.Decode(&messages)
    if err != nil {
        log.Warn().
            Err(err).
            Str("parser", "json").
            Msg("WZ: The file is not in the specified format. See README for more information")
        return nil, err
    }

    for i, m := range messages {
        if m.Recipient == "" || m.Content == "" {
            log.Error().
                Str("parser", "json").
                Int("index", i).
                Msg("WZ: Recipient or Content fields are empty")
            return nil, errors.New(static.EMPTY_FIELD)
        }
    }

    return &messages, nil
}
