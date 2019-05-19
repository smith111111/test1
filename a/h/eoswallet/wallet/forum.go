package wallet

import (
	"encoding/json"
	"fmt"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/forum"
	"github.com/pborman/uuid"
	"github.com/ryanuber/columnize"
	"github.com/spf13/viper"
	"github.com/tidwall/sjson"
	"strconv"
	"time"
	"errors"
)

// Cleans an expired proposal
func CleanProposal(cleaner, target eos.AccountName, proposalName eos.Name, maxCount uint64) {
	action := forum.NewCleanProposal(cleaner, proposalName, maxCount)
	//--target-contract
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// Allows the [proposer] to expire a proposal before its set expiration time.
func Expire(proposer, target eos.AccountName, proposalName eos.Name) {
	action := forum.NewExpire(proposer, proposalName)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// List forum proposals.
func List(proposerStr string, target eos.AccountName, printJson bool) {

	api := getAPI()
	var proposer eos.AccountName

	var err error
	var resp *eos.GetTableRowsResp
	if proposerStr != "" {
		proposer = toAccount(proposerStr, "--from-proposer")
		resp, err = api.GetTableRows(eos.GetTableRowsRequest{
			Code:       string(target),
			Scope:      string(target),
			Table:      string("proposal"),
			Index:      "sec", // Secondary index `by_proposer`
			KeyType:    "name",
			LowerBound: string(proposer),
			Limit:      1000,
			JSON:       true,
		})
		if err != nil {
			errorCheck(fmt.Sprintf("unable to get list of proposals from proposer %q", proposer), err)
		}
	} else {
		resp, err = api.GetTableRows(eos.GetTableRowsRequest{
			Code:  string(target),
			Scope: string(target),
			Table: string("proposal"),
			Limit: 1000,
			JSON:  true,
		})
		if err != nil {
			errorCheck("unable to get list of proposals", err)
		}
	}

	var proposals []struct {
		ProposalName eos.Name        `json:"proposal_name"`
		Proposer     eos.AccountName `json:"proposer"`
		Title        string          `json:"title"`
		ProposalJSON string          `json:"proposal_json"`
		CreatedAt    eos.JSONTime    `json:"created_at"`
		ExpiresAt    eos.JSONTime    `json:"expires_at"`
	}
	err = resp.JSONToStructs(&proposals)
	errorCheck("reading proposal list", err)

	if printJson {
		data, err := json.MarshalIndent(proposals, "", "  ")
		errorCheck("json marshal", err)
		fmt.Println(string(data))
		return
	}

	found := false
	for _, proposal := range proposals {
		if proposerStr == "" || proposal.Proposer == proposer {
			fmt.Println("Proposal name: ", proposal.ProposalName)
			fmt.Println("Proposer: ", proposal.Proposer)
			fmt.Println("Title: ", proposal.Title)
			fmt.Println("JSON: ", proposal.ProposalJSON)
			fmt.Println("Created at: ", proposal.CreatedAt)
			fmt.Println("Expires at: ", proposal.ExpiresAt)
			fmt.Println()

			found = true
		}
	}
	if !found {
		errorCheck("no proposal found", fmt.Errorf("empty list"))
	}
}

// Post a message
// post [poster] [content]
func Post(poster, target eos.AccountName, content string) {

	certify := viper.GetBool("forum-post-cmd-certify")
	newUUID := uuid.New()

	metadata := viper.GetString("forum-post-cmd-metadata")
	if metadata != "" {
		var dump interface{}
		err := json.Unmarshal([]byte(metadata), &dump)
		errorCheck("--metadata is not valid JSON", err)
	} else {
		metadataBytes, _ := json.Marshal(map[string]interface{}{
			"type": viper.GetString("forum-post-cmd-type"),
		})
		metadata = string(metadataBytes)
	}

	replyTo := eos.AccountName(viper.GetString("forum-post-cmd-reply-to"))
	if len(replyTo) != 0 {
		_ = toAccount(string(replyTo), "--reply-to") // only check for errors
	}

	replyToUUID := viper.GetString("forum-post-cmd-reply-to-uuid")

	action := forum.NewPost(poster, newUUID, content, replyTo, replyToUUID, certify, metadata)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// 提交一个建议为投票 Submit a proposition for votes
// propose [proposer] [proposal_name] [title] [proposal_expiration_date]
func Propose(proposer, target eos.AccountName, proposalName eos.Name, title string , expiresAt eos.JSONTime) {

	var err error
	if expiresAt.Before(time.Now()) {
		errorCheck("proposal expiration date must in the future", errors.New("provided time is in the past"))
	}

	proposalJSON := viper.GetString("forum-propose-cmd-json")
	content := viper.GetString("forum-propose-cmd-content")
	jsonType := viper.GetString("forum-propose-cmd-type")
	if proposalJSON == "" && content != "" {
		proposalJSON = "{}"
	}
	proposalJSON, err = sjson.Set(proposalJSON, "content", content)
	// Defaults JSON schema type to `bps-proposal-v1`
	proposalJSON, err = sjson.Set(proposalJSON, "type", jsonType)
	errorCheck("setting content in json", err)

	action := forum.NewPropose(proposer, proposalName, title, proposalJSON, expiresAt)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// Sets the status message for an account.
func Status(from, target eos.AccountName, content string) {
	action := forum.NewStatus(from, content)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// Tally votes according to the `type` of the proposal.
func TallyVotes(target eos.AccountName, proposalName eos.Name) {
	api := getAPI()

	votes, err := getForumVotesRows(api, target, proposalName)
	errorCheck("getting proposal", err)

	tallyStaked := make(map[uint8]int64)
	tallyAccounts := make(map[uint8]int64)
	var totalStaked int64
	for _, vote := range votes {
		tallyStaked[vote.vote.Vote] = tallyStaked[vote.vote.Vote] + int64(vote.account.VoterInfo.Staked)
		totalStaked += int64(vote.account.VoterInfo.Staked)
		tallyAccounts[vote.vote.Vote] = tallyAccounts[vote.vote.Vote] + 1

	}
	totalStakedEOS := eos.NewEOSAsset(totalStaked)

	fmt.Printf("Vote tally for proposal %q:\n", proposalName)
	fmt.Printf("* %d accounts voted\n", len(votes))
	fmt.Printf("* %s staked total\n", totalStakedEOS.String())

	output := []string{
		"Vote value | Num accounts | EOS staked",
		"---------- | ------------ | ----------",
	}
	for k, stakedForVote := range tallyStaked {
		accountsForVote := tallyAccounts[k]
		output = append(output, fmt.Sprintf("%d | %d | %s", k, accountsForVote, eos.NewEOSAsset(stakedForVote).String()))
	}
	fmt.Println(columnize.SimpleFormat(output))
}

func getForumVotesRows(api *eos.API, contract eos.AccountName, proposalName eos.Name) (out []*forumVoteEntry, err error) {
	// lowerBound := "first"
	// for {
	// 	// TODO: Optimize by querying the secondary index..
	// 	resp, err := api.GetTableRows(eos.GetTableRowsRequest{
	// 		Code:       string(contract),
	// 		Scope:      string(contract),
	// 		Table:      string("vote"),
	// 		Index:      "sec",  // Secondary Index `by_proposal` - https://github.com/eoscanada/eosio.forum/blob/master/include/forum.hpp#L99-L115
	// 		KeyType:    "i128", // `by_proposal` is uint128 - Compute as https://github.com/eoscanada/eosio.forum/blob/master/include/forum.hpp#L72-L74
	// 		LowerBound: "first",
	// 		Limit:      1000,
	// 		JSON:       true,
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	var votes []*forumVoteEntry
	// 	if err := json.Unmarshal(resp.Rows, &votes); err != nil {
	// 		return nil, err
	// 	}

	// 	for _, vote := range votes {
	// 		// TODO: optimize with getting only the bare minimum, like:
	// 		// cleosk get table eosio eosio voters -L cancancan234 --limit 1
	// 		acctResp, err := api.GetAccount(entry.vote.Voter)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		entry.account = acctResp

	// 		out = append(out, entry)
	// 	}

	// 	if !resp.More {
	// 		break
	// 	}
	// }

	return
}

type forumVoteEntry struct {
	vote    *forum.Vote
	account *eos.AccountResp
}

// Removes a given post.
// [poster] [post_uuid]
func UnPost(poster, target eos.AccountName, postUUID string) {
	action := forum.NewUnPost(poster, postUUID)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// Cancels a vote for a given proposal.
// [voter] [proposal_name]
func UnVote(voter, target eos.AccountName, proposalName eos.Name) {
	action := forum.NewUnVote(voter, proposalName)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}

// Submit a vote from [voter] on [proposal_name] with a [vote_value].
// vote [voter] [proposal_name] [vote_value]
func Vote(voter, target eos.AccountName, proposalName eos.Name, vote string) {
	// TODO: in a func
	if vote == "yes" {
		vote = "1"
	}
	if vote == "no" {
		vote = "0"
	}
	voteValue, err := strconv.ParseUint(vote, 10, 8)
	errorCheck("expected an integer for vote_value", err)
	if voteValue > 255 {
		errorCheck("vote value cannot exceed 255", fmt.Errorf("vote value too high: %d", voteValue))
	}

	json := viper.GetString("forum-cmd-target-json")

	action := forum.NewVote(voter, proposalName, uint8(voteValue), json)
	action.Account = target

	api := getAPI()
	pushEOSCActions(api, action)
}