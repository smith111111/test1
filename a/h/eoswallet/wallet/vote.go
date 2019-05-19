package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/system"
	"github.com/spf13/viper"
	"sort"
	"strconv"
)

// Cancel all votes currently cast for producers/delegated to a proxy.
// cancel-all [voter name]
func CancelAll(voterName eos.AccountName) {
	api := getAPI()

	noProxy := eos.AccountName("")
	var noVotes []eos.AccountName
	pushEOSCActions(api,
		system.NewVoteProducer(
			voterName,
			noProxy,
			noVotes...,
		),
	)

	fmt.Printf("Consider using `eosc vote status %s` to confirm it has been applied.\n", voterName)
}


type producers []map[string]interface{}

func (p producers) Len() int      { return len(p) }
func (p producers) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p producers) Less(i, j int) bool {
	iv, _ := strconv.ParseFloat(p[i]["total_votes"].(string), 64)
	jv, _ := strconv.ParseFloat(p[j]["total_votes"].(string), 64)
	return iv > jv
}

// Retrieve the list of registered producers.
func ListProducers() {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Scope: "eosio",
			Code:  "eosio",
			Table: "producers",
			JSON:  true,
			Limit: 5000,
		},
	)
	errorCheck("get table rows", err)

	if viper.GetBool("vote-list-cmd-json") {
		data, err := json.MarshalIndent(response.Rows, "", "    ")
		errorCheck("json marshal", err)

		fmt.Println(string(data))
	} else {
		var producers producers
		err := json.Unmarshal(response.Rows, &producers)
		errorCheck("json marshaling", err)

		if viper.GetBool("vote-list-cmd-sort") {
			sort.Slice(producers, producers.Less)
		}

		fmt.Println("List of producers registered to receive votes:")
		for _, p := range producers {
			fmt.Printf("- %s (key: %s)\n", p["owner"], p["producer_key"])
		}
		fmt.Printf("Total of %d registered producers\n", len(producers))

	}
}

// Cast your vote for 1 to 30 producers. View them with 'list-producers'.
// producers [voter name] [producer list]
func Producers(voterName eos.AccountName, producerStringNames ...string) {
	sort.Strings(producerStringNames)

	var producerNames []eos.AccountName
	for _, producerString := range producerStringNames {
		producerNames = append(producerNames, toAccount(producerString, "producer list"))
	}

	api := getAPI()

	fmt.Printf("Voter [%s] voting for: %s\n", voterName, producerNames)
	pushEOSCActions(api,
		system.NewVoteProducer(
			voterName,
			"",
			producerNames...,
		),
	)
}

// Proxy your vote strength to a proxy.
// proxy [voter name] [proxy name]
func Proxy(voterName, proxyName eos.AccountName) {
	api := getAPI()
	fmt.Printf("Voter [%s] voting for proxy: %s\n", voterName, proxyName)

	pushEOSCActions(api,
		system.NewVoteProducer(
			voterName,
			proxyName,
		),
	)
}

// Recast your vote for the same producers or proxy.
// recast [voter name]
func Recast(voterName eos.AccountName) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:       "eosio",
			Scope:      "eosio",
			Table:      "voters",
			JSON:       true,
			LowerBound: string(voterName),
			Limit:      1,
		},
	)
	errorCheck("get table row", err)

	var voterInfos []eos.VoterInfo
	err = response.JSONToStructs(&voterInfos)
	errorCheck("reading voter_info", err)

	found := false
	for _, info := range voterInfos {
		if info.Owner == voterName {
			found = true
			if info.Proxy != "" {
				fmt.Printf("Voter [%s] recasting vote via proxy: %s\n", voterName, info.Proxy)
			} else {
				voterPrefix := ""
				if info.IsProxy != 0 {
					voterPrefix = "Proxy "
				}
				producersList := "no producer"
				if len(info.Producers) >= 1 {
					producersList = fmt.Sprint(info.Producers)
				}
				fmt.Printf("%sVoter [%s] recasting vote for: %s\n", voterPrefix, voterName, producersList)
			}
			pushEOSCActions(api,
				system.NewVoteProducer(
					voterName,
					info.Proxy,
					info.Producers...,
				),
			)
		}
	}
	if !found {
		errorCheck("vote recast", fmt.Errorf("unable to recast vote as no existing vote was found"))
	}
}

// Display the current vote status for a given account.
// status [voter name]
func VoteStatus(voterName eos.AccountName) {
	api := getAPI()

	response, err := api.GetTableRows(
		eos.GetTableRowsRequest{
			Code:       "eosio",
			Scope:      "eosio",
			Table:      "voters",
			JSON:       true,
			LowerBound: string(voterName),
			Limit:      1,
		},
	)
	errorCheck("get table row", err)

	var voterInfos []eos.VoterInfo
	err = response.JSONToStructs(&voterInfos)
	errorCheck("reading voter_info", err)

	found := false
	for _, info := range voterInfos {
		if info.Owner == voterName {
			found = true
			fmt.Println("Voter: ", info.Owner)

			if info.IsProxy != 0 {
				fmt.Println("Registered as a proxy voter: true")
				fmt.Println("Proxied vote weight: ", info.ProxiedVoteWeight)
			} else {
				fmt.Println("Registered as a proxy voter: false")
			}

			if info.Proxy != "" {
				fmt.Println("Voting via proxy: ", info.Proxy)
				fmt.Println("Last vote weight: ", info.LastVoteWeight)

			} else {
				fmt.Println("Producers list: ", info.Producers)
				fmt.Println("Staked amount: ", info.Staked)
				fmt.Printf("Last vote weight: %f\n", info.LastVoteWeight)
			}
		}
	}
	if !found {
		errorCheck("vote status", fmt.Errorf("unable to find vote status for %s", voterName))
	}
}