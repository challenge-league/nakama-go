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
	"fmt"

	log "github.com/micro/go-micro/v2/logger"
	"google.golang.org/protobuf/types/known/wrapperspb"

	nakamaContext "github.com/challenge-league/nakama-go/context"
	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

var cmdTournamentRecordAliases = []string{"tr"}

func getCmdTournamentRecordCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentRecordCreate := &cobra.Command{
		Use:     "tournamentRecord",
		Aliases: cmdTournamentRecordAliases,
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

			tournamentID, _ := cmd.Flags().GetString("tournamentID")
			score, _ := cmd.Flags().GetInt64("score")
			metadata, _ := cmd.Flags().GetString("metadata")
			subscore, _ := cmd.Flags().GetInt64("subscore")

			result, err := cmdBuilder.nakamaCtx.Client.WriteTournamentRecord(cmdBuilder.nakamaCtx.Ctx, &api.WriteTournamentRecordRequest{
				TournamentId: tournamentID,
				Record: &api.WriteTournamentRecordRequest_TournamentRecordWrite{
					Score:    score,
					Metadata: metadata,
					Subscore: subscore,
				},
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdTournamentRecordCreate.Flags().StringP("tournamentID", "i", "", "usage")
	cmdTournamentRecordCreate.Flags().StringP("metadata", "m", "", "usage")
	cmdTournamentRecordCreate.Flags().Int64P("score", "s", 0, "usage")
	cmdTournamentRecordCreate.Flags().Int64P("subscore", "", 0, "usage")
	return cmdTournamentRecordCreate
}

func getCmdTournamentRecordGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentRecordGet := &cobra.Command{
		Use:     "tournamentRecord",
		Aliases: cmdTournamentRecordAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			cursor, _ := cmd.Flags().GetString("cursor")
			expiry, _ := cmd.Flags().GetInt64("expiry")
			tournamentID, _ := cmd.Flags().GetString("tournamentID")
			limit, _ := cmd.Flags().GetInt32("limit")
			ownerIDs, _ := cmd.Flags().GetStringSlice("ownerIDs")
			if len(ownerIDs) == 0 {
				userData, _ := nakamaContext.UserDataFromSession(cmdBuilder.nakamaCtx.Session)
				ownerIDs = []string{userData["uid"].(string)}
			}

			result, err := cmdBuilder.nakamaCtx.Client.ListTournamentRecords(cmdBuilder.nakamaCtx.Ctx, &api.ListTournamentRecordsRequest{
				Cursor:       cursor,
				Expiry:       &wrapperspb.Int64Value{Value: expiry},
				TournamentId: tournamentID,
				Limit:        &wrapperspb.Int32Value{Value: limit},
				OwnerIds:     ownerIDs,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}

	cmdTournamentRecordGet.Flags().StringP("cursor", "c", "", "usage")
	cmdTournamentRecordGet.Flags().Int64P("expiry", "e", 0, "usage")
	cmdTournamentRecordGet.Flags().StringP("tournamentID", "i", "", "usage")
	cmdTournamentRecordGet.Flags().Int32P("limit", "l", 100, "usage")
	cmdTournamentRecordGet.Flags().StringSliceP("ownerIDs", "o", []string{}, "usage")
	return cmdTournamentRecordGet
}

func getCmdTournamentRecordAroundOwnerGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdTournamentRecordGet := &cobra.Command{
		Use: "tournamentRecordAroundOwner",
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
			tournamentID, _ := cmd.Flags().GetString("tournamentID")
			limit, _ := cmd.Flags().GetUint32("limit")

			result, err := cmdBuilder.nakamaCtx.Client.ListTournamentRecordsAroundOwner(cmdBuilder.nakamaCtx.Ctx, &api.ListTournamentRecordsAroundOwnerRequest{
				OwnerId:      ownerID,
				Expiry:       &wrapperspb.Int64Value{Value: expiry},
				TournamentId: tournamentID,
				Limit:        &wrapperspb.UInt32Value{Value: limit},
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}

	cmdTournamentRecordGet.Flags().StringP("ownerID", "o", "", "usage")
	cmdTournamentRecordGet.Flags().Int64P("expiry", "e", 0, "usage")
	cmdTournamentRecordGet.Flags().StringP("tournamentID", "i", "", "usage")
	cmdTournamentRecordGet.Flags().Uint32P("limit", "l", 100, "usage")
	return cmdTournamentRecordGet
}
