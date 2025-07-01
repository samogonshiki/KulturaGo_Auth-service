package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

func FileName(uid int64, email string) string {
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(email))))
	return fmt.Sprintf("avatars/%d_%s.jpg", uid, hex.EncodeToString(sum[:4]))
}
