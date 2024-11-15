package main

// This example demonstrates sending a signed eth_sendRawTransaction request to a
// multioperator builder node with a specific server certificate.

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/flashbots/go-utils/rpcclient"
	"github.com/flashbots/go-utils/rpctypes"
	"github.com/flashbots/go-utils/signature"
)

var (
	// Builder node endpoint and certificate
	endpoint = "https://127.0.0.1:443"
	certPEM  = []byte("-----BEGIN CERTIFICATE-----\nMIIBlTCCATugAwIBAgIQeUQhWmrcFUOKnA/HpBPdODAKBggqhkjOPQQDAjAPMQ0w\nCwYDVQQKEwRBY21lMB4XDTI0MTExNDEyMTExM1oXDTI1MTExNDEyMTExM1owDzEN\nMAsGA1UEChMEQWNtZTBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABJCl4R+DtNqu\nyPYd8a+Ppd4lSIEgKcyGz3Q6HOnZV3D96oxW03e92FBdKUkl5DLxTYo+837u44XL\n11OWmajjKzGjeTB3MA4GA1UdDwEB/wQEAwIChDATBgNVHSUEDDAKBggrBgEFBQcD\nATAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQWBBTjt0S4lYkceJnonMJBEvwjezh3\nvDAgBgNVHREEGTAXgglsb2NhbGhvc3SHBDSuK4SHBH8AAAEwCgYIKoZIzj0EAwID\nSAAwRQIgOzm8ghnR4cKiE76siQ43Q4H2RzoJUmww3NyRVFkcp6oCIQDFZmuI+2tK\n1WlX3whjllaqr33K7kAa9ntihWfo+VB9zg==\n-----END CERTIFICATE-----\n")

	// Transaction and signing key
	rawTxHex         = "0x02f8710183195414808503a1e38a30825208947804a60641a89c9c3a31ab5abea2a18c2b6b48408788c225841b2a9f80c080a0df68a9664190a59005ab6d6cc6b8e5a1e25604f546c36da0fd26ddd44d8f7d50a05b1bcfab22a3017cabb305884d081171e0f23340ae2a13c04eb3b0dd720a0552"
	signerPrivateKey = "0xaccc869c5c3cb397e4833d41b138d3528af6cc5ff4808bb85a1c2ce1c8f04007"
)

func createTransportForSelfSignedCert(certPEM []byte) (*http.Transport, error) {
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(certPEM); !ok {
		return nil, errors.New("failed to add certifcate to pool")
	}
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    certPool,
			MinVersion: tls.VersionTLS12,
		},
	}, nil
}

func exampleSendRawTx() error {
	requestSigner, err := signature.NewSignerFromHexPrivateKey(signerPrivateKey)
	if err != nil {
		return err
	}

	transport, err := createTransportForSelfSignedCert(certPEM)
	if err != nil {
		return err
	}

	client := rpcclient.NewClientWithOpts(endpoint, &rpcclient.RPCClientOpts{
		HTTPClient: &http.Client{
			Transport: transport,
		},
		Signer: requestSigner,
	})

	rawTransaction := hexutil.MustDecode(rawTxHex)
	resp, err := client.Call(context.Background(), "eth_sendRawTransaction", rpctypes.EthSendRawTransactionArgs(rawTransaction))
	if err != nil {
		return err
	}
	if resp != nil && resp.Error != nil {
		return fmt.Errorf("rpc error: %s", resp.Error.Error())
	}
	return nil
}

func main() {
	err := exampleSendRawTx()
	if err != nil {
		panic(err)
	}
}
