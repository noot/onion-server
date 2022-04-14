// Code based off gist: https://gist.github.com/wybiral/8f737644fc140c97b6b26c13b1409837

package main

import (
	"crypto/rand"
	"encoding/base32"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
	"strings"
)

// Hidden service version
const version = byte(0x03)

// Salt used to create checkdigits
const salt = ".onion checksum"

// Generate returns an address corresponding to the given private key (without the .onion)
func GenerateAddress() (string, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", nil, err
	}

	return getServiceID(pub), priv, nil
}

func getCheckdigits(pub ed25519.PublicKey) []byte {
	// Calculate checksum sha3(".onion checksum" || publicKey || version)
	checkstr := []byte(salt)
	checkstr = append(checkstr, pub...)
	checkstr = append(checkstr, version)
	checksum := sha3.Sum256(checkstr)
	return checksum[:2]
}

func getServiceID(pub ed25519.PublicKey) string {
	// Construct onion address base32(publicKey || checkdigits || version)
	checkdigits := getCheckdigits(pub)
	combined := pub[:]
	combined = append(combined, checkdigits...)
	combined = append(combined, version)
	serviceID := base32.StdEncoding.EncodeToString(combined)
	return strings.ToLower(serviceID)
}
