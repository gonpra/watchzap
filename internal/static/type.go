package static

import (
    "strings"
)

// Sets a flag type for []string
type StringArrayVar []string

func (s *StringArrayVar) String() string {
    sa := []string(*s)
    return strings.Join(sa, " ")
}

func (s *StringArrayVar) Get() []string {
    sa := []string(*s)

    return sa
}

func (s *StringArrayVar) Set(value string) error {
    *s = append(*s, value)
    return nil
}
