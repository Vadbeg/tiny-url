package utils

import "crypto/md5"
import "encoding/hex"

func ShortenUrl(url string, length int) string {
	hash := md5.Sum([]byte(url))
	url_hash := hex.EncodeToString(hash[:])[:length]

	return url_hash
}
