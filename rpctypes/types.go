// Package rpctypes implement types commonly used in the Flashbots codebase for receiving and senging requests
package rpctypes

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash"
	"math/big"
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
	BundleVersionV1   = "v1"
	BundleVersionV2   = "v2"
)

var (
	ErrBundleNoTxs              = errors.New("bundle with no txs")
	ErrBundleTooManyTxs         = errors.New("too many txs in bundle")
	ErrMevBundleUnmatchedTx     = errors.New("mev bundle with unmatched tx")
	ErrMevBundleTooDeep         = errors.New("mev bundle too deep")
	ErrUnsupportedBundleVersion = errors.New("unsupported bundle version")
)

type EthSendBundleArgs struct {
	Txs               []hexutil.Bytes `json:"txs"`
	BlockNumber       *hexutil.Uint64 `json:"blockNumber"`
	MinTimestamp      *uint64         `json:"minTimestamp,omitempty"`
	MaxTimestamp      *uint64         `json:"maxTimestamp,omitempty"`
	RevertingTxHashes []common.Hash   `json:"revertingTxHashes,omitempty"`
	ReplacementUUID   *string         `json:"replacementUuid,omitempty"`
	Version           *string         `json:"version,omitempty"`

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
	blockNumber := uint64(0)
	if b.BlockNumber != nil {
		blockNumber = uint64(*b.BlockNumber)
	}
	hash := newHash()
	_ = binary.Write(hash, binary.LittleEndian, blockNumber)
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

	sort.Slice(b.DroppingTxHashes, func(i, j int) bool {
		return bytes.Compare(b.DroppingTxHashes[i][:], b.DroppingTxHashes[j][:]) <= 0
	})
	for _, txHash := range b.DroppingTxHashes {
		_, _ = hash.Write(txHash.Bytes())
	}
	if b.RefundPercent != nil {
		_ = binary.Write(hash, binary.LittleEndian, *b.RefundPercent)
	}

	if b.RefundRecipient != nil {
		_, _ = hash.Write(b.RefundRecipient.Bytes())
	}
	for _, txHash := range b.RefundTxHashes {
		_, _ = hash.Write([]byte(txHash))
	}

	if b.SigningAddress != nil {
		_, _ = hash.Write(b.SigningAddress.Bytes())
	}
	return uuidFromHash(hash)
}

func (b *EthSendBundleArgs) Validate() (common.Hash, uuid.UUID, error) {
	blockNumber := uint64(0)
	if b.BlockNumber != nil {
		blockNumber = uint64(*b.BlockNumber)
	}
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

	if b.Version == nil || *b.Version == BundleVersionV1 {
		// then compute the uuid
		var buf []byte
		buf = binary.AppendVarint(buf, int64(blockNumber))
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

	if *b.Version == BundleVersionV2 {
		// blockNumber, default 0
		blockNumber := uint64(0)
		if b.BlockNumber != nil {
			blockNumber = uint64(*b.BlockNumber)
		}

		// minTimestamp, default 0
		minTimestamp := uint64(0)
		if b.MinTimestamp != nil {
			minTimestamp = *b.MinTimestamp
		}

		// maxTimestamp, default ^uint64(0) (i.e. 0xFFFFFFFFFFFFFFFF in Rust)
		maxTimestamp := ^uint64(0)
		if b.MaxTimestamp != nil {
			maxTimestamp = *b.MaxTimestamp
		}

		// Build up our buffer using variable-length encoding of the block
		// number, minTimestamp, maxTimestamp, #revertingTxHashes, #droppingTxHashes.
		var buf []byte
		buf = binary.AppendUvarint(buf, blockNumber)
		buf = binary.AppendUvarint(buf, minTimestamp)
		buf = binary.AppendUvarint(buf, maxTimestamp)
		buf = binary.AppendUvarint(buf, uint64(len(b.RevertingTxHashes)))
		buf = binary.AppendUvarint(buf, uint64(len(b.DroppingTxHashes)))

		// Append the main txs keccak hash (already computed in hashBytes).
		buf = append(buf, hashBytes...)

		// Sort revertingTxHashes and append them.
		sort.Slice(b.RevertingTxHashes, func(i, j int) bool {
			return bytes.Compare(b.RevertingTxHashes[i][:], b.RevertingTxHashes[j][:]) < 0
		})
		for _, h := range b.RevertingTxHashes {
			buf = append(buf, h[:]...)
		}

		// Sort droppingTxHashes and append them.
		sort.Slice(b.DroppingTxHashes, func(i, j int) bool {
			return bytes.Compare(b.DroppingTxHashes[i][:], b.DroppingTxHashes[j][:]) < 0
		})
		for _, h := range b.DroppingTxHashes {
			buf = append(buf, h[:]...)
		}

		// If a "refund" is present (analogous to the Rust code), we push:
		//   refundPercent (1 byte)
		//   refundRecipient (20 bytes, if an Ethereum address)
		//   #refundTxHashes (varint)
		//   each refundTxHash (32 bytes)
		// NOTE: The Rust code uses a single byte for `refund.percent`,
		//       so we do the same here
		if b.RefundPercent != nil && *b.RefundPercent != 0 {
			if len(b.Txs) == 0 {
				// Bundle with not txs can't be refund-recipient
				return common.Hash{}, uuid.Nil, ErrBundleNoTxs
			}

			// We only keep the low 8 bits of RefundPercent (mimicking Rust's `buff.push(u8)`).
			buf = append(buf, byte(*b.RefundPercent))

			refundRecipient := b.RefundRecipient
			if refundRecipient == nil {
				var tx types.Transaction
				if err := tx.UnmarshalBinary(b.Txs[0]); err != nil {
					return common.Hash{}, uuid.Nil, err
				}
				from, err := types.Sender(types.LatestSignerForChainID(big.NewInt(1)), &tx)
				if err != nil {
					return common.Hash{}, uuid.Nil, err
				}
				refundRecipient = &from
			}
			bts := [20]byte(*refundRecipient)
			// RefundRecipient is a common.Address, which is 20 bytes in geth.
			buf = append(buf, bts[:]...)

			var refundTxHashes []common.Hash
			for _, rth := range b.RefundTxHashes {
				// decode from hex
				refundTxHashes = append(refundTxHashes, common.HexToHash(rth))
			}

			if len(refundTxHashes) == 0 {
				var lastTx types.Transaction
				if err := lastTx.UnmarshalBinary(b.Txs[len(b.Txs)-1]); err != nil {
					return common.Hash{}, uuid.Nil, err
				}
				refundTxHashes = []common.Hash{lastTx.Hash()}
			}

			// #refundTxHashes
			buf = binary.AppendUvarint(buf, uint64(len(refundTxHashes)))

			sort.Slice(refundTxHashes, func(i, j int) bool {
				return bytes.Compare(refundTxHashes[i][:], refundTxHashes[j][:]) < 0
			})
			for _, h := range refundTxHashes {
				buf = append(buf, h[:]...)
			}
		}

		// Now produce a UUID from `buf` using SHA-256 in the same way the Rust code calls
		// `Self::uuid_from_buffer(buff)` (which is effectively a UUIDv5 but with SHA-256).
		finalUUID := uuid.NewHash(sha256.New(), uuid.Nil, buf, 5)

		// Return the main txs keccak hash as well as the computed UUID
		return common.BytesToHash(hashBytes), finalUUID, nil
	}

	return common.Hash{}, uuid.Nil, ErrUnsupportedBundleVersion

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
