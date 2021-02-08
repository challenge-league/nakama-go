/*
Copyright © 2020 Dmitry Kozlov <dmitry.f.kozlov@gmail.com>

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
	"google.golang.org/protobuf/types/known/wrapperspb"

	nakamaContext "github.com/challenge-league/nakama-go/context"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/spf13/cobra"
)

const (
	MAIN_LEADERBOARD = "Main Leaderboard"
)

type LeaderboardRecordWriteRequest struct {
	ID       string
	OwnerID  string
	Username string
	Score    int64
	Subscore int64
	Metadata map[string]interface{}
}

type LeaderboardRecordDeleteRequest struct {
	ID      string
	OwnerID string
}

var cmdLeaderboardRecordAliases = []string{"leaderboard"}

func PrintLeaderboardRecords(leaderboardRecords []*api.LeaderboardRecord) string {
	return fmt.Sprintf("> Match Leaderboard **%v** :\n", leaderboardRecords[0].LeaderboardId) +
		ExecuteTemplate(`{{range $index, $element := .}}> <@{{.Username.Value}}> {{.Score}}.{{.Subscore}}
{{end}}`+"\n", leaderboardRecords)
}

func getCmdLeaderboardRecordAdd(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardRecordAdd := &cobra.Command{
		Use:     "leaderboardRecord",
		Aliases: cmdLeaderboardRecordAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			leaderboardID, _ := cmd.Flags().GetString("leaderboardID")
			score, _ := cmd.Flags().GetInt64("score")
			subscore, _ := cmd.Flags().GetInt64("subscore")
			userID, _ := cmd.Flags().GetString("userID")
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			if userID == "" {
				userID = account.User.Id
			}
			username, _ := cmd.Flags().GetString("username")
			payload, _ := json.Marshal(&LeaderboardRecordWriteRequest{
				ID:       leaderboardID,
				Score:    score,
				Subscore: subscore,
				OwnerID:  userID,
				Username: username,
				Metadata: map[string]interface{}{},
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "LeaderboardRecordWrite", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdLeaderboardRecordAdd.Flags().StringP("leaderboardID", "i", "desc", "usage")
	cmdLeaderboardRecordAdd.Flags().StringP("userID", "o", "desc", "usage")
	cmdLeaderboardRecordAdd.Flags().StringP("username", "u", "desc", "usage")
	cmdLeaderboardRecordAdd.Flags().Int64P("score", "s", 0, "usage")
	cmdLeaderboardRecordAdd.Flags().Int64P("subscore", "", 0, "usage")
	return cmdLeaderboardRecordAdd
}

func getCmdLeaderboardRecordDelete(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardRecordDelete := &cobra.Command{
		Use:     "leaderboardRecord",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			ownerID, _ := cmd.Flags().GetString("ownerID")
			if ownerID == "" {
				userData, _ := nakamaContext.UserDataFromSession(cmdBuilder.nakamaCtx.Session)
				ownerID = userData["uid"].(string)
			}
			payload, _ := json.Marshal(&LeaderboardRecordDeleteRequest{
				ID:      id,
				OwnerID: ownerID,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "LeaderboardRecordDelete", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}

	cmdLeaderboardRecordDelete.Flags().StringP("leaderboardID", "i", "", "usage")
	cmdLeaderboardRecordDelete.Flags().StringP("ownerID", "o", "", "usage")
	return cmdLeaderboardRecordDelete
}

func getCmdLeaderboardRecordsGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardRecordGet := &cobra.Command{
		Use:     "lb [matchID]",
		Aliases: cmdTournamentAliases,
		Short:   "Get the **match leaderboard**",
		Long:    `Get the **match leaderboard**`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			//cursor, _ := cmd.Flags().GetString("cursor")
			//expiry, _ := cmd.Flags().GetInt64("expiry")
			matchID, _ := cmd.Flags().GetString("matchID")
			if matchID == "" && len(args) > 0 {
				matchID = args[0]
			}
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}

			//limit, _ := cmd.Flags().GetInt32("limit")
			//ownerIDs, _ := cmd.Flags().GetStringSlice("ownerIDs")

			var matchState *MatchState
			if matchID != "" {
				matchState, err = getMatchState(cmdBuilder, matchID, "")
			} else {
				matchState, err = getLastUserMatchState(cmdBuilder, account, "")
			}
			if err != nil {
				log.Error(err)
				return err
			}
			if matchState == nil {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No match found for user <@%v>", account.CustomId))
				return nil
			}

			log.Infof("%+v", matchState)
			//result, err := cmdBuilder.nakamaCtx.Client.ListLeaderboardRecords(cmdBuilder.nakamaCtx.Ctx, &api.ListLeaderboardRecordsRequest{
			result, err := cmdBuilder.nakamaCtx.Client.ListLeaderboardRecordsAroundOwner(cmdBuilder.nakamaCtx.Ctx, &api.ListLeaderboardRecordsAroundOwnerRequest{
				LeaderboardId: matchState.MatchID,
				Limit:         &wrapperspb.UInt32Value{Value: MAX_LIST_LIMIT},
				OwnerId:       account.User.Id,
				//Expiry:        &wrapperspb.Int64Value{Value: 1},
			})
			/*
				 ListLeaderboardRecords(cmdBuilder.nakamaCtx.Ctx, &api.ListLeaderboardRecordsRequest{
					Cursor:        cursor,
					Expiry:        &wrapperspb.Int64Value{Value: 0},
					LeaderboardId: matchState.MatchID,
					Limit:         &wrapperspb.Int32Value{Value: limit},
					//OwnerIds:      matchState.,
				})
			*/
			if err != nil {
				log.Error(err)
				return err
			}

			if len(result.GetRecords()) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), PrintLeaderboardRecords(result.GetRecords()))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No leaderboard records found for match **%v**", matchState.MatchID))
			}
			return nil
		},
	}

	//cmdLeaderboardRecordGet.Flags().StringP("cursor", "c", "", "usage")
	//cmdLeaderboardRecordGet.Flags().Int64P("expiry", "e", 0, "usage")
	cmdLeaderboardRecordGet.Flags().StringP("matchID", "m", "", "usage")
	//cmdLeaderboardRecordGet.Flags().Int32P("limit", "l", 100, "usage")
	//cmdLeaderboardRecordGet.Flags().StringSliceP("ownerIDs", "o", []string{}, "usage")
	return cmdLeaderboardRecordGet
}

func getCmdTopLeaderboardRecordsGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "top [discordID]",
		Aliases: cmdTournamentAliases,
		Short:   "Get the **league leaderboard**",
		Long:    `Get the **league leaderboard**`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			discordID, _ := cmd.Flags().GetString("discordID")
			if discordID != "" {
				payload, _ := json.Marshal(&AccountGetRequest{
					Identifier: discordID,
				})
				log.Infof("%+v\n", string(payload))

				result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "AccountByCustomIDGet", Payload: string(payload)})
				if err != nil {
					log.Error(err)
					return err
				}

				if err := json.Unmarshal([]byte(result.Payload), &account); err != nil {
					log.Error(err)
					return err
				}
			}

			result, err := cmdBuilder.nakamaCtx.Client.ListLeaderboardRecordsAroundOwner(cmdBuilder.nakamaCtx.Ctx, &api.ListLeaderboardRecordsAroundOwnerRequest{
				LeaderboardId: MAIN_LEADERBOARD,
				Limit:         &wrapperspb.UInt32Value{Value: MAX_LIST_LIMIT},
				OwnerId:       account.User.Id,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			if len(result.GetRecords()) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), PrintLeaderboardRecords(result.GetRecords()))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No leaderboard records found for the %v", MAIN_LEADERBOARD))
			}
			return nil
		},
	}

	cmd.Flags().StringP("discordID", "d", "", "usage")
	return cmd
}

/*
func getCmdLeaderboardRecordAroundOwnerGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLeaderboardRecordGet := &cobra.Command{
		Use: "leaderboardRecordAroundOwner",
		//Aliases: cmdTournamentAliases,
		Short: "Short",
		Long:  `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			ownerID, _ := cmd.Flags().GetString("ownerID")
			if ownerID == "" {
				userData, _ := nakamaContext.UserDataFromSession(cmdBuilder.nakamaCtx.Session)
				ownerID = userData["uid"].(string)
			}
			expiry, _ := cmd.Flags().GetInt64("expiry")
			leaderboardID, _ := cmd.Flags().GetString("leaderboardID")
			limit, _ := cmd.Flags().GetUint32("limit")

			result, err := cmdBuilder.nakamaCtx.Client.ListLeaderboardRecordsAroundOwner(cmdBuilder.nakamaCtx.Ctx, &api.ListLeaderboardRecordsAroundOwnerRequest{
				OwnerId:       ownerID,
				Expiry:        &wrapperspb.Int64Value{Value: expiry},
				LeaderboardId: leaderboardID,
				Limit:         &wrapperspb.UInt32Value{Value: limit},
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}

	cmdLeaderboardRecordGet.Flags().StringP("ownerID", "o", "", "usage")
	cmdLeaderboardRecordGet.Flags().Int64P("expiry", "e", 0, "usage")
	cmdLeaderboardRecordGet.Flags().StringP("leaderboardID", "i", "", "usage")
	cmdLeaderboardRecordGet.Flags().Uint32P("limit", "l", 100, "usage")
	return cmdLeaderboardRecordGet
}
*/
