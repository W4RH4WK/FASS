package fass

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"io"
)

// Token that is given to a user, it is commonly used for authenticating
// uploads.
type Token = string

// GenerateToken generates a fixed length token.
func GenerateToken() (Token, error) {
	const length = 25

	token := make([]byte, length)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	return base32.StdEncoding.EncodeToString(token), nil
}

// TokenMapping relates tokens to mail addresses. Be careful when handling since
// this is personal identifiable data.
type TokenMapping map[Token]Mail

// NewTokenMapping creates a TokenMapping from the given io.Reader (typically a
// file). The input is expected to contain one mail address per line. Tokens are
// generated on the fly. A token is guaranteed to be unique within the returned
// mapping.
func NewTokenMapping(mailAddresses io.Reader) (mapping TokenMapping, err error) {
	mapping = make(TokenMapping)

	scanner := bufio.NewScanner(mailAddresses)
	for scanner.Scan() {
		var token Token

		// generate token that's not already in use
		for {
			token, err = GenerateToken()
			if err != nil {
				return
			}

			if _, ok := mapping[token]; !ok {
				break
			}
		}

		mapping[token] = scanner.Text()
	}

	return
}
