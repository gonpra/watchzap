package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func YamlParser(body []byte) (*[]Message, error) {
	var messages []Message

	stringBody := strings.TrimSpace(string(body))
	fmt.Println(stringBody)
	err := yaml.Unmarshal([]byte(stringBody), &messages)
	if err != nil {
		log.Warn().
			Err(err).
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
