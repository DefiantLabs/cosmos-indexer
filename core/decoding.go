package core

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	probeClient "github.com/nodersteam/probe/client"
	"github.com/rs/zerolog/log"
)

// InAppTxDecoder Provides an in-app tx decoder.
// The primary use-case for this function is to allow fallback decoding if a TX fails to decode after RPC requests.
// This can happen in a number of scenarios, but mainly due to missing proto definitions.
// We can attempt a personal decode of the TX, and see if we can continue indexing based on in-app conditions (such as message type filters).
// This function skips a large chunk of decoding validations, and is not recommended for general use. Its main point is to skip errors that in
// default Cosmos TX decoders would cause the entire decode to fail.
func InAppTxDecoder(cdc probeClient.Codec) sdk.TxDecoder {
	return func(txBytes []byte) (sdk.Tx, error) {
		var raw tx.TxRaw
		var err error

		err = cdc.Marshaler.Unmarshal(txBytes, &raw)
		if err != nil {
			return nil, err
		}

		var body tx.TxBody
		err = body.Unmarshal(raw.BodyBytes)
		if err != nil {
			log.Err(err).Msgf("failed to unmarshal tx body")
			return nil, errors.New("failed to unmarshal tx body")
		}

		for _, any := range body.Messages {
			var msg sdk.Msg
			// We deliberately ignore errors here to build up a
			// list of properly decoded messages for later analysis.
			cdc.Marshaler.UnpackAny(any, &msg) //nolint:errcheck
		}

		var authInfo tx.AuthInfo

		err = cdc.Marshaler.Unmarshal(raw.AuthInfoBytes, &authInfo)
		if err != nil {
			return nil, errors.New("failed to unmarshal auth info")
		}

		theTx := &tx.Tx{
			Body:       &body,
			AuthInfo:   &authInfo,
			Signatures: raw.Signatures,
		}

		return theTx, nil
	}
}
