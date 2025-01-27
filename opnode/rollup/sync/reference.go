package sync

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// L1Client is the subset of methods that ChainSource needs to determine the L1 block graph
type L1Client interface {
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

// L2Client is the subset of methods that ChainSource needs to determine the L2 block graph
type L2Client interface {
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
}

// ChainSource provides access to the L1 and L2 block graph
type ChainSource interface {
	L1BlockRefByNumber(ctx context.Context, l1Num uint64) (eth.L1BlockRef, error)
	L1HeadBlockRef(ctx context.Context) (eth.L1BlockRef, error)
	L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error)
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
}

func NewChainSource(l1 L1Client, l2 L2Client, genesis *rollup.Genesis) *chainSourceImpl {
	return &chainSourceImpl{l1: l1, l2: l2, genesis: genesis}
}

type chainSourceImpl struct {
	l1      L1Client
	l2      L2Client
	genesis *rollup.Genesis
}

// L1BlockRefByNumber returns the canonical block and parent ids.
func (src chainSourceImpl) L1BlockRefByNumber(ctx context.Context, l1Num uint64) (eth.L1BlockRef, error) {
	return src.l1BlockRefByNumber(ctx, new(big.Int).SetUint64(l1Num))
}

// L1BlockRefByNumber returns the canonical head block and parent ids.
func (src chainSourceImpl) L1HeadBlockRef(ctx context.Context) (eth.L1BlockRef, error) {
	return src.l1BlockRefByNumber(ctx, nil)
}

// l1BlockRefByNumber wraps l1.HeaderByNumber to return an eth.L1BlockRef
// This is internal because the exposed L1BlockRefByNumber takes uint64 instead of big.Ints
func (src chainSourceImpl) l1BlockRefByNumber(ctx context.Context, number *big.Int) (eth.L1BlockRef, error) {
	header, err := src.l1.HeaderByNumber(ctx, number)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L1BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", number, err)
	}
	l1Num := header.Number.Uint64()
	parentNum := l1Num
	if parentNum > 0 {
		parentNum -= 1
	}
	return eth.L1BlockRef{
		Self:   eth.BlockID{Hash: header.Hash(), Number: l1Num},
		Parent: eth.BlockID{Hash: header.ParentHash, Number: parentNum},
	}, nil
}

// L2BlockRefByNumber returns the canonical block and parent ids.
func (src chainSourceImpl) L2BlockRefByNumber(ctx context.Context, l2Num *big.Int) (eth.L2BlockRef, error) {
	block, err := src.l2.BlockByNumber(ctx, l2Num)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Num, err)
	}
	return derive.BlockReferences(block, src.genesis)
}

// L2BlockRefByHash returns the block & parent ids based on the supplied hash. The returned BlockRef may not be in the canonical chain
func (src chainSourceImpl) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	block, err := src.l2.BlockByHash(ctx, l2Hash)
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.L2BlockRef{}, fmt.Errorf("failed to determine block-hash of height %v, could not get header: %w", l2Hash, err)
	}
	return derive.BlockReferences(block, src.genesis)
}
