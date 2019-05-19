package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
)

// Cancels a delayed transaction.
// cancel [cancelling_authority] [transaction_id]
func TxCancel(authority eos.PermissionLevel, transactionID eos.SHA256Bytes) {
	api := getAPI()
	pushEOSCActions(api, system.NewCancelDelay(authority, transactionID))
}

// Create a transaction with a single action
// create [contract] [action] [payload]
func TxCreate(contract eos.AccountName, action eos.ActionName, payload string) {
	forceUnique := viper.GetBool("tx-create-cmd-force-unique")

	var dump map[string]interface{}
	err := json.Unmarshal([]byte(payload), &dump)
	errorCheck("[payload] is not valid JSON", err)

	api := getAPI()
	actionBinary, err := api.ABIJSONToBin(contract, eos.Name(action), dump)
	errorCheck("unable to retrieve action binary from JSON via API", err)

	actions := []*eos.Action{
		&eos.Action{
			Account:    contract,
			Name:       action,
			ActionData: eos.NewActionDataFromHexData([]byte(actionBinary)),
		}}

	var contextFreeActions []*eos.Action
	if forceUnique {
		contextFreeActions = append(contextFreeActions, newNonceAction())
	}

	pushEOSCActionsAndContextFreeActions(api, contextFreeActions, actions)
}

func newNonceAction() *eos.Action {
	return &eos.Action{
		Account: eos.AN("eosio.null"),
		Name:    eos.ActN("nonce"),
		ActionData: eos.NewActionData(system.Nonce{
			Value: hex.EncodeToString(generateRandomNonce()),
		}),
	}
}

func generateRandomNonce() []byte {
	// Use 48 bits of entropy to generate a valid random
	nonce := make([]byte, 6)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		errorCheck("unable to correctly generate nonce", err)
	}

	return nonce
}

// Print the transaction ID for a given transaction file.
// id [transaction.json]
func TxId(filename string) {
	cnt, err := ioutil.ReadFile(filename)
	errorCheck("reading file", err)

	var stx *eos.SignedTransaction
	errorCheck("parsing JSON content", json.Unmarshal(cnt, &stx))

	ptx, err := stx.Pack(eos.CompressionNone)
	errorCheck("packing transaction", err)

	id, _ := ptx.ID()
	fmt.Println(hex.EncodeToString(id))
}

// Push a signed transaction to the chain.  Must be done online.
// push [transaction.json]
func TxPush(filename string) {
	cnt, err := ioutil.ReadFile(filename)
	errorCheck("reading transaction file", err)

	chainID := gjson.GetBytes(cnt, "chain_id").String()
	hexChainID, _ := hex.DecodeString(chainID)

	var signedTx *eos.SignedTransaction
	errorCheck("json unmarshal transaction", json.Unmarshal(cnt, &signedTx))

	api := getAPI()

	packedTx, err := signedTx.Pack(eos.CompressionNone)
	errorCheck("packing transaction", err)

	pushTransaction(api, packedTx, eos.SHA256Bytes(hexChainID))
}

// Sign a transaction produced by --write-transaction and submit it to the chain (unless --write-transaction is passed again).
// sign [transaction.yaml|json]
func TxSign(filename string) {
	cnt, err := ioutil.ReadFile(filename)
	errorCheck("reading transaction file", err)

	var tx *eos.Transaction
	errorCheck("json unmarshal transaction", json.Unmarshal(cnt, &tx))

	api := getAPI()

	var chainID eos.SHA256Bytes
	if infileChainID := gjson.Get(string(cnt), "chain_id").String(); infileChainID != "" {
		chainID = toSHA256Bytes(infileChainID, fmt.Sprintf("chain_id field in %q", filename))
	} else if cliChainID := viper.GetString("global-offline-chain-id"); cliChainID != "" {
		chainID = toSHA256Bytes(cliChainID, "--offline-chain-id")
	} else {
		// getInfo
		resp, err := api.GetInfo()
		errorCheck("get info", err)
		chainID = resp.ChainID
	}

	signedTx, packedTx := optionallySignTransaction(tx, chainID, api)

	optionallyPushTransaction(signedTx, packedTx, chainID, api)
}

// Unpack a transaction produced by --write-transaction and display all its actions (for review).  This does not submit anything to the chain.
// unpack [transaction.yaml|json]
func TxUnpack(filename string) {
	cnt, err := ioutil.ReadFile(filename)
	errorCheck("reading transaction file", err)

	var tx *eos.SignedTransaction
	errorCheck("json unmarshal transaction", json.Unmarshal(cnt, &tx))

	api := getAPI()

	for _, act := range tx.ContextFreeActions {
		errorCheck("context free action unpack", txUnpackAction(api, act))
	}
	for _, act := range tx.Actions {
		errorCheck("action unpack", txUnpackAction(api, act))
	}

	cnt, err = json.MarshalIndent(tx, "", "  ")
	errorCheck("marshalling signed transaction", err)

	fmt.Println(string(cnt))
}

func txUnpackAction(api *eos.API, act *eos.Action) error {
	hexBytes, ok := act.Data.(string)
	if !ok {
		return fmt.Errorf("action data expected to be hex bytes as string, was %T", act.Data)
	}
	bytes, err := hex.DecodeString(hexBytes)
	if err != nil {
		return fmt.Errorf("invalid hex bytes stream: %s", err)
	}

	data, err := api.ABIBinToJSON(act.Account, eos.Name(act.Name), bytes)
	if err != nil {
		return fmt.Errorf("chain abi_bin_to_json: %s", err)
	}

	act.Data = data
	return nil
}