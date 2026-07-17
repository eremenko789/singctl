package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var projectEmojiHex = regexp.MustCompile(`(?i)^[0-9a-f]{4,8}$`)

// NormalizeProjectEmoji converts CLI emoji input to lowercase hex for the API.
// Hex 4–8 chars pass through as lowercase; a single non-ASCII rune becomes hex;
// empty, multi-rune, ASCII words, and single ASCII letters/digits are rejected.
func NormalizeProjectEmoji(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("emoji не задан")
	}
	if projectEmojiHex.MatchString(s) {
		return strings.ToLower(s), nil
	}
	if utf8.RuneCountInString(s) != 1 {
		return "", fmt.Errorf("emoji должен быть одним символом или hex (4–8 символов)")
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return "", fmt.Errorf("некорректный emoji")
	}
	if r <= 0x7F {
		return "", fmt.Errorf("emoji должен быть unicode-символом или hex (не ASCII-словом)")
	}
	return strconv.FormatInt(int64(r), 16), nil
}
