package wallet

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/msig"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"gcwallet/eoswallet/analysis"
)

// Approve a transaction in the eosio.msig contract
// approve [proposer] [proposal name] [approver[@active]]
func Approve(proposer eos.AccountName, proposalName eos.Name, approver string) {
	api := getAPI()

	requested, err := permissionToPermissionLevel(approver)
	if err != nil {
		fmt.Printf("Error with requested permission: %s\n", err)
		os.Exit(1)
	}

	pushEOSCActions(api, msig.NewApprove(proposer, proposalName, requested))
}

// Cancel a transaction in the eosio.msig contract"
// cancel [proposer] [proposal name] [canceler]
func Cancel(proposer, canceler eos.AccountName, proposalName eos.Name) {
	api := getAPI()

	pushEOSCActions(api,
		msig.NewCancel(proposer, proposalName, canceler),
	)
}

// Execute a transaction in the eosio.msig contract
// exec [proposer] [proposal name] [executer]
func Exec(proposer, executer eos.AccountName, proposalName eos.Name) {
	api := getAPI()

	pushEOSCActions(api,
		msig.NewExec(proposer, proposalName, executer),
	)
}

// Shows the list of all active proposals for a given proposer in the eosio.msig contract.
func ListProposals(proposer eos.AccountName, printJson bool) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:  "eosio.msig",
			Scope: string(proposer),
			Table: "approvals",
			JSON:  true,
		},
	)
	errorCheck("get table row", err)

	var approvalsInfo []struct {
		ProposalName       eos.Name              `json:"proposal_name"`
		RequestedApprovals []eos.PermissionLevel `json:"requested_approvals"`
		ProvidedApprovals  []eos.PermissionLevel `json:"provided_approvals"`
	}
	err = response.JSONToStructs(&approvalsInfo)
	errorCheck("reading approvals_info list", err)

	if printJson {
		data, err := json.MarshalIndent(approvalsInfo, "", "  ")
		errorCheck("json marshal", err)
		fmt.Println(string(data))
		return
	}

	for _, info := range approvalsInfo {
		fmt.Println("Proposal name:", info.ProposalName)
		fmt.Println("Requested approvals:", info.RequestedApprovals)
		fmt.Println("Provided approvals:", info.ProvidedApprovals)
		fmt.Println()
	}
	if len(approvalsInfo) == 0 {
		errorCheck("No multisig proposal found", fmt.Errorf("not found"))
	}
}

// Propose a new transaction in the eosio.msig contract Pass --requested-permissions
func NewPropose(proposer eos.AccountName, proposalName eos.Name, transactionFileName string) {
	api := getAPI()

	cnt, err := ioutil.ReadFile(transactionFileName)
	errorCheck("reading transaction file", err)

	var tx *eos.Transaction
	err = json.Unmarshal(cnt, &tx)
	errorCheck("parsing transaction file", err)

	requested, err := permissionsToPermissionLevels(viper.GetStringSlice("msig-propose-cmd-requested-permissions"))
	errorCheck("requested permissions", err)
	if len(requested) == 0 {
		errorCheck("--requested-permissions", errors.New("missing values"))
	}

	pushEOSCActions(api,
		msig.NewPropose(proposer, proposalName, requested, tx),
	)
}

// Review a proposal in the eosio.msig contract
func ReviewPropose(proposer eos.AccountName, proposalName eos.Name) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:       "eosio.msig",
			Scope:      string(proposer),
			Table:      "proposal",
			JSON:       true,
			LowerBound: string(proposalName),
			Limit:      1,
		},
	)
	errorCheck("get table row", err)

	var transactions []struct {
		ProposalName eos.Name     `json:"proposal_name"`
		Transaction  eos.HexBytes `json:"packed_transaction"`
	}
	err = response.JSONToStructs(&transactions)
	errorCheck("reading proposed transactions", err)

	var tx *eos.Transaction
	for _, txData := range transactions {
		if txData.ProposalName == proposalName {
			err := eos.UnmarshalBinary(txData.Transaction, &tx)
			errorCheck("unmarshalling packed transaction", err)

			ana := analysis.NewAnalyzer(viper.GetBool("msig-review-cmd-dump"))
			ana.API = api
			err = ana.AnalyzeTransaction(tx)
			errorCheck("analyzing", err)

			fmt.Println("Proposer:", proposer)
			fmt.Println("Proposal name:", proposalName)
			fmt.Println()
			os.Stdout.Write(ana.Writer.Bytes())
		}
	}
	if tx == nil {
		errorCheck("multisig proposal", fmt.Errorf("not found"))
	}
}

// Shows the status of a given proposal and its approvals in the eosio.msig contract.
func ProposalStatus(proposer eos.AccountName, proposalName eos.Name, printJson bool) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:       "eosio.msig",
			Scope:      string(proposer),
			Table:      "approvals",
			JSON:       true,
			LowerBound: string(proposalName),
			Limit:      1,
		},
	)
	errorCheck("get table row", err)

	var approvalsInfo []struct {
		ProposalName       eos.Name              `json:"proposal_name"`
		RequestedApprovals []eos.PermissionLevel `json:"requested_approvals"`
		ProvidedApprovals  []eos.PermissionLevel `json:"provided_approvals"`
	}
	err = response.JSONToStructs(&approvalsInfo)
	errorCheck("reading approvals_info", err)

	var found bool
	for _, info := range approvalsInfo {
		if info.ProposalName == proposalName {
			found = true

			if printJson {
				data, err := json.MarshalIndent(info, "", "  ")
				errorCheck("json marshal", err)
				fmt.Println(string(data))
			} else {
				fmt.Println("Proposer:", proposer)
				fmt.Println("Proposal name:", proposalName)
				fmt.Println("Requested approvals:", info.RequestedApprovals)
				fmt.Println("Provided approvals:", info.ProvidedApprovals)
				fmt.Println()
			}
		}
	}
	if !found {
		errorCheck("multisig proposal", fmt.Errorf("not found"))
	}
}

// Unapprove a transaction in the eosio.msig contract
// unapprove [proposer] [proposal name] [actor@permission]
func Unapprove(proposer eos.AccountName, proposalName eos.Name, permission string) {
	api := getAPI()

	requested, err := permissionToPermissionLevel(permission)
	errorCheck("requested permission", err)

	pushEOSCActions(api,
		msig.NewUnapprove(proposer, proposalName, requested),
	)
}
