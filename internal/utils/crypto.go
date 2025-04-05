package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

type WireguardKeyPair struct {
	PrivateKey *string
	PublicKey  *string
}

func GenerateNewKeyPair() (*WireguardKeyPair, error) {
	var privateKey [32]byte
	_, err := rand.Read(privateKey[:])
	if err != nil {
		fmt.Println("Error generating private key:", err)
		return nil, err
	}

	// Curve25519 clamping function
	// https://github.com/WireGuard/wireguard-tools/blob/master/src/genkey.c#L75
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	privateKeyBase64 := base64.StdEncoding.EncodeToString(privateKey[:])
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey[:])

	keyPair := &WireguardKeyPair{
		PrivateKey: &privateKeyBase64,
		PublicKey:  &publicKeyBase64,
	}

	return keyPair, nil
}
