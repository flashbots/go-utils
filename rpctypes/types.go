// Package rpctypes implement types commonly used in the Flashbots codebase for receiving and senging requests
package rpctypes

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"
)

// Note on optional Signer field:
// * when receiving from Flashbots or other builders this field should be set
// * otherwise its set from the request signature by orderflow proxy
//   in this case it can be empty! @should we prohibit that?

// eth_SendBundle

const (
	BundleTxLimit     = 100
	MevBundleTxLimit  = 50
	MevBundleMaxDepth = 1
)

var (
	ErrBundleNoTxs          = errors.New("bundle with no txs")
	ErrBundleTooManyTxs     = errors.New("too many txs in bundle")
	ErrMevBundleUnmatchedTx = errors.New("mev bundle with unmatched tx")
	ErrMevBundleTooDeep     = errors.New("mev bundle too deep")
)

type EthSendBundleArgs struct {
	Txs               []hexutil.Bytes `json:"txs"`
	BlockNumber       hexutil.Uint64  `json:"blockNumber"`
	MinTimestamp      *uint64         `json:"minTimestamp,omitempty"`
	MaxTimestamp      *uint64         `json:"maxTimestamp,omitempty"`
	RevertingTxHashes []common.Hash   `json:"revertingTxHashes,omitempty"`
	ReplacementUUID   *string         `json:"replacementUuid,omitempty"`

	// fields available only when receiving from the Flashbots or other builders, not users
	ReplacementNonce *uint64         `json:"replacementNonce,omitempty"`
	SigningAddress   *common.Address `json:"signingAddress,omitempty"`

	DroppingTxHashes []common.Hash   `json:"droppingTxHashes,omitempty"` // not supported (from beaverbuild)
	UUID             *string         `json:"uuid,omitempty"`             // not supported (from beaverbuild)
	RefundPercent    *uint64         `json:"refundPercent,omitempty"`    // not supported (from beaverbuild)
	RefundRecipient  *common.Address `json:"refundRecipient,omitempty"`  // not supported (from beaverbuild)
	RefundTxHashes   []string        `json:"refundTxHashes,omitempty"`   // not supported (from titanbuilder)
}

// mev_sendBundle

const (
	RevertModeAllow = "allow"
	RevertModeDrop  = "drop"
	RevertModeFail  = "fail"
)

type MevBundleInclusion struct {
	BlockNumber hexutil.Uint64 `json:"block"`
	MaxBlock    hexutil.Uint64 `json:"maxBlock"`
}

type MevBundleBody struct {
	Hash       *common.Hash       `json:"hash,omitempty"`
	Tx         *hexutil.Bytes     `json:"tx,omitempty"`
	Bundle     *MevSendBundleArgs `json:"bundle,omitempty"`
	CanRevert  bool               `json:"canRevert,omitempty"`
	RevertMode string             `json:"revertMode,omitempty"`
}

type MevBundleValidity struct {
	Refund       []RefundConstraint `json:"refund,omitempty"`
	RefundConfig []RefundConfig     `json:"refundConfig,omitempty"`
}

type RefundConstraint struct {
	BodyIdx int `json:"bodyIdx"`
	Percent int `json:"percent"`
}

type RefundConfig struct {
	Address common.Address `json:"address"`
	Percent int            `json:"percent"`
}

type MevBundleMetadata struct {
	// Signer should be set by infra that verifies user signatures and not user
	Signer           *common.Address `json:"signer,omitempty"`
	ReplacementNonce *int            `json:"replacementNonce,omitempty"`
	// Used for cancelling. When true the only thing we care about is signer,replacement_nonce and RawShareBundle::replacement_uuid
	Cancelled *bool `json:"cancelled,omitempty"`
}

type MevSendBundleArgs struct {
	Version         string             `json:"version"`
	ReplacementUUID string             `json:"replacementUuid,omitempty"`
	Inclusion       MevBundleInclusion `json:"inclusion"`
	// when empty its considered cancel
	Body     []MevBundleBody    `json:"body"`
	Validity MevBundleValidity  `json:"validity"`
	Metadata *MevBundleMetadata `json:"metadata,omitempty"`

	// must be empty
	Privacy *json.RawMessage `json:"privacy,omitempty"`
}

// eth_sendRawTransaction

type EthSendRawTransactionArgs hexutil.Bytes

func (tx EthSendRawTransactionArgs) MarshalText() ([]byte, error) {
	return hexutil.Bytes(tx).MarshalText()
}

func (tx *EthSendRawTransactionArgs) UnmarshalJSON(input []byte) error {
	return (*hexutil.Bytes)(tx).UnmarshalJSON(input)
}

func (tx *EthSendRawTransactionArgs) UnmarshalText(input []byte) error {
	return (*hexutil.Bytes)(tx).UnmarshalText(input)
}

// eth_cancelBundle

type EthCancelBundleArgs struct {
	ReplacementUUID string          `json:"replacementUuid"`
	SigningAddress  *common.Address `json:"signingAddress"`
}

// bid_subsidiseBlock

type BidSubsisideBlockArgs uint64

/// unique key
/// unique key is used to deduplicate requests, its will give different results then bundle uuid

func newHash() hash.Hash {
	return sha256.New()
}

func uuidFromHash(h hash.Hash) uuid.UUID {
	version := 5
	s := h.Sum(nil)
	var uuid uuid.UUID
	copy(uuid[:], s)
	uuid[6] = (uuid[6] & 0x0f) | uint8((version&0xf)<<4)
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // RFC 4122 variant
	return uuid
}

func (b *EthSendBundleArgs) UniqueKey() uuid.UUID {
	hash := newHash()
	_ = binary.Write(hash, binary.LittleEndian, b.BlockNumber)
	for _, tx := range b.Txs {
		_, _ = hash.Write(tx)
	}
	if b.MinTimestamp != nil {
		_ = binary.Write(hash, binary.LittleEndian, b.MinTimestamp)
	}
	if b.MaxTimestamp != nil {
		_ = binary.Write(hash, binary.LittleEndian, b.MaxTimestamp)
	}
	sort.Slice(b.RevertingTxHashes, func(i, j int) bool {
		return bytes.Compare(b.RevertingTxHashes[i][:], b.RevertingTxHashes[j][:]) <= 0
	})
	for _, txHash := range b.RevertingTxHashes {
		_, _ = hash.Write(txHash.Bytes())
	}
	if b.ReplacementUUID != nil {
		_, _ = hash.Write([]byte(*b.ReplacementUUID))
	}
	if b.ReplacementNonce != nil {
		_ = binary.Write(hash, binary.LittleEndian, *b.ReplacementNonce)
	}

	if b.SigningAddress != nil {
		_, _ = hash.Write(b.SigningAddress.Bytes())
	}
	return uuidFromHash(hash)
}

func (b *EthSendBundleArgs) Validate() (common.Hash, uuid.UUID, error) {
	if len(b.Txs) > BundleTxLimit {
		return common.Hash{}, uuid.Nil, ErrBundleTooManyTxs
	}
	// first compute keccak hash over the txs
	hasher := sha3.NewLegacyKeccak256()
	for _, rawTx := range b.Txs {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(rawTx); err != nil {
			return common.Hash{}, uuid.Nil, err
		}
		hasher.Write(tx.Hash().Bytes())
	}
	hashBytes := hasher.Sum(nil)

	// then compute the uuid
	var buf []byte
	buf = binary.AppendVarint(buf, int64(b.BlockNumber))
	buf = append(buf, hashBytes...)
	sort.Slice(b.RevertingTxHashes, func(i, j int) bool {
		return bytes.Compare(b.RevertingTxHashes[i][:], b.RevertingTxHashes[j][:]) <= 0
	})
	for _, txHash := range b.RevertingTxHashes {
		buf = append(buf, txHash[:]...)
	}
	return common.BytesToHash(hashBytes),
		uuid.NewHash(sha256.New(), uuid.Nil, buf, 5),
		nil
}

func (b *MevSendBundleArgs) UniqueKey() uuid.UUID {
	hash := newHash()
	uniqueKeyMevSendBundle(b, hash)
	return uuidFromHash(hash)
}

func uniqueKeyMevSendBundle(b *MevSendBundleArgs, hash hash.Hash) {
	hash.Write([]byte(b.ReplacementUUID))
	_ = binary.Write(hash, binary.LittleEndian, b.Inclusion.BlockNumber)
	_ = binary.Write(hash, binary.LittleEndian, b.Inclusion.MaxBlock)
	for _, body := range b.Body {
		if body.Bundle != nil {
			uniqueKeyMevSendBundle(body.Bundle, hash)
		} else if body.Tx != nil {
			hash.Write(*body.Tx)
		}
		// body.Hash should not occur at this point
		if body.CanRevert {
			hash.Write([]byte{1})
		} else {
			hash.Write([]byte{0})
		}
		hash.Write([]byte(body.RevertMode))
	}
	_, _ = hash.Write(b.Metadata.Signer.Bytes())
}

func (b *MevSendBundleArgs) Validate() (common.Hash, error) {
	// only cancell call can be without txs
	// cancell call must have ReplacementUUID set
	if len(b.Body) == 0 && b.ReplacementUUID == "" {
		return common.Hash{}, ErrBundleNoTxs
	}
	return hashMevSendBundle(0, b)
}

func hashMevSendBundle(level int, b *MevSendBundleArgs) (common.Hash, error) {
	if level > MevBundleMaxDepth {
		return common.Hash{}, ErrMevBundleTooDeep
	}
	hasher := sha3.NewLegacyKeccak256()
	for _, body := range b.Body {
		if body.Hash != nil {
			return common.Hash{}, ErrMevBundleUnmatchedTx
		} else if body.Bundle != nil {
			innerHash, err := hashMevSendBundle(level+1, body.Bundle)
			if err != nil {
				return common.Hash{}, err
			}
			hasher.Write(innerHash.Bytes())
		} else if body.Tx != nil {
			tx := new(types.Transaction)
			if err := tx.UnmarshalBinary(*body.Tx); err != nil {
				return common.Hash{}, err
			}
			hasher.Write(tx.Hash().Bytes())
		}
	}
	return common.BytesToHash(hasher.Sum(nil)), nil
}

func (tx *EthSendRawTransactionArgs) UniqueKey() uuid.UUID {
	hash := newHash()
	_, _ = hash.Write(*tx)
	return uuidFromHash(hash)
}

func (b *EthCancelBundleArgs) UniqueKey() uuid.UUID {
	hash := newHash()
	_, _ = hash.Write([]byte(b.ReplacementUUID))
	_, _ = hash.Write(b.SigningAddress.Bytes())
	return uuidFromHash(hash)
}

func (b *BidSubsisideBlockArgs) UniqueKey() uuid.UUID {
	hash := newHash()
	_ = binary.Write(hash, binary.LittleEndian, uint64(*b))
	return uuidFromHash(hash)
}
