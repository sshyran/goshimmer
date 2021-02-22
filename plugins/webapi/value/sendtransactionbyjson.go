package value

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/goshimmer/packages/tangle/payload"
	"github.com/iotaledger/goshimmer/plugins/issuer"
	"github.com/iotaledger/goshimmer/plugins/messagelayer"
	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58/base58"
)

var (
	sendTxByJSONMu sync.Mutex

	// ErrMalformedIdentityID defines a malformed identityID error.
	ErrMalformedIdentityID = fmt.Errorf("malformed identityID")
	// ErrMalformedInputs defines a malformed inputs error.
	ErrMalformedInputs = fmt.Errorf("malformed inputs")
	// ErrMalformedOutputs defines a malformed outputs error.
	ErrMalformedOutputs = fmt.Errorf("malformed outputs")
	// ErrMalformedData defines a malformed data error.
	ErrMalformedData = fmt.Errorf("malformed data")
	// ErrMalformedColor defines a malformed color error.
	ErrMalformedColor = fmt.Errorf("malformed color")
	// ErrMalformedPublicKey defines a malformed publicKey error.
	ErrMalformedPublicKey = fmt.Errorf("malformed publicKey")
	// ErrMalformedSignature defines a malformed signature error.
	ErrMalformedSignature = fmt.Errorf("malformed signature")
	// ErrWrongSignature defines a wrong signature error.
	ErrWrongSignature = fmt.Errorf("wrong signature")
	// ErrSignatureVersion defines a unsupported signature version error.
	ErrSignatureVersion = fmt.Errorf("unsupported signature version")
)

// sendTransactionByJSONHandler sends a transaction.
func sendTransactionByJSONHandler(c echo.Context) error {
	sendTxByJSONMu.Lock()
	defer sendTxByJSONMu.Unlock()

	var request SendTransactionByJSONRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, SendTransactionByJSONResponse{Error: err.Error()})
	}

	tx, err := NewTransactionFromJSON(request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, SendTransactionByJSONResponse{Error: err.Error()})
	}

	// validate transaction
	// err = valuetransfers.Tangle().ValidateTransactionToAttach(tx)
	// if err != nil {
	// 	return c.JSON(http.StatusBadRequest, SendTransactionByJSONResponse{Error: err.Error()})
	// }

	issueTransaction := func() (*tangle.Message, error) {
		msg, e := issuer.IssuePayload(tx, messagelayer.Tangle())
		if e != nil {
			return nil, c.JSON(http.StatusBadRequest, SendTransactionResponse{Error: e.Error()})
		}
		return msg, nil
	}

	_, err = messagelayer.AwaitMessageToBeBooked(issueTransaction, tx.ID(), maxBookedAwaitTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, SendTransactionByJSONResponse{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, SendTransactionByJSONResponse{TransactionID: tx.ID().String()})
}

// NewTransactionFromJSON returns a new transaction from a given JSON request or an error.
func NewTransactionFromJSON(request SendTransactionByJSONRequest) (*ledgerstate.Transaction, error) {
	// prepare inputs
	inputs := make(ledgerstate.Inputs, len(request.Inputs))
	for i, input := range request.Inputs {
		b, err := base58.Decode(input)
		in, _, err := ledgerstate.InputFromBytes(b)
		if err != nil || len(b) != ledgerstate.OutputIDLength {
			return nil, ErrMalformedInputs
		}
		inputs[i] = in
	}

	// prepare outputs
	outputsMap := make(map[ledgerstate.Address]*ledgerstate.ColoredBalances)
	outputs := make([]ledgerstate.Output, len(request.Outputs))
	for i, output := range request.Outputs {
		balances := make(map[ledgerstate.Color]uint64)
		address, err := ledgerstate.AddressFromBase58EncodedString(output.Address)
		if err != nil {
			return nil, ErrMalformedOutputs
		}

		for _, b := range output.Balances {
			var color ledgerstate.Color
			if b.Color == "IOTA" {
				color = ledgerstate.ColorIOTA
			} else {
				colorBytes, err := base58.Decode(b.Color)
				color, _, err = ledgerstate.ColorFromBytes(colorBytes)
				if err != nil || len(colorBytes) != ledgerstate.ColorLength {
					return nil, ErrMalformedColor
				}
			}
			balances[color] += uint64(b.Value)
		}

		outputsMap[address] = ledgerstate.NewColoredBalances(balances)
		outputs[i] = ledgerstate.NewSigLockedColoredOutput(outputsMap[address], address)
	}

	aManaPledgeID, err := identity.ParseID(request.AManaPledge)
	if err != nil {
		return nil, ErrMalformedIdentityID
	}
	cManaPledgeID, err := identity.ParseID(request.CManaPledge)
	if err != nil {
		return nil, ErrMalformedIdentityID
	}

	txEssence := ledgerstate.NewTransactionEssence(
		0,
		request.Timestamp,
		aManaPledgeID,
		cManaPledgeID,
		ledgerstate.NewInputs(inputs...),
		ledgerstate.NewOutputs(outputs...),
	)

	// add data payload
	payload, _, err := payload.FromBytes(request.Payload)
	if err != nil {
		return nil, ErrMalformedData
	}

	txEssence.SetPayload(payload)

	// add signatures
	unlockBlocks := make(ledgerstate.UnlockBlocks, len(request.UnlockBlocks))
	for i, signature := range request.UnlockBlocks {
		switch ledgerstate.SignatureType(signature.Type) {

		case ledgerstate.ED25519SignatureType:
			pubKeyBytes, err := base58.Decode(signature.PublicKey)
			if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
				return nil, ErrMalformedPublicKey
			}

			signatureBytes, err := base58.Decode(signature.Signature)
			if err != nil || len(signatureBytes) != ed25519.SignatureSize {
				return nil, ErrMalformedSignature
			}

			marshalUtil := marshalutil.New(1 + ed25519.PublicKeySize + ed25519.SignatureSize)
			marshalUtil.WriteUint8(uint8(ledgerstate.ED25519SignatureType))
			marshalUtil.WriteBytes(pubKeyBytes[:])
			marshalUtil.WriteBytes(signatureBytes[:])

			s, _, err := ledgerstate.ED25519SignatureFromBytes(marshalUtil.Bytes())
			if err != nil {
				return nil, ErrWrongSignature
			}
			unlockBlocks[i] = ledgerstate.NewSignatureUnlockBlock(s)

		case ledgerstate.BLSSignatureType:
			pubKeyBytes, err := base58.Decode(signature.PublicKey)
			if err != nil || len(pubKeyBytes) != bls.PublicKeySize {
				return nil, ErrMalformedPublicKey
			}

			signatureBytes, err := base58.Decode(signature.Signature)
			if err != nil || len(signatureBytes) != bls.SignatureSize {
				return nil, ErrMalformedSignature
			}

			marshalUtil := marshalutil.New(bls.SignatureSize + bls.PublicKeySize)
			marshalUtil.WriteUint8(uint8(ledgerstate.BLSSignatureType))
			marshalUtil.WriteBytes(pubKeyBytes[:])
			marshalUtil.WriteBytes(signatureBytes[:])

			s, _, err := ledgerstate.BLSSignatureFromBytes(marshalUtil.Bytes())
			if err != nil {
				return nil, ErrWrongSignature
			}
			unlockBlocks[i] = ledgerstate.NewSignatureUnlockBlock(s)
		default:
			return nil, ErrSignatureVersion
		}
	}

	// prepare transaction
	tx := ledgerstate.NewTransaction(txEssence, unlockBlocks)

	return tx, nil
}

// SendTransactionByJSONRequest holds the transaction object(json) to send.
// e.g.,
// {
// 	"inputs": string[],
// 	"outputs": {
// 	   "address": string,
// 	   "balances": {
// 		   "value": number,
// 		   "color": string
// 	   }[];
// 	 }[],
// 	 "data": []byte,
// 	 "signatures": {
// 		"version": number,
// 		"publicKey": string,
// 		"signature": string
// 	   }[]
//  }
type SendTransactionByJSONRequest struct {
	Timestamp    time.Time     `json:"timestamp"`
	AManaPledge  string        `json:"a_mana_pledge"`
	CManaPledge  string        `json:"c_mana_pledge"`
	Inputs       []string      `json:"inputs"`
	Outputs      []Output      `json:"outputs"`
	Payload      []byte        `json:"payload,omitempty"`
	UnlockBlocks []UnlockBlock `json:"unlock_blocks"`
}

// SendTransactionByJSONResponse is the HTTP response from sending transaction.
type SendTransactionByJSONResponse struct {
	TransactionID string `json:"transaction_id,omitempty"`
	Error         string `json:"error,omitempty"`
}
