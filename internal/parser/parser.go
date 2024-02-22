package parser

import (
	"bytes"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/watchzap/internal/static"
)

type Message struct {
	Recipient  string `json:"recipient"`
	Content    string `json:"content"`
	Attachment string `json:"attachment"`
}

func DecodeUTF16(b []byte) ([]byte, error) {
	if len(b)%2 != 0 {
		return []byte{}, fmt.Errorf(static.INVALID_BYTES)
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	// Remove the Byte Order Mark (BOM). The BOM identifies that the text is UTF-8 encoded, but it should be removed before decoding
	// For more info see here: https://stackoverflow.com/questions/31398044/got-error-invalid-character-%c3%af-looking-for-beginning-of-value-from-json-unmar#31399046
	body := ret.Bytes()
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	return body, nil
}
