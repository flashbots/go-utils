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

type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
	hexAddress string
}

func NewSigner(privateKey *ecdsa.PrivateKey) Signer {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return Signer{
		privateKey: privateKey,
		hexAddress: address.Hex(),
		address:    address,
	}
}

func NewRandomSigner() (*Signer, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	signer := NewSigner(privateKey)
	return &signer, nil
}

func (s *Signer) Address() common.Address {
	return s.address
}

// Create takes a body and a private key and returns a X-Flashbots-Signature header value.
// The header value can be included in a HTTP request to sign the body.
func (s *Signer) Create(body []byte) (string, error) {
	signature, err := crypto.Sign(
		accounts.TextHash([]byte(hexutil.Encode(crypto.Keccak256(body)))),
		s.privateKey,
	)
	if err != nil {
		return "", err
	}
	// To maintain compatibility with the EVM `ecrecover` precompile, the recovery ID in the last
	// byte is encoded as v = 27/28 instead of 0/1.  This also ensures we generate the same signatures as other
	// popular libraries like ethers.js, and tooling like `cast wallet sign` and MetaMask.
	//
	// See:
	//   - Yellow Paper, Appendix E & F. https://ethereum.github.io/yellowpaper/paper.pdf
	//   - https://www.evm.codes/precompiled (ecrecover is the 1st precompile at 0x01)
	//
	if signature[len(signature)-1] < 27 {
		signature[len(signature)-1] += 27
	}

	header := fmt.Sprintf("%s:%s", s.hexAddress, hexutil.Encode(signature))
	return header, nil
}
