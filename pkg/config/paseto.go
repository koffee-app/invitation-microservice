package config

import (
	"encoding/hex"
	"os"

	"golang.org/x/crypto/ed25519"
)

var pasetoPublic ed25519.PublicKey
var pasetoPrivate ed25519.PrivateKey

func PasetoInit() {
	key, ok := os.LookupEnv("PASETO_PUBLIC_KEY")
	key2, ok2 := os.LookupEnv("PASETO_PRIVATE_KEY")

	if !ok || !ok2 {
		panic("Paseto Key Public not retrieved correctly...")
	}

	b, _ := hex.DecodeString(key)
	pasetoPublic = ed25519.PublicKey(b)
	b, _ = hex.DecodeString(key2)
	pasetoPrivate = ed25519.PrivateKey(b)
}

// PrivateKeyParsed  Returns the parsed key for use in paseto
func PrivateKeyParsed() ed25519.PrivateKey {
	return pasetoPrivate
}

// PublicKeyParsed Returns the parsed key for use in paseto
func PublicKeyParsed() ed25519.PublicKey {
	return pasetoPublic
}
