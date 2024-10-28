package signature_test

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flashbots/go-utils/signature"
	"github.com/stretchr/testify/require"
)

// TestSignatureVerify tests the signature verification function.
func TestSignatureVerify(t *testing.T) {
	// For most of these test cases, we first need to generate a signature
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	signerAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	body := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["%s","pending"],"id":1}`,
		signerAddress,
	)

	sig, err := crypto.Sign(
		accounts.TextHash([]byte(hexutil.Encode(crypto.Keccak256([]byte(body))))),
		privateKey,
	)
	require.NoError(t, err)

	header := fmt.Sprintf("%s:%s", signerAddress, hexutil.Encode(sig))

	t.Run("header is empty", func(t *testing.T) {
		_, err := signature.Verify("", []byte{})
		require.ErrorIs(t, err, signature.ErrNoSignature)
	})

	t.Run("header is valid", func(t *testing.T) {
		verifiedAddress, err := signature.Verify(header, []byte(body))
		require.NoError(t, err)
		require.Equal(t, signerAddress, verifiedAddress)
	})

	t.Run("header is invalid", func(t *testing.T) {
		_, err := signature.Verify("invalid", []byte(body))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("header has extra bytes", func(t *testing.T) {
		_, err := signature.Verify(header+"deadbeef", []byte(body))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("header has missing bytes", func(t *testing.T) {
		_, err := signature.Verify(header[:len(header)-8], []byte(body))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("body is empty", func(t *testing.T) {
		_, err := signature.Verify(header, []byte{})
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("body is invalid", func(t *testing.T) {
		_, err := signature.Verify(header, []byte(`{}`))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("body has extra bytes", func(t *testing.T) {
		_, err := signature.Verify(header, []byte(body+"..."))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})

	t.Run("body has missing bytes", func(t *testing.T) {
		_, err := signature.Verify(header, []byte(body[:len(body)-8]))
		require.ErrorIs(t, err, signature.ErrInvalidSignature)
	})
}

// TestVerifySignatureFromMetaMask ensures that a signature generated by MetaMask
// can be verified by this package.
func TestVerifySignatureFromMetaMask(t *testing.T) {
	// Source: use the "Sign Message" feature in Etherscan
	// to sign the keccak256 hash of `Hello`
	// Published to https://etherscan.io/verifySig/255560
	messageHash := crypto.Keccak256Hash([]byte("Hello")).Hex()
	require.Equal(t, `0x06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2`, messageHash)
	signerAddress := common.HexToAddress(`0x4bE0Cd2553356b4ABb8b6a1882325dAbC8d3013D`)
	signatureHash := `0xbf36915334f8fa93894cd54d491c31a89dbf917e9a4402b2779b73d21ecf46e36ff07db2bef6d10e92c99a02c1c5ea700b0b674dfa5d3ce9220822a7ebcc17101b`
	header := signerAddress.Hex() + ":" + signatureHash
	verifiedAddress, err := signature.Verify(
		header,
		[]byte(`Hello`),
	)
	require.NoError(t, err)
	require.Equal(t, signerAddress, verifiedAddress)
}

// TestVerifySignatureFromCast ensures that the signature generated by the `cast` CLI
// can be verified by this package.
func TestVerifySignatureFromCast(t *testing.T) {
	// Source: use `cast wallet sign` in the `cast` CLI
	// to sign the keccak256 hash of `Hello`:
	// `cast wallet sign --interactive $(cast from-utf8 $(cast keccak Hello))`
	// NOTE: The call to from-utf8 is required as cast wallet sign
	// interprets inputs with a leading 0x as a byte array, not a string.
	// Published to https://etherscan.io/verifySig/255562
	messageHash := crypto.Keccak256Hash([]byte("Hello")).Hex()
	require.Equal(t, `0x06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2`, messageHash)
	signerAddress := common.HexToAddress(`0x2485Aaa7C5453e04658378358f5E028150Dc7606`)
	signatureHash := `0xff2aa92eb8d8c2ca04f1755a4ddbff4bda6a5c9cefc8b706d5d8a21d3aa6fe7a20d3ec062fb5a4c1656fd2c14a8b33ca378b830d9b6168589bfee658e83745cc1b`
	header := signerAddress.Hex() + ":" + signatureHash
	verifiedAddress, err := signature.Verify(
		header,
		[]byte(`Hello`),
	)
	require.NoError(t, err)
	require.Equal(t, signerAddress, verifiedAddress)
}

// TestSignatureCreateAndVerify uses a randomly generated private key
// to create a signature and then verifies it.
func TestSignatureCreateAndVerify(t *testing.T) {
	signer, err := signature.NewRandomSigner()
	require.NoError(t, err)

	signerAddress := signer.Address()
	body := fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["%s","pending"],"id":1}`,
		signerAddress,
	)

	header, err := signer.Create([]byte(body))
	require.NoError(t, err)

	verifiedAddress, err := signature.Verify(header, []byte(body))
	require.NoError(t, err)
	require.Equal(t, signerAddress, verifiedAddress)
}

// TestSignatureCreateCompareToCastAndEthers uses a static private key
// and compares the signature generated by this package to the signatures
// generated by the `cast` CLI and ethers.js.
func TestSignatureCreateCompareToCastAndEthers(t *testing.T) {
	// This purposefully uses the already highly compromised keypair from the go-ethereum book:
	// https://goethereumbook.org/transfer-eth/
	// privateKey = fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19
	signer, err := signature.NewSignerFromHexPrivateKey("0xfad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	require.NoError(t, err)

	address := signer.Address()
	body := []byte("Hello")

	// I generated the signature using the cast CLI:
	//
	// 	cast wallet sign --private-key fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19  $(cast from-utf8 $(cast keccak Hello))
	//
	// (As mentioned above, the call to from-utf8 is required as cast wallet
	//  sign interprets inputs with a leading 0x as a byte array, not a string.)
	//
	// As well as the following ethers script:
	//
	// import { Wallet } from "ethers";
	// import { id } from 'ethers/lib/utils'
	// var w = new Wallet("0xfad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	// `${ await w.getAddress() }:${ await w.signMessage(id("Hello")) }`
	//'0x96216849c49358B10257cb55b28eA603c874b05E:0x1446053488f02d460c012c84c4091cd5054d98c6cfca01b65f6c1a72773e80e60b8a4931aeee7ed18ce3adb45b2107e8c59e25556c1f871a8334e30e5bddbed21c'

	expectedSignature := "0x1446053488f02d460c012c84c4091cd5054d98c6cfca01b65f6c1a72773e80e60b8a4931aeee7ed18ce3adb45b2107e8c59e25556c1f871a8334e30e5bddbed21c"
	expectedAddress := common.HexToAddress("0x96216849c49358B10257cb55b28eA603c874b05E")
	expectedHeader := fmt.Sprintf("%s:%s", expectedAddress, expectedSignature)
	require.Equal(t, expectedAddress, address)

	header, err := signer.Create(body)
	require.NoError(t, err)
	require.Equal(t, expectedHeader, header)

	verifiedAddress, err := signature.Verify(header, body)
	require.NoError(t, err)
	require.Equal(t, expectedAddress, verifiedAddress)
}

func BenchmarkSignatureCreate(b *testing.B) {
	signer, err := signature.NewRandomSigner()
	require.NoError(b, err)

	body := []byte("Hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := signer.Create(body)
		require.NoError(b, err)
	}
}

// benchmark signature verification
func BenchmarkSignatureVerify(b *testing.B) {
	signer, err := signature.NewRandomSigner()
	require.NoError(b, err)

	body := "body"
	header, err := signer.Create([]byte(body))
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := signature.Verify(header, []byte(body))
		require.NoError(b, err)
	}
}
