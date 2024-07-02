package prompt

import (
    "fmt"

    "github.com/manifoldco/promptui"
    "github.com/rs/zerolog/log"

    "github.com/watchzap/internal/static"
)

func YesNo(prefix string) bool {
    prompt := promptui.Select{
        Label: fmt.Sprintf("%s [Yes/No]", prefix),
        Items: []string{"Yes", "No"},
    }

    _, result, err := prompt.Run()
    if err != nil {
        log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
    }

    return result == "Yes"
}

func Select(label string, items []string) string {
    prompt := promptui.Select{
        Label: label,
        Items: items,
    }

    _, result, err := prompt.Run()
    if err != nil {
        log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
    }

    return result
}

func Input(label string, validate func(i string) error) string {
    prompt := promptui.Prompt{
        Label:    label,
        Validate: validate,
    }

    result, err := prompt.Run()
    if err != nil {
        log.Fatal().Err(err).Msg(static.INTERNAL_SERVER_ERROR)
    }

    return result
}
