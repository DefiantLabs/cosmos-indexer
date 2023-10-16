package core

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32" // nolint:staticcheck
)

// consider not using globals
var (
	addressRegex  *regexp.Regexp
	addressPrefix string
)

// TODO query this list from the DB
var baseChainPrefixes = []string{
	"juno",
	"cosmos",
	"osmo",
}

func GetAddressPrefix(address string) string {
	for _, chain := range baseChainPrefixes {
		if strings.HasPrefix(address, chain) {
			regex := fmt.Sprintf("(?P<prefix>%s(valoper)?)1[a-z0-9]{38}", chain)
			r := regexp.MustCompile(regex)
			matches := r.FindStringSubmatch(address)

			// the match array will be in the order: full match, then prefix
			if len(matches) >= 2 {
				return matches[1]
			}
		}
	}

	return ""
}

func IsAddressEqual(addr1 string, prefix1 string, addr2 string, prefix2 string) bool {
	bAddr1, err1 := cosmostypes.GetFromBech32(addr1, prefix1)
	bAddr2, err2 := cosmostypes.GetFromBech32(addr2, prefix2)

	if err1 != nil || err2 != nil {
		return false
	}

	return bytes.Equal(bAddr1, bAddr2)
}

func SetupAddressRegex(addressRegexPattern string) {
	addressRegex, _ = regexp.Compile(addressRegexPattern)
}

func SetupAddressPrefix(addressPrefixString string) {
	addressPrefix = addressPrefixString
}

func ExtractTransactionAddresses(tx tx.MergedTx) []string {
	messagesAddresses := util.WalkFindStrings(tx.Tx.Body.Messages, addressRegex)
	// Consider walking logs - needs benchmarking compared to whole string search on raw log
	logAddresses := addressRegex.FindAllString(tx.TxResponse.RawLog, -1)
	addressMap := make(map[string]string)
	for _, v := range append(messagesAddresses, logAddresses...) {
		addressMap[v] = ""
	}
	uniqueAddresses := make([]string, len(addressMap))
	i := 0
	for k := range addressMap {
		uniqueAddresses[i] = k
		i++
	}
	return uniqueAddresses
}

func ParseSignerAddress(pubkeyString string, keytype string) (retstring string, reterror error) {
	defer func() {
		if r := recover(); r != nil {
			reterror = fmt.Errorf(fmt.Sprintf("Error parsing signer address into Bech32: %v", r))
			retstring = ""
		}
	}()

	pubkey, err := getPubKeyFromRawString(pubkeyString, keytype)
	if err != nil {
		fmt.Println("Error getting public key from raw string")
		fmt.Println(err)
		return "", err
	}

	// this panics if conversion fails
	bech32address := cosmostypes.MustBech32ifyAddressBytes(addressPrefix, pubkey.Address().Bytes())
	return bech32address, nil
}

// the following code is taken from here https://github.com/cosmos/cosmos-sdk/blob/9ff6d5441db2260e7877724df65c0f2b8251d991/client/debug/main.go
// they do a check in bytesToPubkey for the keytype of "ed25519", we may want to pass in the keytype but this seems to work
// with secp256k1 keys without passing in the keytype.
// The key type seems to be in the @type key of in the public_key block in signer_infos so we could potentially pass it in there
func getPubKeyFromRawString(pkstr string, keytype string) (cryptotypes.PubKey, error) {
	bz, err := hex.DecodeString(pkstr)
	if err == nil {
		pk, ok := bytesToPubkey(bz, keytype)
		if ok {
			return pk, nil
		}
	}

	bz, err = base64.StdEncoding.DecodeString(pkstr)
	if err == nil {
		pk, ok := bytesToPubkey(bz, keytype)
		if ok {
			return pk, nil
		}
	}

	pk, err := legacybech32.UnmarshalPubKey(legacybech32.AccPK, pkstr) // nolint:staticcheck
	if err == nil {
		return pk, nil
	}

	pk, err = legacybech32.UnmarshalPubKey(legacybech32.ValPK, pkstr) // nolint:staticcheck
	if err == nil {
		return pk, nil
	}

	pk, err = legacybech32.UnmarshalPubKey(legacybech32.ConsPK, pkstr) // nolint:staticcheck
	if err == nil {
		return pk, nil
	}

	return nil, fmt.Errorf("pubkey '%s' invalid; expected hex, base64, or bech32 of correct size", pkstr)
}

func bytesToPubkey(bz []byte, keytype string) (cryptotypes.PubKey, bool) {
	if keytype == "ed25519" {
		if len(bz) == ed25519.PubKeySize {
			return &ed25519.PubKey{Key: bz}, true
		}
	}

	if len(bz) == secp256k1.PubKeySize {
		return &secp256k1.PubKey{Key: bz}, true
	}
	return nil, false
}
