package wallet

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/msig"
	"github.com/eoscanada/eos-go/p2p"
	"github.com/eoscanada/eos-go/system"
	"github.com/eoscanada/eos-go/token"
	"github.com/ryanuber/columnize"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Runs a p2p protocol-level proxy, and stop sync'ing the chain at the given block-num.
func ChainFreeze(accountName eos.AccountName) {
	chainID, err := hex.DecodeString(viper.GetString("tools-chain-freeze-cmd-chain-id"))
	errorCheck("parsing chain id", err)
	proxy := p2p.NewProxy(
		p2p.NewOutgoingPeer(viper.GetString("tools-chain-freeze-cmd-peer1-p2p-address"), "eos-proxy", nil),
		p2p.NewOutgoingPeer(viper.GetString("tools-chain-freeze-cmd-peer2-p2p-address"), "eos-proxy", &p2p.HandshakeInfo{ChainID: chainID}),
	)

	proxy.RegisterHandler(chainFreezeHandler)
	err = proxy.ConnectAndStart()
	errorCheck("client start", err)
}

var chainFreezeHandler = p2p.HandlerFunc(func(envelope *p2p.Envelope) {
	blockModulo := viper.GetInt("tools-chain-freeze-cmd-on-block-modulo")
	actions := viper.GetString("tools-chain-freeze-cmd-on-actions")

	p2pMsg := envelope.Packet.P2PMessage
	switch m := p2pMsg.(type) {
	case *eos.SignedBlock:
		fmt.Printf("Receiving block %d sign from %s\n", m.BlockNumber(), m.Producer)

		doExec := false

		if blockModulo != 0 {
			if int(m.BlockNumber())%blockModulo == 0 {
				// run EXEC, block and continue after
				doExec = true
				goto runexec
			}
		}

		if actions != "" {
			for _, trx := range m.Transactions {
				unpacked, err := trx.Transaction.Packed.Unpack()
				if err != nil {
					fmt.Printf("Error unpacking transactions in block %d: %s\n", m.BlockNumber(), err)
					os.Exit(1)
				}

				for _, act := range unpacked.Transaction.Actions {
					actstr := fmt.Sprintf("%s:%s", act.Account, act.Name)
					if strings.Contains(actions, actstr) {
						doExec = true
						goto runexec
					}
				}
			}
		}

	runexec:
		if doExec {
			if err := runExec(); err != nil {
				fmt.Println("Error running exec:", err)
				os.Exit(1)
			}
		}
	}
})

func runExec() error {
	cmd := exec.Command(viper.GetString("tools-chain-freeze-cmd-exec-cmd"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}


// Convert a value to and from name-encoded strings
// EOS name encoding creates strings or up to 12 characters out of uint64 values.
// This command auto-detects encoding and converts it to different encodings.
func Names(input string) {

	showFrom := map[string]uint64{}

	// 将一个字符串账户名转换成一个uint64数字
	baseHex, err := hex.DecodeString(input)
	if err == nil {
		if len(baseHex) == 8 {
			showFrom["hex"] = binary.LittleEndian.Uint64(baseHex)
			showFrom["hex_be"] = binary.BigEndian.Uint64(baseHex)
		} else if len(baseHex) == 4 {
			showFrom["hex"] = uint64(binary.LittleEndian.Uint32(baseHex))
			showFrom["hex_be"] = uint64(binary.BigEndian.Uint32(baseHex))
		}
	}

	// 将一个字符串账户名转换成一个uint64数字
	fromName, err := eos.StringToName(input)
	if err == nil {
		showFrom["name"] = fromName
	}

	// 将一个字符串账户名转换成一个uint64数字
	fromUint64, err := strconv.ParseUint(input, 10, 64)
	if err == nil {
		showFrom["uint64"] = fromUint64
	}

	someFound := false
	rows := []string{"| from \\ to | hex | hex_be | name | uint64", "| --------- | --- | ------ | ---- | ------ |"}
	for _, from := range []string{"hex", "hex_be", "name", "uint64"} {
		val, found := showFrom[from]
		if !found {
			continue
		}
		someFound = true

		row := []string{from}
		for _, to := range []string{"hex", "hex_be", "name", "uint64"} {

			cnt := make([]byte, 8)
			switch to {
			case "hex":
				binary.LittleEndian.PutUint64(cnt, val)
				row = append(row, hex.EncodeToString(cnt))
			case "hex_be":
				binary.BigEndian.PutUint64(cnt, val)
				row = append(row, hex.EncodeToString(cnt))

			case "name":
				row = append(row, eos.NameToString(val))

			case "uint64":
				row = append(row, strconv.FormatUint(val, 10))
			}
		}
		rows = append(rows, "| "+strings.Join(row, " | ")+" |")
	}

	if !someFound {
		fmt.Printf("Couldn't decode %q with any of these methods: hex, hex_be, name, uint64\n", input)
		os.Exit(1)
	}

	fmt.Println("")
	fmt.Println(columnize.SimpleFormat(rows))
	fmt.Println("")
}

// Create a multisig transaction that both parties need to approve in order to do an atomic sale of your account.
// sell-account [sold account] [buyer account] [beneficiary account] [amount]
func SellAccount(soldAccount, buyerAccount, beneficiaryAccount eos.AccountName, amount string) {
	saleAmount, err := eos.NewEOSAssetFromString(amount)
	errorCheck(`sale "amount" invalid`, err)
	proposalName := viper.GetString("tools-sell-account-cmd-proposal-name")
	memo := viper.GetString("tools-sell-account-cmd-memo")

	api := getAPI()

	soldAccountData, err := api.GetAccount(soldAccount)
	errorCheck("could not find sold account on chain: "+string(soldAccount), err)

	if len(soldAccountData.Permissions) > 2 {
		fmt.Println("WARNING: your account has more than 2 permissions.")
		fmt.Println("This operation hands off control of `owner` and `active` keys.")
		fmt.Println("Please clean-up your permissions before selling your account.")
		os.Exit(1)
	}

	buyerAccountData, err := api.GetAccount(buyerAccount)
	errorCheck("could not find buyer's account on chain", err)

	_, err = api.GetAccount(beneficiaryAccount)
	errorCheck("could not find beneficiary's account on chain", err)

	// buyer的权限
	buyerPermText := viper.GetString("tools-sell-account-cmd-buyer-permission")
	if buyerPermText == "" {
		buyerPermText = string(buyerAccount)
	}
	buyerPerm, err := eos.NewPermissionLevel(buyerPermText)
	errorCheck(`invalid "buyer-permission"`, err)

	// seller的权限
	myPermText := viper.GetString("tools-sell-account-cmd-seller-permission")
	if myPermText == "" {
		myPermText = string(soldAccount)
	}
	myPerm, err := eos.NewPermissionLevel(myPermText)
	errorCheck(`invalid "seller-permission"`, err)

	targetOwnerAuth, err := sellAccountFindAuthority(buyerAccountData, "owner")
	errorCheck("error finding buyer's owner permission", err)
	targetActiveAuth, err := sellAccountFindAuthority(buyerAccountData, "active")
	errorCheck("error finding buyer's owner permission", err)

	infoResp, err := api.GetInfo() // 为当前头区块ID
	errorCheck("couldn't get_info from chain", err)

	tx := eos.NewTransaction([]*eos.Action{
		system.NewUpdateAuth(soldAccount, eos.PermissionName("owner"), eos.PermissionName(""), targetOwnerAuth, eos.PermissionName("owner")),
		system.NewUpdateAuth(soldAccount, eos.PermissionName("active"), eos.PermissionName("owner"), targetActiveAuth, eos.PermissionName("active")),
		token.NewTransfer(buyerAccount, beneficiaryAccount, saleAmount, memo),
	}, &eos.TxOptions{HeadBlockID: infoResp.HeadBlockID})
	tx.SetExpiration(viper.GetDuration("tools-sell-account-cmd-sale-expiration"))

	fmt.Println("Submitting `eosio.msig` proposal:")
	fmt.Printf("  proposer: %s\n", soldAccount)
	fmt.Printf("  proposal_name: %s\n", proposalName)
	fmt.Println("If this transaction is successful, have the other party approve and execute the multisig proposal to an atomic swap.")
	fmt.Println("Review this proposal with:")
	fmt.Printf("  eosc multisig review %s %s", soldAccount, proposalName)
	fmt.Println("")
	msigPermissions := []eos.PermissionLevel{buyerPerm, myPerm, eos.PermissionLevel{Actor: soldAccount, Permission: eos.PermissionName("owner")}}
	pushEOSCActions(api, msig.NewPropose(soldAccount, eos.Name(proposalName), msigPermissions, tx))
}

func sellAccountFindAuthority(data *eos.AccountResp, targetPerm string) (eos.Authority, error) {
	for _, perm := range data.Permissions {
		if perm.PermName == targetPerm {
			return perm.RequiredAuth, nil
		}
	}
	return eos.Authority{}, fmt.Errorf("permission %q not found in account %q", targetPerm, data.AccountName)
}