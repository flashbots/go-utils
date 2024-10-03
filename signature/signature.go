// Package signature provides functionality for interacting with the X-Flashbots-Signature header.
package signature

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// HTTPHeader is the name of the X-Flashbots-Signature header.
const HTTPHeader = "X-Flashbots-Signature"

var (
	ErrNoSignature      = errors.New("no signature provided")
	ErrInvalidSignature = errors.New("invalid signature provided")
)

// Verify takes a X-Flashbots-Signature header and a body and verifies that the signature is valid for the body.
// It returns the signing address if the signature is valid or an error if the signature is invalid.
func Verify(header string, body []byte) (signingAddress string, err error) {
	if header == "" {
		return "", ErrNoSignature
	}

	address, signatureStr, found := strings.Cut(header, ":")
	if !found {
		return "", fmt.Errorf("%w: missing separator", ErrInvalidSignature)
	}

	signature, err := hexutil.Decode(signatureStr)
	if err != nil || len(signature) == 0 {
		return "", fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	if signature[len(signature)-1] >= 27 {
		signature[len(signature)-1] -= 27
	}
	if signature[len(signature)-1] > 1 {
		return "", fmt.Errorf("%w: invalid recovery id", ErrInvalidSignature)
	}

	hashedBody := crypto.Keccak256Hash(body).Hex()
	messageHash := accounts.TextHash([]byte(hashedBody))
	signaturePublicKeyBytes, err := crypto.Ecrecover(messageHash, signature)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	publicKey, err := crypto.UnmarshalPubkey(signaturePublicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}
	signaturePubkey := *publicKey
	signaturePubKeyAddress := crypto.PubkeyToAddress(signaturePubkey).Hex()

	// case-insensitive equality check
	if !strings.EqualFold(signaturePubKeyAddress, address) {
		return "", fmt.Errorf("%w: signing address mismatch", ErrInvalidSignature)
	}

	signatureNoRecoverID := signature[:len(signature)-1] // remove recovery id
	if !crypto.VerifySignature(signaturePublicKeyBytes, messageHash, signatureNoRecoverID) {
		return "", fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	return signaturePubKeyAddress, nil
}

// Create takes a body and a private key and returns a X-Flashbots-Signature header value.
// The header value can be included in a HTTP request to sign the body.
func Create(body []byte, privateKey *ecdsa.PrivateKey) (header string, err error) {
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	signature, err := crypto.Sign(
		accounts.TextHash([]byte(hexutil.Encode(crypto.Keccak256(body)))),
		privateKey,
	)
	if err != nil {
		return "", err
	}
	// add 27 to last byte if its less than 27 to make it compatible with ethereum
	if signature[len(signature)-1] < 27 {
		signature[len(signature)-1] += 27
	}

	header = fmt.Sprintf("%s:%s", address, hexutil.Encode(signature))
	return header, nil
}
