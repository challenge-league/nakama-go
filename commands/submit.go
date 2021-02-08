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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	SUBMIT_COLLECTION = "submit_data"
)

type Submit struct {
	Datetime  time.Time
	Score     int64
	Subscore  int64
	ProofLink string
}

type Submits struct {
	Submits []*Submit
	MatchID string
	UserID  string
	Version string
}

type SubmitCreateRequest struct {
	Submit  *Submit
	MatchID string
	UserID  string
}

func PrintSubmit(submit *Submit) string {
	return ExecuteTemplate(
		"```"+DISCORD_BLOCK_CODE_TYPE+"\n"+
			`Score: {{.Score}}.{{.Subscore}} | Date: {{.Datetime | formatTimeAsDate}} | Proof: {{.ProofLink}}`+"```\n",
		submit)
}

func getCmdSubmit(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdSubmit := &cobra.Command{
		Use:     "submit [score] [proof_link] [matchID]",
		Aliases: []string{"score", "s"},
		Short:   "Submit a **new score** and a **proof link** to the match leaderboard",
		Long:    `Submit a **new score** and a **proof link** to the match leaderboard`,
		//Args:    matchAll(cobra.MinimumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			matchID, _ := cmd.Flags().GetString("matchID")
			scoreF, _ := cmd.Flags().GetFloat64("score")
			proof, _ := cmd.Flags().GetString("proof")
			var err error
			if scoreF == 0 && len(args) >= 2 {
				if scoreF, err = strconv.ParseFloat(args[0], 64); err != nil {
					log.Error(err)
					return err
				}
			}

			if scoreF == 0 {
				return fmt.Errorf("score is **required** and must not equal 0")
			}

			score := int64(scoreF)
			_, subscoreF := math.Modf(scoreF)
			subscore := int64(0)
			if subscoreF != 0 {
				subscore, err = strconv.ParseInt(fmt.Sprintf("%.10f", subscoreF)[2:], 10, 64)
				if err != nil {
					log.Error(err)
					return err
				}
			}

			if proof == "" && len(args) >= 2 {
				proof = args[1]
			}

			if proof != "" {
				if !strings.HasPrefix(proof, "http") {
					proof = "http://" + proof
				}
				if !IsValidUrl(proof) {
					return fmt.Errorf("'%v' is not a valid url for the proof link", proof)
				}
			} else {
				return fmt.Errorf("proof link is **required**")
			}

			if matchID == "" && len(args) == 3 {
				matchID = args[2]
			}

			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}

			var matchState *MatchState
			if matchID != "" {
				matchState, err = getMatchState(cmdBuilder, matchID, MATCH_COLLECTION)
			} else {
				matchState, err = getLastUserMatchState(cmdBuilder, account, MATCH_COLLECTION)
			}

			if err != nil {
				log.Error(err)
				return err
			}

			if matchState == nil {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No active matches found for <@%v>", account.CustomId))
				return nil
			}

			if !(matchState.Active && matchState.Started) {
				fmt.Fprintf(cmd.OutOrStdout(), "Match is not active or not started yet")
				return nil
			}

			payload, _ := json.Marshal(&SubmitCreateRequest{
				Submit: &Submit{
					ProofLink: proof,
					Score:     score,
					Subscore:  subscore,
				},
				MatchID: matchState.MatchID,
				UserID:  account.User.Id,
			})
			log.Infof("%+v\n", string(payload))

			_, err = cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "SubmitCreate", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			return nil
		},
	}
	cmdSubmit.Flags().StringP("matchID", "m", "", "Match ID")
	cmdSubmit.Flags().Float64P("score", "s", 0, "Score value **must** be positive float number. ")
	cmdSubmit.Flags().StringP("proof", "p", "", "Proof link **must** be a valid URL starting with **http** or **https**")
	return cmdSubmit
}
