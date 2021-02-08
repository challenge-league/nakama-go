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
	"strings"
	"time"

	log "github.com/micro/go-micro/v2/logger"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

type TournamentCreateRequest struct {
	ID            string
	SortOrder     string // one of: "desc", "asc"
	Operator      string // one of: "best", "set", "incr", "decr"
	ResetSchedule string
	Metadata      map[string]interface{}
	Title         string
	Description   string
	Category      int
	StartTime     int  // start now
	EndTime       int  // never end, repeat the tournament each day forever
	Duration      int  // in seconds
	MaxSize       int  // first 10,000 players who join
	MaxNumScore   int  // each player can have 3 attempts to score
	JoinRequired  bool // must join to compete
	Debug         bool `json:"debug"`
}

type TournamentDeleteRequest struct {
	ID string
}

var cmdTournamentAliases = []string{}

func getCmdTournamentCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentCreate := &cobra.Command{
		Use:     "tournament ",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		//Args:      matchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs),
		//ValidArgs: validArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			//initTournamentCreateFlags(cmd)
			err := cmd.ParseFlags(args)
			if err != nil {
				log.Error(err)
				return err
			}

			sortOrder, _ := cmd.Flags().GetString("sortOrder")
			operator, _ := cmd.Flags().GetString("operator")
			resetSchedule, err := cmd.Flags().GetString("resetSchedule")
			title, _ := cmd.Flags().GetString("title")
			desc, _ := cmd.Flags().GetString("desc")
			category, _ := cmd.Flags().GetInt("category")
			duration, _ := cmd.Flags().GetInt("duration")
			maxSize, _ := cmd.Flags().GetInt("maxSize")
			maxNumScore, _ := cmd.Flags().GetInt("maxNumScore")
			joinRequired, _ := cmd.Flags().GetBool("joinRequired")
			debug, _ := cmd.Flags().GetBool("debug")

			payload, _ := json.Marshal(&TournamentCreateRequest{
				ID:        uuid.Must(uuid.NewV4()).String(),
				SortOrder: sortOrder,
				Operator:  operator,
				//ResetSchedule: strings.ReplaceAll(strings.Join(resetSchedule, " "), "+", "*"),
				ResetSchedule: strings.ReplaceAll(resetSchedule, ",", " "),
				Metadata:      map[string]interface{}{},
				Title:         title,
				Description:   desc,
				Category:      category,
				StartTime:     int(time.Now().UTC().Unix()),
				EndTime:       0,
				Duration:      duration,
				MaxSize:       maxSize,
				MaxNumScore:   maxNumScore,
				JoinRequired:  joinRequired,
				Debug:         debug,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "TournamentCreate", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdTournamentCreate.Flags().StringP("sortOrder", "s", "desc", "usage")
	cmdTournamentCreate.Flags().StringP("operator", "o", "best", "usage")
	cmdTournamentCreate.Flags().StringP("resetSchedule", "r", "", "0,12,*,*,*")
	cmdTournamentCreate.Flags().StringP("title", "t", "", "usage")
	cmdTournamentCreate.MarkFlagRequired("title")
	cmdTournamentCreate.Flags().StringP("desc", "", "", "usage")
	cmdTournamentCreate.Flags().IntP("category", "c", 1, "usage")
	cmdTournamentCreate.Flags().IntP("duration", "", 3600, "usage")
	cmdTournamentCreate.Flags().IntP("maxSize", "", 10000, "usage")
	cmdTournamentCreate.Flags().IntP("maxNumScore", "", 3, "usage")
	cmdTournamentCreate.Flags().BoolP("joinRequired", "j", true, "usage")
	cmdTournamentCreate.Flags().BoolP("debug", "", true, "usage")
	return cmdTournamentCreate
}

func getCmdTournamentGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentGet := &cobra.Command{
		Use:     "tournament",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			categoryStart, _ := cmd.Flags().GetUint32("categoryStart")
			categoryEnd, _ := cmd.Flags().GetUint32("categoryEnd")
			timeStartString, _ := cmd.Flags().GetString("timeStart")
			timeEndString, _ := cmd.Flags().GetString("timeEnd")
			timeStartParsed, _ := time.Parse(TIME_LAYOUT, timeStartString)
			timeEndParsed, _ := time.Parse(TIME_LAYOUT, timeEndString)
			cursor, _ := cmd.Flags().GetString("cursor")
			limit, _ := cmd.Flags().GetInt32("limit")
			result, err := cmdBuilder.nakamaCtx.Client.ListTournaments(cmdBuilder.nakamaCtx.Ctx, &api.ListTournamentsRequest{
				CategoryStart: &wrapperspb.UInt32Value{Value: categoryStart},
				CategoryEnd:   &wrapperspb.UInt32Value{Value: categoryEnd},
				Cursor:        cursor,
				StartTime:     &wrapperspb.UInt32Value{Value: uint32(timeStartParsed.Unix())},
				EndTime:       &wrapperspb.UInt32Value{Value: uint32(timeEndParsed.Unix())},
				Limit:         &wrapperspb.Int32Value{Value: limit},
			})
			/*result, err = cmdBuilder.nakamaCtx.Client.ListTournaments(cmdBuilder.nakamaCtx.Ctx, &api.ListTournamentsRequest{
				CategoryStart: &wrapperspb.UInt32Value{Value: categoryStart},
				CategoryEnd:   &wrapperspb.UInt32Value{Value: categoryEnd},
				Cursor:        cursor,
				Limit:         &wrapperspb.Int32Value{Value: limit},
			})
			*/
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdTournamentGet.Flags().Uint32P("categoryStart", "", 0, "usage")
	cmdTournamentGet.Flags().Uint32P("categoryEnd", "", 127, "usage")
	cmdTournamentGet.Flags().StringP("cursor", "c", "", "usage")
	cmdTournamentGet.Flags().StringP("timeStart", "s", TIME_START_DEFAULT, "usage: "+TIME_LAYOUT)
	cmdTournamentGet.Flags().StringP("timeEnd", "e", TIME_END_DEFAULT, "usage: "+TIME_LAYOUT)
	cmdTournamentGet.Flags().Int32P("limit", "l", MAX_LIST_LIMIT, "usage")
	return cmdTournamentGet
}

func getCmdTournamentDelete(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentDelete := &cobra.Command{
		Use:     "tournament [id]",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		Args:    matchAll(cobra.MinimumNArgs(1)),
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

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "TournamentDelete", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdTournamentDelete.Flags().StringP("id", "i", "", "usage")
	return cmdTournamentDelete
}

func getCmdTournamentJoin(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentJoin := &cobra.Command{
		Use:     "tournament [id]",
		Aliases: cmdTournamentAliases,
		Short:   "Short",
		Long:    `Long`,
		Args:    matchAll(cobra.MinimumNArgs(1)),
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

			result, err := cmdBuilder.nakamaCtx.Client.JoinTournament(cmdBuilder.nakamaCtx.Ctx, &api.JoinTournamentRequest{
				TournamentId: id,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdTournamentJoin.Flags().StringP("id", "i", "", "usage")
	return cmdTournamentJoin
}
