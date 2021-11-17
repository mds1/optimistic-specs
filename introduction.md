# Optimistic Ethereum Introduction

Optimistic Ethereum is an _EVM equivalent_, _optimistic rollup_ protocol designed to _scale Ethereum_ while remaining maximally compatible with existing Ethereum infrastructure. This document provides an overview of the protocol to provide context for the rest of the specification.

## Table of Contents
1. [Foundations](#foundations)
2. [Protocol Guarantees](#protocol-guarantees)
3. [Network Participants](#network-participants)
4. [Key Interaction Diagrams](#key-interaction-diagrams)
    - [Depositing and Transacting](#depositing-and-sending-transactions)
    - [Withdrawing](#withdrawing)
    - [Verifying the Network](#verifying-the-network)
5. [Next Steps](#next-steps)

## Foundations

### What is Ethereum scalability?

Scaling Ethereum means increasing the number of useful transactions the Ethereum network can process. Ethereum's limited resources, specifically bandwidth, computation, and storage, constrain the number of transactions which can be processed on the network. Of the three resources, computation and storage are currently the most significant bottlenecks. These bottlenecks limit the supply of transactions, leading to extremely high fees. Scaling Ethereum and reducing fees can be achieved by increasing the supply of these limited resources.

### What is Optimistic Rollup?

[Optimistic rollup](https://vitalik.ca/general/2021/01/05/rollup.html) is a layer 2 scalability technique which increases the computation & storage capacity of Ethereum without sacrificing security or decentralization. Transaction data is submitted on-chain but executed off-chain. If there is an error in the off-chain execution, a fraud proof can be submitted on-chain to correct the error and protect user funds. In the same way you don't go to court unless there is a dispute, you don't execute transactions on on-chain unless there is an error.

### What is EVM Equivalence?

[EVM Equivalence](https://medium.com/ethereum-optimism/introducing-evm-equivalence-5c2021deb306) is complete compliance with the state transition function described in the Ethereum yellow paper, the formal definition of the protocol. By conforming to the Ethereum standard across EVM equivalent rollups, smart contract developers can write once and deploy anywhere.

### 🎶 All together now 🎶

#### Optimistic Ethereum is an _EVM equivalent_, _optimistic rollup_ protocol designed to _scale Ethereum_.

## Protocol Guarantees

In order to scale Ethereum without sacrificing security, we must preserve 3 critical properties of Ethereum layer 1: liveness, availability, and validity.

1. **Liveness** - Anyone must be able to extend the rollup chain by sending transactions at any time.
    - There are two ways transactions can be sent to the rollup chain: 1) via the sequencer, and 2) directly on layer 1. The sequencer provides low latency & low cost transactions, while sending transactions directly to layer 1 provides censorship resistance.
2. **Availability** - Anyone must be able to download the rollup chain.
    - All information required to derive the chain is embedded into layer 1 blocks. That way as long as the layer 1 chain is available, so is the rollup.
3. **Validity** - All transactions must be correctly executed and all withdrawals correctly processed.
    - The rollup state and withdrawals are managed on an L1 contract called the `L2 State Oracle`. This oracle is guaranteed to _only_ finalize correct (ie. valid) rollup block hashes given a **single honest verifier** assumption. If there is ever an invalid block hash asserted on layer 1, an honest verifier will prove it is invalid and win a bond.

**Footnote**: There are two main ways to enforce validity of a rollup: fraud proofs (optimistic rollup) and validity proofs (zkRollup). For the purposes of this spec we only focus on fraud proofs but it is worth noting that validity proofs can also be plugged in once they have been made feasible.

## Network Participants

There are three actors in Optimistic Ethereum: users, sequencers, and verifiers.

![Network Overview](./assets/network-participants-overview.svg)

### Users

At the heart of the network are users (us!). Users can:

1. Deposit or withdraw tokens by sending transactions to Ethereum mainnet.
2. Use EVM smart contracts on layer 2 by sending transactions to the sequencers.
3. View the status of transactions using block explorers provided by network verifiers.

### Sequencers

The sequencer is the primary block producer. There may be one sequencer **or** many using a consensus protocol. For 1.0.0, there is just one sequencer.  In general, specifications may use "the sequencer" to be a stand-in term for the consensus protocol operated by multiple sequencers.

The sequencer:
1. Accepts user off-chain transactions (exposes `eth_sendRawTransaction`, validates fees, ...)
2. Observes on-chain transactions (primarily, deposit events coming from L1)
3. Consolidates both kinds of transactions into L2 blocks with a specific ordering.
4. Propagates consolidated L2 blocks to L1, by submitting two things as calldata to L1:
    -  The pending off-chain transactions accepted in step 1.
    -  Sufficient information about the ordering of the on-chain transactions to successfully reconstruct the blocks from step 3., purely by watching L1.

The sequencer also provides access to block data as early as step 3., so that users may access real-time state in advance of L1 confirmation if they so choose.

### Verifiers

Verifiers serve two purposes:

1. Serving rollup data to users; and
2. Verifying rollup integrity and disputing invalid assertions.

In order for the network to remain secure there must be **at least** one honest verifier who is able to verify the integrity of the rollup chain & serve blockchain data to users.

## Key Interaction Diagrams

The following diagrams demonstrate how protocol components are utilized during key user interactions in order to provide context when diving into any particular component specification.

### Depositing and Sending Transactions

Users will often begin their L2 journey by depositing ETH from L1. Once they have ETH to pay fees, they'll start sending transactions on L2. The following diagram demonstrates this interaction and all key Optimistic Ethereum components which are utilized:

![Diagram of Depositing and Sending Transactions](./assets/sequencer-handling-deposits-and-transactions.svg)

Links to components mentioned in this diagram:

- Rollup Feeds (WIP)
- Rollup Node (WIP)
- Execution Engine (WIP)
- Sequencer Batch Submitter (WIP)
- L2 State Oracle (WIP)
- Block Hash Assertion Submitter (WIP)
- Fraud Proof VM (WIP)

### Withdrawing

Just as important as depositing, it is critical that users can withdraw from the rollup. Withdrawals are initiated by normal transactions on L2, but then completed using a transaction on L1 after the dispute period has elapsed.

![Diagram of Withdrawing](./assets/user-withdrawing-to-l1.svg)

Links to components mentioned in this diagram:

- Rollup Feeds (WIP)
- L2 State Oracle (WIP)
- Bridge (WIP)


### Verifying The Network

To ensure only valid withdrawals are processed, it is critical that verifiers check each block hash assertion that the sequencer makes and disputes it if there is an error. The following diagram demonstrates a verifier on the network that detects an invalid assertion and submits a dispute, thereby protecting user funds:

![Diagram of a verifier disputing an assertion](./assets/verifier-executing-fraud-proof.svg)

Links to components mentioned in this diagram:

- Rollup Node (WIP)
- Execution Engine (WIP)
- L2 State Oracle (WIP)
- Challenger Agent (WIP)
- Fraud Proof VM (WIP)

## Next Steps

This is a choose your own adventure. Are you interested in how a verifier works under the hood? Maybe you want to dive deep into the bit flippin' Fraud Proof VM? All key components have been linked at least once in this doc, so you should now have the context you need to dive in deeper. [The world is yours](https://www.youtube.com/watch?v=e5PnuIRnJW8)!