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

	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

var cmdGroupUsersAliases = []string{"gu"}

func getCmdGroupUsersAdd(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUsersAdd := &cobra.Command{
		Use:     "groupUsers",
		Aliases: cmdGroupUsersAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			userIDs, _ := cmd.Flags().GetStringSlice("userIDs")
			result, err := cmdBuilder.nakamaCtx.Client.AddGroupUsers(cmdBuilder.nakamaCtx.Ctx, &api.AddGroupUsersRequest{
				GroupId: groupID,
				UserIds: userIDs,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupUsersAdd.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUsersAdd.Flags().StringSliceP("userIDs", "u", []string{""}, "usage")
	return cmdGroupUsersAdd
}

func getCmdGroupUsersBan(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUsersBan := &cobra.Command{
		Use:     "groupUsers",
		Aliases: cmdGroupUsersAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			userIDs, _ := cmd.Flags().GetStringSlice("usersIDs")
			result, err := cmdBuilder.nakamaCtx.Client.BanGroupUsers(cmdBuilder.nakamaCtx.Ctx, &api.BanGroupUsersRequest{
				GroupId: groupID,
				UserIds: userIDs,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}

	cmdGroupUsersBan.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUsersBan.Flags().StringSliceP("userIDs", "u", []string{}, "usage")
	return cmdGroupUsersBan
}

func getCmdGroupUsersKick(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUsersKick := &cobra.Command{
		Use:     "groupUsers",
		Aliases: cmdGroupUsersAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			userIDs, _ := cmd.Flags().GetStringSlice("usersIDs")
			result, err := cmdBuilder.nakamaCtx.Client.KickGroupUsers(cmdBuilder.nakamaCtx.Ctx, &api.KickGroupUsersRequest{
				GroupId: groupID,
				UserIds: userIDs,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupUsersKick.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUsersKick.Flags().StringSliceP("userIDs", "u", []string{}, "usage")
	return cmdGroupUsersKick
}

func getCmdGroupUsersGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUsersGet := &cobra.Command{
		Use:     "groupUsers",
		Aliases: cmdGroupUsersAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			if groupID == "" && len(args) > 0 {
				groupID = args[0]
			}
			limit, _ := cmd.Flags().GetInt32("limit")
			cursor, _ := cmd.Flags().GetString("cursor")
			state, _ := cmd.Flags().GetInt32("state")
			result, err := cmdBuilder.nakamaCtx.Client.ListGroupUsers(cmdBuilder.nakamaCtx.Ctx, &api.ListGroupUsersRequest{
				GroupId: groupID,
				Limit:   &wrapperspb.Int32Value{Value: limit},
				Cursor:  cursor,
				State:   &wrapperspb.Int32Value{Value: state},
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupUsersGet.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUsersGet.Flags().Int32P("limit", "l", MAX_LIST_LIMIT, "usage")
	cmdGroupUsersGet.Flags().Int32P("state", "s", 0, "usage")
	cmdGroupUsersGet.Flags().StringP("cursor", "c", "", "usage")
	return cmdGroupUsersGet
}

func getCmdGroupUsersPromote(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUsersPromote := &cobra.Command{
		Use:     "groupUsers",
		Aliases: cmdGroupUsersAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			userIDs, _ := cmd.Flags().GetStringSlice("userIDs")
			result, err := cmdBuilder.nakamaCtx.Client.PromoteGroupUsers(cmdBuilder.nakamaCtx.Ctx, &api.PromoteGroupUsersRequest{
				GroupId: groupID,
				UserIds: userIDs,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupUsersPromote.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUsersPromote.Flags().StringSliceP("userIDs", "u", []string{}, "usage")
	return cmdGroupUsersPromote
}
