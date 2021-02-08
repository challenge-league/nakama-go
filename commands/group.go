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
	"math"

	"google.golang.org/protobuf/types/known/wrapperspb"

	log "github.com/micro/go-micro/v2/logger"

	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

var cmdGroupAliases = []string{}

func getCmdGroupCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupCreate := &cobra.Command{
		Use:     "group ",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			langTag, _ := cmd.Flags().GetString("langTag")
			avatarUrl, _ := cmd.Flags().GetString("avatarUrl")
			open, _ := cmd.Flags().GetBool("open")
			maxCount, _ := cmd.Flags().GetInt32("maxCount")
			result, err := cmdBuilder.nakamaCtx.Client.CreateGroup(
				cmdBuilder.nakamaCtx.Ctx, &api.CreateGroupRequest{
					AvatarUrl:   avatarUrl,
					Description: desc,
					LangTag:     langTag,
					MaxCount:    maxCount,
					Name:        name,
					Open:        open,
				},
			)
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupCreate.Flags().StringP("name", "n", "", "usage")
	cmdGroupCreate.Flags().StringP("desc", "d", "", "usage")
	cmdGroupCreate.Flags().StringP("avatarUrl", "a", "", "usage")
	cmdGroupCreate.Flags().StringP("langTag", "l", "", "usage")
	cmdGroupCreate.Flags().BoolP("open", "o", true, "usage")
	cmdGroupCreate.Flags().Int32P("maxCount", "m", math.MaxInt32, "usage")
	return cmdGroupCreate
}

func getCmdGroupGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupGet := &cobra.Command{
		Use:     "group",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			name, _ := cmd.Flags().GetString("name")
			if name == "" && len(args) > 0 {
				name = args[0]
			}
			cursor, _ := cmd.Flags().GetString("cursor")
			limit, _ := cmd.Flags().GetInt32("limit")
			result, err := cmdBuilder.nakamaCtx.Client.ListGroups(cmdBuilder.nakamaCtx.Ctx, &api.ListGroupsRequest{
				Name:   name,
				Cursor: cursor,
				Limit:  &wrapperspb.Int32Value{Value: limit},
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupGet.Flags().StringP("name", "n", "", "usage")
	cmdGroupGet.Flags().StringP("cursor", "c", "", "usage")
	cmdGroupGet.Flags().Int32P("limit", "l", MAX_LIST_LIMIT, "usage")
	return cmdGroupGet
}

func getCmdGroupUpdate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupUpdate := &cobra.Command{
		Use:     "group",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			langTag, _ := cmd.Flags().GetString("langTag")
			avatarUrl, _ := cmd.Flags().GetString("avatarUrl")
			result, err := cmdBuilder.nakamaCtx.Client.UpdateGroup(
				cmdBuilder.nakamaCtx.Ctx, &api.UpdateGroupRequest{
					GroupId:     groupID,
					AvatarUrl:   &wrapperspb.StringValue{Value: avatarUrl},
					Description: &wrapperspb.StringValue{Value: desc},
					LangTag:     &wrapperspb.StringValue{Value: langTag},
					Name:        &wrapperspb.StringValue{Value: name},
				},
			)
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupUpdate.Flags().StringP("groupID", "g", "", "usage")
	cmdGroupUpdate.Flags().StringP("name", "n", "", "usage")
	cmdGroupUpdate.Flags().StringP("desc", "d", "", "usage")
	cmdGroupUpdate.Flags().StringP("avatarUrl", "a", "", "usage")
	cmdGroupUpdate.Flags().StringP("langTag", "l", "", "usage")
	return cmdGroupUpdate
}

func getCmdGroupDelete(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupDelete := &cobra.Command{
		Use:     "group",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			if groupID == "" && len(args) > 0 {
				groupID = args[0]
			}

			result, err := cmdBuilder.nakamaCtx.Client.DeleteGroup(cmdBuilder.nakamaCtx.Ctx, &api.DeleteGroupRequest{GroupId: groupID})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupDelete.Flags().StringP("groupID", "g", "", "usage")
	return cmdGroupDelete
}

func getCmdGroupJoin(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupJoin := &cobra.Command{
		Use:     "group",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			if groupID == "" && len(args) > 0 {
				groupID = args[0]
			}

			result, err := cmdBuilder.nakamaCtx.Client.JoinGroup(cmdBuilder.nakamaCtx.Ctx, &api.JoinGroupRequest{GroupId: groupID})
			if err != nil {
				log.Error(result)
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupJoin.Flags().StringP("groupID", "g", "", "usage")
	return cmdGroupJoin
}

func getCmdGroupLeave(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdGroupLeave := &cobra.Command{
		Use:     "group",
		Aliases: cmdGroupAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			groupID, _ := cmd.Flags().GetString("groupID")
			if groupID == "" && len(args) > 0 {
				groupID = args[0]
			}

			result, err := cmdBuilder.nakamaCtx.Client.LeaveGroup(cmdBuilder.nakamaCtx.Ctx, &api.LeaveGroupRequest{GroupId: groupID})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdGroupLeave.Flags().StringP("groupID", "g", "", "usage")
	return cmdGroupLeave
}
