/*
Copyright © 2020 Dmitry Kozlov <dmtiry.f.kozlov@gmail.com>

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
	"strconv"
	"strings"

	log "github.com/micro/go-micro/v2/logger"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/heroiclabs/nakama-common/api"

	"github.com/spf13/cobra"
)

const (
	USER_DATA_COLLECTION = "user_data"
	USER_LAST_DATA_KEY   = "last"
)

// Custom types for the gRPC payload minimization
type DiscordUser struct {
	AuthorID      string
	ChannelID     string
	Discriminator string
	GuildID       string
	Username      string
	MessageID     string
}

type NakamaUser struct {
	CustomID    string
	Username    string
	ID          string
	DisplayName string
	Wallet      string
}

type User struct {
	Nakama  *NakamaUser
	Discord *DiscordUser
}

type UserData struct {
	UserID           string
	MatchID          string
	TicketID         string
	Version          string
	DiscordChannelID string
	DiscordGuildID   string
}

type LastUserDataCreateRequest struct {
	UserID   string
	UserData *UserData
}

type AccountGetRequest struct {
	Identifier string
}

type UserIdentity int

const (
	DiscordUndefined UserIdentity = iota
	DiscordUsernameWithDiscriminator
	DiscordID
	DiscordIDWithSpecialCharacters
)

func PrintAccount(account *api.Account) string {
	return ExecuteTemplate(
		fmt.Sprintf("> User: <@%v>\n", account.CustomId)+
			"```"+DISCORD_BLOCK_CODE_TYPE+"\n"+
			`UserID: {{.User.Id}}
DiscordID: {{.CustomId}}
Username: {{.User.Username}}
Created: {{.User.CreateTime | formatTimestampAsDate }}
Updated: {{.User.UpdateTime | formatTimestampAsDate }}`+"```\n"+
			fmt.Sprintf("> AvatarUrl: %v\n", account.User.AvatarUrl),
		account)
}

func createOrUpdateLastUserData(cmdBuilder *commandsBuilder, account *api.Account, userData *UserData) error {
	currentUserData, err := getLastUserData(cmdBuilder, account)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Infof("%+v", currentUserData)

	currentUserData = PatchStructByNewStruct(currentUserData, userData).(*UserData)
	if currentUserData != nil {
		if err := writeLastUserData(cmdBuilder, currentUserData); err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

func writeLastUserData(cmdBuilder *commandsBuilder, userData *UserData) error {
	log.Infof("%+v", userData)

	_, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "LastUserDataCreate", Payload: string(Marshal(&LastUserDataCreateRequest{
		UserID:   userData.UserID,
		UserData: userData,
	}))})
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func getUserData(cmdBuilder *commandsBuilder, key string, account *api.Account) (*UserData, error) {
	log.Info(account.User.Id)
	storageObjects, err := readUserStorageObjects(cmdBuilder, USER_DATA_COLLECTION, key, account.User.Id)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	metadata := make(map[string]string)
	if err := json.Unmarshal([]byte(account.User.Metadata), &metadata); err != nil {
		return nil, err
	}

	log.Info("%+v", metadata)

	discordChannelID := ""
	if v, ok := metadata["ChannelID"]; ok {
		discordChannelID = v
	}
	log.Info("%+v", discordChannelID)

	discordGuildID := ""
	if v, ok := metadata["GuildID"]; ok {
		discordGuildID = v
	}
	log.Info("%+v", discordGuildID)

	if len(storageObjects) == 0 {
		log.Error(fmt.Errorf("No user data found for %+v, trying to restore from account metadata", account))

		userData := &UserData{
			UserID:           account.User.Id,
			DiscordChannelID: discordChannelID,
			DiscordGuildID:   discordGuildID,
		}

		if err := writeLastUserData(cmdBuilder, userData); err != nil {
			log.Error(err)
			return nil, err
		}

		return userData, nil
	}

	var userData *UserData
	if err := json.Unmarshal([]byte(storageObjects[0].Value), &userData); err != nil {
		log.Error(err)
		return nil, err
	}
	userData.Version = storageObjects[0].Version

	if userData.DiscordChannelID == "" {
		userData.DiscordChannelID = discordChannelID
	}

	if userData.DiscordGuildID == "" {
		userData.DiscordGuildID = discordGuildID
	}

	return userData, nil
}

func getLastUserData(cmdBuilder *commandsBuilder, account *api.Account) (*UserData, error) {
	return getUserData(cmdBuilder, USER_LAST_DATA_KEY, account)
}

var cmdUserAliases = []string{"u"}

func UnmarshalTeamUser(value []byte) (*TeamUser, error) {
	var user *TeamUser
	err := json.Unmarshal(value, &user)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return user, nil
}

func GetTeamUsersFromMatch(matchState *MatchState) []*TeamUser {
	var teamUsers []*TeamUser
	for _, team := range matchState.Teams {
		for _, teamUser := range team.TeamUsers {
			teamUsers = append(teamUsers, teamUser)
		}
	}
	return teamUsers
}

func GetUsersFromMatch(matchState *MatchState) []*User {
	var users []*User
	for _, team := range matchState.Teams {
		for _, teamUser := range team.TeamUsers {
			users = append(users, teamUser.User)
		}
	}
	return users
}

func detectUserIdentity(identifier string) UserIdentity {
	if strings.Contains(identifier, "#") {
		return DiscordUsernameWithDiscriminator
	}

	if strings.HasPrefix(identifier, "<") && strings.Contains(identifier, "@") && strings.HasSuffix(identifier, ">") {
		return DiscordIDWithSpecialCharacters
	}

	if _, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		return DiscordID
	}

	return DiscordUndefined
}

func getAccount(cmdBuilder *commandsBuilder, identifier string) (*api.Account, error) {
	userIdentity := detectUserIdentity(identifier)
	switch userIdentity {
	case DiscordID:
		return getAccountByDiscordID(cmdBuilder, identifier)
	case DiscordIDWithSpecialCharacters:
		identifier = strings.ReplaceAll(identifier, "<", "")
		identifier = strings.ReplaceAll(identifier, "@", "")
		identifier = strings.ReplaceAll(identifier, ">", "")
		return getAccountByDiscordID(cmdBuilder, identifier)
	case DiscordUsernameWithDiscriminator:
		return getAccountByDiscordUsername(cmdBuilder, identifier)
	default:
		return nil, fmt.Errorf("Discord user ID is invalid")
	}
	return nil, nil
}

func getAccountByDiscordUsername(cmdBuilder *commandsBuilder, discordUsername string) (*api.Account, error) {
	var account *api.Account
	var err error
	payload, _ := json.Marshal(&AccountGetRequest{
		Identifier: discordUsername,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "AccountByUsernameGet", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(result.Payload), &account); err != nil {
		log.Error(err)
		return nil, err
	}
	return account, nil
}

func getAccountByDiscordID(cmdBuilder *commandsBuilder, discordID string) (*api.Account, error) {
	var account *api.Account
	var err error
	payload, _ := json.Marshal(&AccountGetRequest{
		Identifier: discordID,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "AccountByCustomIDGet", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(result.Payload), &account); err != nil {
		log.Error(err)
		return nil, err
	}
	return account, nil
}

func getCmdUserGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "user",
		Aliases: cmdUserAliases,
		Short:   "Get user",
		Long:    `Get user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			user, _ := cmd.Flags().GetString("user")
			account, err := getAccount(cmdBuilder, user)
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), PrintAccount(account))
			return nil
		},
	}
	cmd.Flags().StringP("user", "u", "", "Sepcify a specific user by the discord username#1234, @username or <@discord_user_id> (https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-)")
	return cmd
}

var cmdUserGroupsAliases = []string{"ug"}

func getCmdUserGroupsGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdUserGroupsGet := &cobra.Command{
		Use:     "userGroups",
		Aliases: cmdUserGroupsAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			userID, _ := cmd.Flags().GetString("userID")
			if userID == "" && len(args) > 0 {
				userID = args[0]
			}
			cursor, _ := cmd.Flags().GetString("cursor")
			state, _ := cmd.Flags().GetInt32("state")
			limit, _ := cmd.Flags().GetInt32("limit")
			result, err := cmdBuilder.nakamaCtx.Client.ListUserGroups(cmdBuilder.nakamaCtx.Ctx, &api.ListUserGroupsRequest{
				UserId: userID,
				State:  &wrapperspb.Int32Value{Value: state},
				Limit:  &wrapperspb.Int32Value{Value: limit},
				Cursor: cursor,
			})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			return nil
		},
	}
	cmdUserGroupsGet.Flags().StringP("userID", "i", "", "usage")
	cmdUserGroupsGet.Flags().StringP("cursor", "c", "", "usage")
	cmdUserGroupsGet.Flags().Int32P("state", "s", 0, "usage")
	cmdUserGroupsGet.Flags().Int32P("limit", "l", MAX_LIST_LIMIT, "usage")
	return cmdUserGroupsGet
}
