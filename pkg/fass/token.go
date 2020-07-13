package fass

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"io"
	"regexp"
)

// Token that is given to a user, it is commonly used for authenticating
// uploads.
type Token = string

var tokenFormat = regexp.MustCompile(`^[0-9A-Za-z=]{0,64}$`)

// TokenHasValidFormat verifies that the given token follows the required format.
func TokenHasValidFormat(t Token) bool {
	return tokenFormat.Match([]byte(t))
}

// GenerateToken generates a fixed length token.
func GenerateToken() Token {
	const length = 25

	tokenData := make([]byte, length)
	_, err := rand.Read(tokenData)
	if err != nil {
		panic(err)
	}

	token := base32.StdEncoding.EncodeToString(tokenData)
	if !TokenHasValidFormat(token) {
		panic("generated token has invalid format")
	}

	return token
}

// TokenMapping relates tokens to mail addresses. Be careful when handling since
// this is personal identifiable data.
type TokenMapping map[Token]Mail

// Store stores the mapping as file.
func (m TokenMapping) Store(filepath string) error {
	return marshalToFile(filepath, m)
}

// NewTokenMapping creates a TokenMapping from the given io.Reader (typically a
// file). The input is expected to contain one mail address per line. Tokens are
// generated on the fly. A token is guaranteed to be unique within the returned
// mapping.
func NewTokenMapping(mailAddresses io.Reader) (mapping TokenMapping) {
	mapping = make(TokenMapping)

	scanner := bufio.NewScanner(mailAddresses)
	for scanner.Scan() {
		var token Token

		// generate token that's not already in use
		for {
			token = GenerateToken()

			if _, ok := mapping[token]; !ok {
				break
			}
		}

		mapping[token] = scanner.Text()
	}

	return
}

func LoadTokenMapping(mappingFilepath string) (TokenMapping, error) {
	mapping := make(TokenMapping)
	err := unmarshalFromFile(mappingFilepath, &mapping)
	return mapping, err
}
