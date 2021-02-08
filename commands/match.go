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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/golang/protobuf/ptypes/wrappers"
	log "github.com/micro/go-micro/v2/logger"

	//	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/challenge-league/nakama-go/context"
	"github.com/heroiclabs/nakama-common/api"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/spf13/cobra"
)

const (
	DEFAULT_NAKAMA_MATCH_MODULE = "league"

	MATCH_COLLECTION         = "match_data"
	MATCH_ARCHIVE_COLLECTION = "match_archive_data"

	MATCH_PROFILE_1_VS_1 = "1vs1"
	MATCH_PROFILE_2_VS_2 = "2vs2"
	MATCH_PROFILE_3_VS_3 = "3vs3"
	MATCH_PROFILE_4_VS_4 = "4vs4"
	MATCH_PROFILE_5_VS_5 = "5vs5"

	MATCH_TYPE_MATCH_MAKER    = "match_maker"
	MATCH_TYPE_CAPTAINS_DRAFT = "captains_draft"

	MATCH_STATUS_CREATED                     = "Created"
	MATCH_STATUS_CAPTAINS_DRAFT_IN_PROGRESS  = "Players draft in progress"
	MATCH_STATUS_AWAITNG_USERS_READY         = "Awaiting users ready"
	MATCH_STATUS_IN_PROGRESS                 = "In progress"
	MATCH_STATUS_AWAITNG_RESULTS             = "Awaiting results"
	MATCH_STATUS_ENDED_AFTER_TIME_EXPIRED    = "The match ended after the time expired"
	MATCH_STATUS_COMPLETED_AHEAD_OF_SCHEDULE = "The match was completed ahead of schedule"
	MATCH_STATUS_CANCELED                    = "Canceled"

	SEARCH_MIN_DURATION = "minDuration"
	SEARCH_MAX_DURATION = "maxDuration"
	SEARCH_MIN_DATE     = "minDate"
	SEARCH_MAX_DATE     = "maxDate"

	DEFAULT_MATCH_DURATION_HOURS = 3
	MIN_MATCH_DURATION_HOURS     = 1
	MAX_MATCH_DURATION_HOURS     = 48
)

var (
	cmdMatchAliases   = []string{"m"}
	MATCH_MAKER_MODES = []string{
		MATCH_PROFILE_1_VS_1,
	}

	CAPTAINS_DRAFT_MODES_MAP = map[string]*CaptainsDraftMode{
		MATCH_PROFILE_1_VS_1: &CaptainsDraftMode{TeamCount: 2, UsersInTeam: 1},
		MATCH_PROFILE_2_VS_2: &CaptainsDraftMode{TeamCount: 2, UsersInTeam: 2, UsersPerCaptainTurn: []int{1, 1}},
		MATCH_PROFILE_3_VS_3: &CaptainsDraftMode{TeamCount: 2, UsersInTeam: 3, UsersPerCaptainTurn: []int{1, 2, 1}},
		MATCH_PROFILE_4_VS_4: &CaptainsDraftMode{TeamCount: 2, UsersInTeam: 4, UsersPerCaptainTurn: []int{1, 2, 2, 1}},
		MATCH_PROFILE_5_VS_5: &CaptainsDraftMode{TeamCount: 2, UsersInTeam: 5, UsersPerCaptainTurn: []int{1, 2, 2, 2, 1}},
	}

	CAPTAIN_DRAFT_MODES = GetKeysFromMap(CAPTAINS_DRAFT_MODES_MAP)
)

type CaptainsDraftMode struct {
	TeamCount           int
	UsersInTeam         int
	UsersPerCaptainTurn []int
}

type DiscordMessage struct {
	ID        string
	ChannelID string
	GuildID   string
}

type MatchState struct {
	Active                 bool
	CancelUserIDs          []string
	CaptainUserIDs         []string
	CaptainTurnUserID      string
	DateTimeStart          time.Time
	DateTimeEnd            time.Time
	Duration               time.Duration
	ActualDateTimeEnd      time.Time
	ActualDuration         time.Duration
	Debug                  bool
	Started                bool
	Status                 string
	Teams                  []*Team
	MatchID                string
	MatchProfile           string
	MatchType              string
	PoolUserIDs            []string
	PoolUserCustomIDs      []string
	Results                []*MatchResult
	ReadyUserIDs           []string
	StorageUserID          string
	StorageCollection      string
	Version                string
	DiscordChannels        []*DiscordChannel
	DiscordNewMatchMessage DiscordMessage
	MaxNumScore            int
}

type MatchCreateRequest struct {
	Module string
	Params map[string]interface{}
}

type MatchGetRequest struct {
	ID     string
	Params map[string]interface{}
}

type MatchPoolJoinRequest struct {
	MatchID string
	UserID  string
}

type MatchPoolPickRequest struct {
	MatchID       string
	CaptainUserID string
	UserID        string
}

type MatchStateGetRequest struct {
	ID                string
	StorageCollection string
}

type MatchStateListGetRequest struct {
	StorageCollection string
	Key               string
	UserID            string
}

func PrintNewMatchMessage(s *MatchState) string {
	return "> **New match found!**\n" +
		PrintMatchState(s) +
		"```" + DISCORD_BLOCK_CODE_TYPE + "\n" +
		"To start the game please type:" + "```\n" +
		//fmt.Sprintf("**dl ready** or **dl ready %v**", s.MatchID)
		fmt.Sprintf("> **dl ready**")
}

func PrintMatchState(matchState *MatchState) string {
	return ExecuteTemplate(
		"```"+DISCORD_BLOCK_CODE_TYPE+"\n"+`MatchID: {{.MatchID}}`+"```\n"+
			PrintTeams(matchState.Teams)+
			`> Active: {{if .Active}}**True**{{else}}**False**{{end}} 
> Mode: **{{.MatchProfile}}**
> Status: **{{.Status}}**
> Duration: **{{ .Duration | formatDuration }}**{{ if .Started }}
> Start date: **{{ .DateTimeStart | formatTimeAsDate }}**
> End date: **{{ .DateTimeEnd | formatTimeAsDate }}**
{{ if .Active }}> Elapsed time: **{{ .DateTimeStart | getDurationSinceDate | formatDuration }}**{{end}}{{ if .ActualDateTimeEnd | dateIsNotZero }}
> Actual end date: **{{ .ActualDateTimeEnd | formatTimeAsDate }}**{{end}}{{ if .ActualDuration }}
> Actual duration: **{{ .ActualDuration | formatDuration }}**{{end}}{{end}}`+"\n"+
			PrintDraftPool(matchState)+
			PrintMatchResults(matchState)+
			PrintMatchReadyUserIDs(matchState),
		matchState)
}

func PrintDraftPool(matchState *MatchState) string {
	return ExecuteTemplate(
		`
{{ if .MatchType | isCaptainsDraft }}
> Captain Draft mode: 
{{if .CaptainTurnUserID}}> CaptainTurnUserID: <@{{ .CaptainTurnUserID }}>{{end}}`+"\n"+`
{{if .CaptainUserIDs}}> Captain User IDs: 
{{range $index, $element := .CaptainUserIDs}}> <@{{.}}>
{{end}}{{end}}`+"\n"+`
{{if .PoolUserCustomIDs}}> Draft Pool User IDs: 
{{range $index, $element := .PoolUserCustomIDs}}> <@{{.}}>
{{end}}{{end}}`+"\n"+
			`{{end}}`,
		matchState)
}

/*
func PrintMatchReward(matchState *MatchState) string {
	return ExecuteTemplate(
		`> Reward:

		`
		matchState)
}
*/

func PrintMatchStateEmbed(matchState *MatchState) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Color:       0x00ff00,
		Timestamp:   string(time.Now().UTC().Unix()),
		Description: PrintMatchState(matchState),
	}
}

func getMatchState(cmdBuilder *commandsBuilder, matchID string, collection string) (*MatchState, error) {
	payload, _ := json.Marshal(&MatchStateGetRequest{
		ID:                matchID,
		StorageCollection: collection,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchStateGet", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var matchState *MatchState
	if err := json.Unmarshal([]byte(result.Payload), &matchState); err != nil {
		log.Error(err)
		return nil, err
	}

	return matchState, nil
}

func getLastUserMatchState(cmdBuilder *commandsBuilder, account *api.Account, collection string) (*MatchState, error) {
	userData, err := getLastUserData(cmdBuilder, account)

	if err != nil {
		log.Error(err)
		return nil, err
	}
	if userData == nil {
		return nil, nil
	}
	if userData.MatchID == "" {
		return nil, nil
	}
	return getMatchState(cmdBuilder, userData.MatchID, collection)
}

func getMatchStateList(cmdBuilder *commandsBuilder, collection string) ([]*MatchState, error) {
	payload, _ := json.Marshal(&MatchStateListGetRequest{
		StorageCollection: collection,
		UserID:            context.NakamaSystemUserID,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchStateListGet", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	var matchStateList []*MatchState
	if result.Payload == "" {
		return matchStateList, nil
	}

	if err := json.Unmarshal([]byte(result.Payload), &matchStateList); err != nil {
		log.Error(err)
		return nil, err
	}

	return matchStateList, nil
}

func getCmdMatchCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdMatchCreate := &cobra.Command{
		Use:     "match ",
		Aliases: cmdMatchAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			module, _ := cmd.Flags().GetString("module")
			payload, _ := json.Marshal(&MatchCreateRequest{
				Module: module,
				Params: map[string]interface{}{},
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchCreate", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))
			return nil
		},
	}
	cmdMatchCreate.Flags().StringP("module", "m", DEFAULT_NAKAMA_MATCH_MODULE, "usage")
	return cmdMatchCreate
}

func getCmdMatchGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdMatchGet := &cobra.Command{
		Use:     "match ",
		Aliases: cmdMatchAliases,
		Short:   "Get match state",
		Long:    `Get match state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("matchID")
			all, _ := cmd.Flags().GetBool("all")
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			if all {
				matchStateList, err := getMatchStateList(cmdBuilder, "")
				if err != nil {
					log.Error(err)
					return err
				}
				if len(matchStateList) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No match found for user <@%v>", account.CustomId))
					return nil
				}
				for _, matchState := range matchStateList {
					if err != nil {
						log.Error(err)
						return err
					}

					fmt.Fprintf(cmd.OutOrStdout(), PrintMatchState(matchState))
				}
				return nil
			}
			var matchState *MatchState
			if id != "" {
				matchState, err = getMatchState(cmdBuilder, id, "")

			} else {
				matchState, err = getLastUserMatchState(cmdBuilder, account, "")
			}
			if err != nil {
				log.Error(err)
				return err
			}
			if matchState == nil {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No match found for user <@%v>", account.CustomId))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), PrintMatchState(matchState))
			}
			return nil
		},
	}
	cmdMatchGet.Flags().StringP("matchID", "m", "", "usage")
	cmdMatchGet.Flags().BoolP("all", "a", false, "-all to get all active matches")

	return cmdMatchGet
}

func getCmdMatchGet2(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdMatchGet := &cobra.Command{
		Use:     "match ",
		Aliases: cmdMatchAliases,
		Short:   "Short",
		Long:    `Long`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("matchID")
			limit, _ := cmd.Flags().GetInt32("limit")
			authoritative, _ := cmd.Flags().GetBool("authoritative")
			label, _ := cmd.Flags().GetString("label")
			minSize, _ := cmd.Flags().GetInt32("minSize")
			maxSize, _ := cmd.Flags().GetInt32("maxSize")
			query, _ := cmd.Flags().GetString("query")
			if id != "" {
				payload, _ := json.Marshal(&MatchGetRequest{
					ID: id,
				})
				log.Infof("%+v\n", string(payload))

				result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchGet", Payload: string(payload)})
				if err != nil {
					log.Error(err)
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result.Payload))

			} else {
				result, err := cmdBuilder.nakamaCtx.Client.ListMatches(cmdBuilder.nakamaCtx.Ctx, &api.ListMatchesRequest{
					Limit:         &wrapperspb.Int32Value{Value: limit},
					Authoritative: &wrappers.BoolValue{Value: authoritative},
					Label:         &wrappers.StringValue{Value: label},
					MinSize:       &wrappers.Int32Value{Value: minSize},
					MaxSize:       &wrappers.Int32Value{Value: maxSize},
					Query:         &wrappers.StringValue{Value: query},
				})

				if err != nil {
					log.Error(err)
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(result))
			}
			return nil
		},
	}
	cmdMatchGet.Flags().StringP("matchID", "m", "", "usage")
	cmdMatchGet.Flags().Int32P("limit", "l", MAX_LIST_LIMIT, "usage")
	cmdMatchGet.Flags().BoolP("authoritative", "a", true, "usage")
	cmdMatchGet.Flags().StringP("label", "", "", "usage")
	cmdMatchGet.Flags().Int32P("minSize", "", 0, "usage")
	cmdMatchGet.Flags().Int32P("maxSize", "", 0, "usage")
	cmdMatchGet.Flags().StringP("query", "q", "", "usage")
	return cmdMatchGet
}

func getCmdJoinMatchPool(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "join ",
		Aliases: []string{"j"},
		Short:   "Join to the captains draft pool by the MatchID",
		Long:    `Join to the captains draft pool by the MatchID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return poolJoin(cmdBuilder, cmd, args, true)
		},
	}
	cmd.Flags().StringP("matchID", "m", "", "usage")
	return cmd
}

func getCmdPickUserFromMatchPool(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pick ",
		Aliases: []string{"p"},
		Short:   "Pick user from the captains draft pool by the UserID",
		Long:    `Pick user from the captains draft pool by the UserID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)

			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}

			captainUserID := account.User.Id
			ticketState, err := getLastUserTicketState(cmdBuilder, account)
			matchID := ticketState.MatchID

			userID, _ := cmd.Flags().GetString("userID")
			if userID == "" && len(args) > 0 {
				userID = args[0]
				log.Infof("%+s", userID)
			} else {
				return fmt.Errorf("Please specify the UserID to pick from the draft pool")
			}
			account, err = getAccount(cmdBuilder, userID)
			if err != nil {
				log.Error(err)
				return err
			}
			log.Infof("%+s %+s", matchID, account)

			payload, _ := json.Marshal(MatchPoolPickRequest{
				MatchID:       matchID,
				CaptainUserID: captainUserID,
				UserID:        account.User.Id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "PoolPick", Payload: string(payload)})
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
	cmd.Flags().StringP("userID", "u", "", "usage")
	return cmd
}

func getCmdAddUserToMatchPool(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add ",
		Aliases: []string{"a"},
		Short:   "Add user to the captains draft pool by the UserID",
		Long:    `Add user to the captains draft pool by the UserID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return poolJoin(cmdBuilder, cmd, args, false)
		},
	}
	cmd.Flags().StringP("userID", "u", "", "usage")
	return cmd
}

func poolJoin(cmdBuilder *commandsBuilder, cmd *cobra.Command, args []string, isJoin bool) error {
	log.Infof("%+v\n", args)

	account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return err
	}

	matchID, _ := cmd.Flags().GetString("matchID")
	log.Infof("%+s", matchID)
	var ticketState *TicketState
	log.Infof("%+s", isJoin)
	if isJoin {
		if matchID == "" && len(args) > 0 {
			matchID = args[0]
			log.Infof("%+s", matchID)
		} else {
			return fmt.Errorf("Please specify the MatchID to join the captains draft pool")
		}
		log.Infof("%+s", matchID)
	} else {
		ticketState, err = getLastUserTicketState(cmdBuilder, account)
		matchID = ticketState.MatchID

		userID, _ := cmd.Flags().GetString("userID")
		if userID == "" && len(args) > 0 {
			userID = args[0]
			log.Infof("%+s", userID)
		} else {
			return fmt.Errorf("Please specify the UserID to add to the draft pool")
		}
		account, err = getAccount(cmdBuilder, userID)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	log.Infof("%+s %+s", matchID, account)

	ticketState, err = getLastUserTicketState(cmdBuilder, account)
	if err != nil {
		log.Error(err)
		return err
	}

	if ticketState != nil {
		fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("Existing ticket found for <@%v>, please complete the game", account.CustomId))
		return nil
	}

	ticketState, err = createCaptainsDraftTicketState(cmdBuilder, cmd, account, matchID, true)
	if err != nil {
		log.Error(err)
		return err
	}

	payload, _ := json.Marshal(MatchPoolJoinRequest{
		MatchID: matchID,
		UserID:  account.User.Id,
	})
	log.Infof("%+v\n", string(payload))

	result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "PoolJoin", Payload: string(payload)})
	if err != nil {
		log.Error(err)
		return err
	}
	if result.Payload != "" {
		fmt.Fprintf(cmd.OutOrStdout(), result.Payload)
	}
	return nil
}
