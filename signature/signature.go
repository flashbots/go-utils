// Package signature provides functionality for interacting with the X-Flashbots-Signature header.
package signature

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
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
func Verify(header string, body []byte) (common.Address, error) {
	if header == "" {
		return common.Address{}, ErrNoSignature
	}

	parsedSignerStr, parsedSignatureStr, found := strings.Cut(header, ":")
	if !found {
		return common.Address{}, fmt.Errorf("%w: missing separator", ErrInvalidSignature)
	}

	parsedSignature, err := hexutil.Decode(parsedSignatureStr)
	if err != nil || len(parsedSignature) == 0 {
		return common.Address{}, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	if parsedSignature[len(parsedSignature)-1] >= 27 {
		parsedSignature[len(parsedSignature)-1] -= 27
	}
	if parsedSignature[len(parsedSignature)-1] > 1 {
		return common.Address{}, fmt.Errorf("%w: invalid recovery id", ErrInvalidSignature)
	}

	hashedBody := crypto.Keccak256Hash(body).Hex()
	messageHash := accounts.TextHash([]byte(hashedBody))
	recoveredPublicKeyBytes, err := crypto.Ecrecover(messageHash, parsedSignature)
	if err != nil {
		return common.Address{}, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	recoveredPublicKey, err := crypto.UnmarshalPubkey(recoveredPublicKeyBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}
	recoveredSigner := crypto.PubkeyToAddress(*recoveredPublicKey)

	// case-insensitive equality check
	parsedSigner := common.HexToAddress(parsedSignerStr)
	if recoveredSigner.Cmp(parsedSigner) != 0 {
		return common.Address{}, fmt.Errorf("%w: signing address mismatch", ErrInvalidSignature)
	}

	signatureNoRecoverID := parsedSignature[:len(parsedSignature)-1] // remove recovery id
	if !crypto.VerifySignature(recoveredPublicKeyBytes, messageHash, signatureNoRecoverID) {
		return common.Address{}, fmt.Errorf("%w: %w", ErrInvalidSignature, err)
	}

	return recoveredSigner, nil
}

// Create takes a body and a private key and returns a X-Flashbots-Signature header value.
// The header value can be included in a HTTP request to sign the body.
func Create(body []byte, privateKey *ecdsa.PrivateKey) (header string, err error) {
	signer := crypto.PubkeyToAddress(privateKey.PublicKey)
	signature, err := crypto.Sign(
		accounts.TextHash([]byte(hexutil.Encode(crypto.Keccak256(body)))),
		privateKey,
	)
	if err != nil {
		return "", err
	}
	// To maintain compatibility with the EVM `ecrecover` precompile, the recovery ID in the last
	// byte is encoded as v = 27/28 instead of 0/1.
	//
	// See:
	//   - Yellow Paper, Appendix E & F. https://ethereum.github.io/yellowpaper/paper.pdf
	//   - https://www.evm.codes/precompiled (ecrecover is the 1st precompile at 0x01)
	//
	if signature[len(signature)-1] < 27 {
		signature[len(signature)-1] += 27
	}

	header = fmt.Sprintf("%s:%s", signer, hexutil.Encode(signature))
	return header, nil
}
