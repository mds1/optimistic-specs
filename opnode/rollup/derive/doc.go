// Package derive provides the data transformation functions that take L1 data
// and turn it into L2 blocks and results. Certian L2 data is also able to
// turned back into L1 data.
//
// The flow is data is as follows
// receipts -> l2.PayloadAttributes with `payload_attributes.go`
// l2.PayloadAttributes -> l2.ExecutionPayload with `execution_payload.go`
// L2 block -> Corresponding L1 block with `invert.go`
//
// The Payload Atrribute derivation stage is a pure function.
// The Execution Payload derivation stage relies on the L2 execution engine to perform the
// state update.
// The inversion step is a pure function.
//
// The steps should be keep separate to enable easier testing.
package derive
