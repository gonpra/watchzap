package static

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/hex"
)

func EncodeBase64(value []byte) (string, error) {
    data := base64.StdEncoding.EncodeToString(value)

    return string(data), nil
}

func DecodeBase64(b64 string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(b64)
    if err != nil {
        return "", err
    }

    return string(data), nil
}

func hash(key string) string {
    hasher := md5.New()
    hasher.Write([]byte(key))
    return hex.EncodeToString(hasher.Sum(nil))
}
