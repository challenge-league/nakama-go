/*
Copyright © 2020 Dmitry Kozlov dmitry.f.kozlov@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package commands

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TeamResult struct {
	TeamNumber int
	Votes      int
}

type MatchResult struct {
	UserID     string
	DiscordID  string
	ProofLink  string
	TeamNumber int
	Win        bool
	Draw       bool
	DateTime   time.Time
}

type MatchResultRequest struct {
	MatchResult *MatchResult
	MatchID     string
}

func PrintMatchResults(matchState *MatchState) string {
	return ExecuteTemplate(
		`{{if .Results}}> Results: 
{{range $index, $element := .Results}}>   <@{{.DiscordID}}> {{if not .Draw }}**Team {{.TeamNumber}}** {{if .Win}}**Win**{{else}}**Lose**{{end}}{{else}}**Draw**{{end}} {{.ProofLink}} {{.DateTime | formatTimeAsDate}}
{{end}}`+"\n"+`{{end}}`,
		matchState)
}

func createResult(cmdBuilder *commandsBuilder, cmd *cobra.Command, args []string, win bool, draw bool) error {
	log.Infof("%+v\n", args)
	account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return err
	}
	matchID, _ := cmd.Flags().GetString("matchID")
	var matchState *MatchState
	if len(args) == 0 {
		matchState, err = getLastUserMatchState(cmdBuilder, account, MATCH_COLLECTION)
	}

	proof, _ := cmd.Flags().GetString("proof")

	if proof == "" {
		if draw && len(args) == 2 {
			proof = args[1]
		}

		if !draw && len(args) == 3 {
			proof = args[2]
		}
	}

	if proof != "" {
		if !strings.HasPrefix(proof, "http") {
			proof = "http://" + proof
		}
		if !IsValidUrl(proof) {
			return fmt.Errorf("'%v' is not a valid url for the proof link", proof)
		}
	}

	if matchID == "" {
		if !draw && len(args) >= 2 {
			matchState, err = getMatchState(cmdBuilder, args[1], MATCH_COLLECTION)
			if err != nil {
				log.Error(err)
				return err
			}
		}

		if draw && len(args) >= 1 {
			matchState, err = getMatchState(cmdBuilder, args[0], MATCH_COLLECTION)
			if err != nil {
				log.Error(err)
				return err
			}
		}
	}

	teamID, _ := cmd.Flags().GetInt("teamID")
	if !draw {
		if teamID == -1 && len(args) >= 1 {
			teamID, err = strconv.Atoi(args[0])
			if err == nil {
				log.Error(err)
				return err
			}
		}

		if teamID == -1 {
			teamID = GetTeamNumberFromUserAndMatch(account.User.Id, matchState)
		}

		if teamID < -1 || teamID > (len(matchState.Teams)-1) {
			return fmt.Errorf("Incorrect team number %v", teamID)
		}
	}

	if matchState == nil {
		return fmt.Errorf("No match found for <@%v>", account.CustomId)
	}

	payload, _ := json.Marshal(MatchResultRequest{
		MatchResult: &MatchResult{
			UserID:     account.User.Id,
			DiscordID:  account.CustomId,
			TeamNumber: teamID,
			Win:        win,
			Draw:       draw,
			ProofLink:  proof,
		},
		MatchID: matchState.MatchID,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchResult", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return err
	}
	if result.Payload != "" {
		fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
	}
	return nil
}

func setupResultFlags(cmd *cobra.Command, isDraw bool) {
	cmd.Flags().StringP("matchID", "m", "", "usage")
	cmd.Flags().StringP("proof", "p", "", "Proof link **must** be a valid URL starting with **http** or **https**")
	if !isDraw {
		cmd.Flags().IntP("teamID", "t", -1, "Team #")
	}
}

func getCmdResultWinCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "win [teamID] [matchID] [proof_link]",
		Aliases: []string{"victory"},
		Short:   "Report an early **win** before the match ends",
		Long:    `Report an early **win** before the match ends`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createResult(cmdBuilder, cmd, args, true, false)
		},
	}
	setupResultFlags(cmd, false)
	return cmd
}

func getCmdResultDrawCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "draw [matchID] [proof_link]",
		Aliases: []string{"victory"},
		Short:   "Report an early **draw** before the match ends",
		Long:    `Report an early **draw** before the match ends`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createResult(cmdBuilder, cmd, args, false, true)
		},
	}
	setupResultFlags(cmd, true)
	return cmd
}

func getCmdResultLossCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lose [teamID] [matchID] [proof_link]",
		Aliases: []string{"ff", "loss"},
		Short:   "Report an early **loss** before the match ends",
		Long:    `Report an early **loss** before the match ends`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createResult(cmdBuilder, cmd, args, false, false)
		},
	}
	setupResultFlags(cmd, false)
	return cmd
}
