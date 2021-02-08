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

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

type LeaderboardCreateRequest struct {
	ID            string
	Authoritative bool
	SortOrder     string
	Operator      string
	ResetSchedule string
	Metadata      map[string]interface{}
}

type LeaderboardDeleteRequest struct {
	ID string
}

var cmdLeaderboardAliases = []string{"l"}

func getCmdLeaderboardCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardCreate := &cobra.Command{
		Use:     "leaderboard ",
		Aliases: cmdLeaderboardAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			authoritative, _ := cmd.Flags().GetBool("authoritative")
			sortOrder, _ := cmd.Flags().GetString("sortOrder")
			operator, _ := cmd.Flags().GetString("operator")
			resetSchedule, _ := cmd.Flags().GetString("resetSchedule")
			payload, _ := json.Marshal(&LeaderboardCreateRequest{
				ID:            uuid.Must(uuid.NewV4()).String(),
				Authoritative: authoritative,
				Metadata:      map[string]interface{}{},
				Operator:      operator,
				ResetSchedule: resetSchedule,
				SortOrder:     sortOrder,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "LeaderboardCreate", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdLeaderboardCreate.Flags().BoolP("authoritative", "a", true, "usage")
	cmdLeaderboardCreate.Flags().StringP("sortOrder", "s", "desc", "usage")
	cmdLeaderboardCreate.Flags().StringP("operator", "o", "best", "usage")
	cmdLeaderboardCreate.Flags().StringP("resetSchedule", "r", "", "0 12 * * *")
	return cmdLeaderboardCreate
}

func getCmdLeaderboardDelete(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardDelete := &cobra.Command{
		Use:     "leaderboard",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			payload, _ := json.Marshal(&TournamentDeleteRequest{
				ID: id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "LeaderboardDelete", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdLeaderboardDelete.Flags().StringP("id", "i", "", "usage")
	return cmdLeaderboardDelete
}
