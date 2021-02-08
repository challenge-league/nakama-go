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
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/spf13/cobra"
)

func getCmdLogin(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdLogin := &cobra.Command{
		Use:     "login",
		Aliases: []string{"l"},
		Short:   "**Login** to the league by email or uid",
		Long:    "**Login** to the league by email or uid",
		RunE: func(cmd *cobra.Command, args []string) error {
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			if err := writeLastUserData(cmdBuilder, &UserData{
				UserID:           account.User.Id,
				DiscordChannelID: cmdBuilder.nakamaCtx.DiscordMsg.ChannelID,
				DiscordGuildID:   cmdBuilder.nakamaCtx.DiscordMsg.GuildID,
			}); err != nil {
				log.Error(err)
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("User <@%v> has successfully logged in", account.CustomId))
			return nil
		},
	}
	return cmdLogin
}
