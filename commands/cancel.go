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

	log "github.com/micro/go-micro/v2/logger"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/heroiclabs/nakama-common/api"
	"open-match.dev/open-match/pkg/pb"

	"github.com/spf13/cobra"
)

type MatchCancelRequest struct {
	MatchID string
	UserID  string
}

func getCmdCancel(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use: "cancel [ticketID]",
		//Aliases: []string{"r"},
		Short: "**Cancel** the ticket for a new match if the match has not started yet",
		Long:  `**Cancel** the ticket for a new match if the match has not started yet`,
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

			if len(args) == 0 {
				ticketState, err = getLastUserTicketState(cmdBuilder, account)
			}

			if ticketID == "" && len(args) > 0 {
				ticketState, err = getTicketState(cmdBuilder, args[0], account)
			}

			if err != nil {
				log.Error(err)
				return err
			}

			if ticketState == nil {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No tickets found for <@%v>", account.CustomId))
				return nil
			}

			if ticketState.MatchID != "" {
				payload, _ := json.Marshal(MatchCancelRequest{
					MatchID: ticketState.MatchID,
					UserID:  account.User.Id,
				})
				log.Infof("%+v\n", string(payload))

				result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchCancel", Payload: string(payload)})
				if err != nil {
					log.Error(err)
					return err
				}

				if result.Payload != "" {
					fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
				}
				return nil
			}

			err = deleteTicketState(cmdBuilder, ticketState.Ticket.Id, ticketState.UserID)
			if err != nil {
				log.Error(err)
				return err
			}

			payload, _ := json.Marshal(pb.DeleteTicketRequest{
				TicketId: ticketState.Ticket.Id,
			})
			log.Infof("%+v\n", string(payload))
			fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("Ticket **%v** was not assigned to any match, just deleting it", ticketState.Ticket.Id))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "OpenMatchFrontendTicketDelete", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			if result.Payload != "" {
				fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			}

			if err := createOrUpdateLastUserData(cmdBuilder, account, &UserData{
				UserID:   account.User.Id,
				MatchID:  PATCH_NULL_VALUE,
				TicketID: PATCH_NULL_VALUE,
			}); err != nil {
				log.Error(err)
				return err
			}

			return nil
		},
	}
	cmd.Flags().StringP("ticketID", "t", "", "usage")
	return cmd
}
