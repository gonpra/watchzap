package parser

import (
    "errors"

    "github.com/rs/zerolog/log"
    "gopkg.in/yaml.v3"
)

func YamlParser(body []byte) (*[]Message, error) {
    var messages []Message

    err := yaml.Unmarshal(body, &messages)
    if err != nil {
        log.Warn().
            Str("parser", "yaml").
            Msg("WZ: The file is not in the specified format. See README for more information")
        return nil, err
    }

    for i, m := range messages {
        if m.Recipient == "" || m.Content == "" {
            log.Warn().
                Str("parser", "yaml").
                Int("index", i).
                Msg("WZ: Recipient or Content fields are empty")
            return nil, errors.New("empty field")
        }
    }

    return &messages, nil
}
