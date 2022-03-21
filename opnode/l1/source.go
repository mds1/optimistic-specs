package l1

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum-optimism/optimistic-specs/opnode/eth"
	"github.com/ethereum-optimism/optimistic-specs/opnode/rollup/derive"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Source struct {
	client     *ethclient.Client
	downloader *Downloader
}

func NewSource(client *ethclient.Client) Source {
	return Source{
		client:     client,
		downloader: NewDownloader(client),
	}
}

func (s Source) BlockLinkByNumber(ctx context.Context, num uint64) (self eth.BlockID, parent eth.BlockID, err error) {
	header, err := s.client.HeaderByNumber(ctx, big.NewInt(int64(num)))
	if err != nil {
		// w%: wrap the error, we still need to detect if a canonical block is not found, a.k.a. end of chain.
		return eth.BlockID{}, eth.BlockID{}, fmt.Errorf("failed to determine block-hash of height %d, could not get header: %w", num, err)
	}
	parentNum := num
	if parentNum > 0 {
		parentNum -= 1
	}
	return eth.BlockID{Hash: header.Hash(), Number: num}, eth.BlockID{Hash: header.ParentHash, Number: parentNum}, nil

}

func (s Source) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	return s.client.SubscribeNewHead(ctx, ch)
}

func (s Source) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	return s.client.HeaderByHash(ctx, hash)
}

func (s Source) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return s.client.HeaderByNumber(ctx, number)
}

func (s Source) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return s.client.TransactionReceipt(ctx, txHash)
}

func (s Source) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return s.client.BlockByHash(ctx, hash)
}

func (s Source) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	return s.client.BlockByNumber(ctx, number)
}

func (s Source) Fetch(ctx context.Context, id eth.BlockID) (*types.Block, []*types.Receipt, error) {
	return s.downloader.Fetch(ctx, id)
}

func (s Source) Close() {
	s.client.Close()
}

func (s Source) FetchL1Info(ctx context.Context, id eth.BlockID) (derive.L1Info, error) {
	return s.client.BlockByHash(ctx, id.Hash)
}

func (s Source) FetchReceipts(ctx context.Context, id eth.BlockID, receiptHash common.Hash) ([]*types.Receipt, error) {
	_, receipts, err := s.Fetch(ctx, id)
	if err != nil {
		return nil, err
	}
	// Sanity-check: external L1-RPC sources are notorious for not returning all receipts,
	// or returning them out-of-order. Verify the receipts against the expected receipt-hash.
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(receipts), hasher)
	if receiptHash != computed {
		return nil, fmt.Errorf("failed to validate receipts of %s, computed receipt-hash %s does not match expected hash %d", id, computed, receiptHash)
	}
	return receipts, err
}

func (s Source) FetchTransactions(ctx context.Context, window []eth.BlockID) ([]*types.Transaction, error) {
	var txns []*types.Transaction
	for _, id := range window {
		block, err := s.client.BlockByHash(ctx, id.Hash)
		if err != nil {
			return nil, err
		}
		txns = append(txns, block.Transactions()...)
	}
	return txns, nil
}
