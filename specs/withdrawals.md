# Withdrawals

<!-- All glossary references in this file. -->
[g-deposited]: glossary.md#deposited-transaction
[g-withdrawal]: glossary.md#withdrawal
[g-mpt]: glossary.md#merkle-patricia-trie
[g-relayer]: glossary.md#withdrawals
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [Constants](#constants)
- [Data structures](#data-structures)
- [Withdrawal verification](#withdrawal-verification)
- [Withdrawal Flow](#withdrawal-flow)
  - [Rationale](#rationale)
- [The L2 Withdrawal Contract](#the-l2-withdrawal-contract)
- [Trie Interface](#trie-interface)
- [Key differences from Current System](#key-differences-from-current-system)
- [Security Considerations](#security-considerations)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

Withdrawals are transactions initiated on L2, and executed on L1. They may be used to transfer data, ether or both from
L1 to L2.

**Vocabulary note**:

- _withdrawing transaction_ refers specifically to an L2 transaction
- _withdrawn transaction_ refers specifically to an L1 transaction
- _withdrawal_ can refer to the transaction at various stages

## Constants

| Name                  | Value     | Unit    |
| --------------------- | --------- | ------- |
| `FINALIZATION_WINDOW` | `604_800` | seconds |

This `FINALIZATION_WINDOW` value is equivalent to 7 days.

## Data Structures and Type Aliases

1. A `WithdrawalMessage` is encoded in a struct as follows:

    ```js
    struct WithdrawalMessage {
        uint256 nonce;
        address sender;
        uint256 value;
        bytes message;
    }
    ```

1. The `L2OutputTimestamp` is an alias for `uint256`, and MUST be a multiple of the `SUBMISSION_INTERVAL` described
  in the [L2 Output](./proposals.md#constants) document.

1. The `WithdrawalsRoot` is an alias for a `bytes32` value, corresponding to the [MPT][g-mpt]
  storage root of the Withdrawals predeploy contract at `0x4200000000000000000000000000000000000015` (described below).

1. The `WithdrawalsRootInclusionProof` proof contains the data necessary to prove that the provided `WithdrawalsRoot` is
  included in the [SSZ merkleization](https://github.com/ethereum/consensus-specs/blob/dev/ssz/simple-serialize.md#merkleization)
  of the `L2Output` as defined in the
  [L2 output commitment construction](./proposals.md#l2-output-commitment-construction).

1. A `WithdrawalMessageInclusionProof` is an MPT proof encoded in a struct as follows:

   ```js
   struct WithdrawalMessageInclusionProof {
       WithdrawalMessage message;
       bytes32 l2withdrawalsRoot;
       bytes memory _key;      // storage key of the withdrawal message commitment
       bytes memory _value;    // Always bytes32(1) (boolean validity status of the message)
       bytes memory _proof;    // MPT inclusion proof for the key/value
   }
   ```
1. This document also refers to the `L2Output` type as defined in the
  [L2 output commitment construction](./proposals.md#l2-output-commitment-construction).

## Withdrawal verification

In order to verify a `WithdrawalMessage`, the following inputs must be provided:

| Type                              | Name               |
| --------------------------------- | ------------------ |
| `WithdrawalMessage`               | `message`          |
| `L2OutputTimestamp`               | `timestamp`        |
| `WithdrawalsRootInclusionProof`   | `storageRootProof` |
| `WithdrawalMessageInclusionProof` | `messageProof`     |

These inputs must satisfy the following conditions:

1. The `timestamp` is at least `FINALIZATION_WINDOW` seconds old.
1. `OutputOracle.l2Outputs(timestamp)` returns a non-zero value `l2Output`.
1. The `storageRootProof` is valid for the `l2Output` returned in step 2, according to the SSZ encoding described in the
   [L2 output commitment construction](./proposals.md#l2-output-commitment-construction).
1. The `messageProof` is valid for the provided `message` and `messageProof`

## Withdrawal Flow

**On L2:**

1. An L2 account sends a withdrawal message (and possibly also ETH) to the `Withdrawoooor` predeploy contract.
   This is a very simple contract that stores a mapping from the hash of the `WithdrawalMessage` as defined above to a
   boolean value. (`mapping (bytes32 => bool) withdrawalMessages`)
2. If ETH is being withdrawn, it can eventually be burned by deploying a contract which immediately `SELFDESTRUCT`s.
   This has the benefit of using an existing EVM mechanism for removing the ETH from the world state, without having to
   add a diff to L2 Geth.

**On L1:**

1. A [relayer][g-relayer] submits the required inputs to the `DepositFeed` contract. The relayer may or may not be the same entity which
   initiated the withdrawal on L2.
2. The `DepositFeed` contract retrieves the output root from the `OutputOracle`'s `l2Outputs()` function, and performs
   the remainder of the verification process internally.
3. If verification is successful, the message is forwarded to the target.
    1. If the message call is successful, the hash is stored in a `successfulMessages` mapping.
    2. Otherwise it is stored in a `repeatableMessages` mapping.
4. If verification fails, the call reverts.

## The L2 Withdrawals Contract

The L2 Withdrawals predeploy is a simple contract at `0x4200000000000000000000000000000000000015` which stores messages
to be withdrawn.

It contains a mapping which records withdrawals.

```js
interface Withdrawoooor {

    event WithdrawalMessage(
        address indexed target,
        address sender,
        bytes message,
        uint256 value,
        uint256 messageNonce, // this is a global nonce value for all withdrawal messages
        uint256 gasLimit
    );

    function withdraw(
        address target,
        address sender,
        bytes message,
        uint256 gasLimit
    ) external payable;

    function burn();
}
```

## Security Considerations

### Key Properties of Withdrawal Verification

1. It should not be possible 'double spend' a withdrawal, ie. to relay a `WithdrawalMessage` on L1 which does not correspond to a message
  initiated on L2. For reference, see [this writeup][polygon-dbl-spend] of a vulnerability of this type found on Polygon.

    [polygon-dbl-spend]: https://gerhard-wagner.medium.com/double-spending-bug-in-polygons-plasma-bridge-2e0954ccadf1


1. For each `WithdrawalMessage` initiated on L2 (ie. with a unique `messageNonce`), the following properties must hold:
    1. There should be one (and only one) relayable message.
    1. It should not be possible to relay the message with any of its fields modified, ie.
        1. Modifying the `sender` field would enable a 'spoofing' attack.
        1. Modifying the `target`, `message`, or `value` fields would enable an attacker to dangerously change the intended outcome of the withdrawal.
        1. Modifying the `gasLimit` could make the cost of relaying too high, or allow the relayer to cause execution to fail (out of gas) in the `target`.


### Handling Successfully Verified Messages That Fail When Relayed

If the execution of the relayed call fails in the `target` contracts, it is unfortunately not possible to determine
whether or not it was 'supposed' to fail, and whether or not it should be 'replayable'.
Thus we provide the following mitigations:
1. The minimum gas amount to be


[Insufficient Gas Griefing]:(https://swcregistry.io/docs/SWC-126)
