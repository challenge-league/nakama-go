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

	"github.com/heroiclabs/nakama-common/api"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MatchReadyRequest struct {
	MatchID string
	UserID  string
}

type UserReady struct {
	Ready     bool
	UserID    string
	DiscordID string
}

func PrintMatchReadyUserIDs(matchState *MatchState) string {
	teamUsersReady := GetUsersReady(matchState)
	if matchState.Status == MATCH_STATUS_AWAITNG_USERS_READY {
		msg := ""
		if len(teamUsersReady) > 0 {
			msg += "> Ready:\n"
			for teamNumber, usersReady := range teamUsersReady {
				log.Infof("%+v %+v", teamNumber, usersReady)
				msg += fmt.Sprintf("> Team **%v**: %v\n", teamNumber, PrintUsersReady(usersReady))
			}

		}

		notReadyCount := 0
		for _, usersReady := range teamUsersReady {
			for _, user := range usersReady {
				if !user.Ready {
					notReadyCount += 1
				}
			}
		}

		if notReadyCount > 0 {
			msg += "\n> Not ready:\n"
			for teamNumber, usersReady := range teamUsersReady {
				msg += fmt.Sprintf("> Team **%v**: %v\n", teamNumber, PrintUsersNotReady(usersReady))
			}
		}
		return msg + "\n"
	}
	return ""
}

func PrintUsersReady(usersReady []*UserReady) string {
	log.Infof("%+v", usersReady)
	return ExecuteTemplate(
		`{{range $index, $element := .}} {{if .Ready}}**<@{{.DiscordID}}>**{{end}}{{end}}`,
		usersReady)
}

func PrintUsersNotReady(usersReady []*UserReady) string {
	log.Infof("%+v", usersReady)
	return ExecuteTemplate(
		`{{range $index, $element := .}} {{if not .Ready}}<@{{.DiscordID}}>{{end}}{{end}}`,
		usersReady)
}

func getCmdReady(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdReady := &cobra.Command{
		Use:     "ready [id]",
		Aliases: []string{"r"},
		Short:   "Indicate the readiness for a new match",
		Long: `Indicate the readiness for a new match 
Player can be ready even before a new match is found`,
		//Args:    matchAll(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			ticketID, _ := cmd.Flags().GetString("ticketID")
			var ticketState *TicketState
			if ticketID == "" && len(args) > 0 {
				ticketState, err = getTicketState(cmdBuilder, args[0], account)
			} else {
				ticketState, err = getLastUserTicketState(cmdBuilder, account)
			}
			if err != nil {
				log.Error(err)
				return err
			}
			if ticketState == nil {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No tickets found for <@%v>", account.CustomId))
				return nil
			}
			if ticketState.MatchID == "" {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("Ticket **%v** is not assigned to any match", ticketState.Ticket.Id))
				return nil
			}

			payload, _ := json.Marshal(MatchReadyRequest{
				MatchID: ticketState.MatchID,
				UserID:  account.User.Id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchReady", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			if result.Payload != "" {
				fmt.Fprintf(cmd.OutOrStdout(), result.Payload)
			}
			return nil
		},
	}
	cmdReady.Flags().StringP("ticketID", "t", "", "usage")
	return cmdReady
}
