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
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/heroiclabs/nakama-common/api"
	"open-match.dev/open-match/pkg/pb"

	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
)

const (
	TICKET_COLLECTION          = "ticket_data"
	TICKET_EXTENSION_USER      = "user"
	MATCH_EXTENSION_MATCH_TYPE = "match_type"
)

var cmdTicketAliases = []string{"go", "new", "search", "ticket", "t"}

type TicketState struct {
	Ticket        *pb.Ticket
	CaptainsDraft bool
	MatchID       string
	UserID        string
	DiscordID     string
	Version       string
	UserReady     bool
}

type TicketStateCreateRequest struct {
	TicketState *TicketState
	UserID      string
}

func getTicketState(cmdBuilder *commandsBuilder, ticketID string, account *api.Account) (*TicketState, error) {
	storageObjects, err := listUserStorageObjects(cmdBuilder, TICKET_COLLECTION, account.User.Id, "")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if storageObjects == nil {
		return nil, fmt.Errorf("No tickets found for <@%v>", account.CustomId)
	}

	var ticketState *TicketState
	if err := json.Unmarshal([]byte(storageObjects[0].Value), &ticketState); err != nil {
		log.Error(err)
		return nil, err
	}

	return ticketState, nil
}

func deleteTicketState(cmdBuilder *commandsBuilder, ticketID string, userID string) error {
	_, err := cmdBuilder.nakamaCtx.Client.DeleteStorageObjects(cmdBuilder.nakamaCtx.Ctx, &api.DeleteStorageObjectsRequest{
		ObjectIds: []*api.DeleteStorageObjectId{
			&api.DeleteStorageObjectId{
				Collection: TICKET_COLLECTION,
				Key:        ticketID,
			},
		},
	})

	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func getLastUserTicketState(cmdBuilder *commandsBuilder, account *api.Account) (*TicketState, error) {
	userData, err := getLastUserData(cmdBuilder, account)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Info(userData)
	if userData == nil {
		return nil, nil
	}

	if userData.TicketID == "" {
		return nil, nil
	}

	return getTicketState(cmdBuilder, userData.TicketID, account)
}

func PrintTicketState(ticketState *TicketState) string {
	return ExecuteTemplate(
		fmt.Sprintf("> User: <@%v>\n", ticketState.DiscordID)+
			"```"+DISCORD_BLOCK_CODE_TYPE+"\n"+
			`TicketID: {{.Ticket.Id}}
MatchID: {{if .MatchID}}{{.MatchID}}{{else}}Not assigned{{end}}
SearchFields: `+PrintTicketSearchFields(ticketState)+` 
CreateTime: {{.Ticket.CreateTime | formatTimestampAsDate }}`+"```\n",
		ticketState)
}

func PrintTicketSearchFields(TicketState *TicketState) string {
	return ExecuteTemplate(
		`{{ .Tags }} {{ .DoubleArgs.maxDuration }} hours, `,
		TicketState.Ticket.SearchFields)
}

func createCaptainsDraftTicketState(cmdBuilder *commandsBuilder, cmd *cobra.Command, account *api.Account, matchID string, ready bool) (*TicketState, error) {
	log.Infof("%+v", account)
	duration, _ := cmd.Flags().GetInt("duration")
	tags, _ := cmd.Flags().GetStringSlice("mode")
	userData, err := getLastUserData(cmdBuilder, account)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	ticketState := &TicketState{
		Ticket: &pb.Ticket{
			Id:         uuid.Must(uuid.NewV4()).String(),
			CreateTime: &timestamppb.Timestamp{Seconds: time.Now().UTC().Unix()},
			Extensions: map[string]*anypb.Any{
				TICKET_EXTENSION_USER: &anypb.Any{Value: Marshal(
					&TeamUser{
						User: &User{
							Discord: &DiscordUser{
								AuthorID:      account.CustomId,
								Username:      strings.Split(account.User.Username, "#")[0],
								ChannelID:     userData.DiscordChannelID,
								Discriminator: strings.Split(account.User.Username, "#")[1],
								GuildID:       userData.DiscordGuildID,
							},
							Nakama: &NakamaUser{
								CustomID:    account.CustomId,
								DisplayName: account.User.DisplayName,
								ID:          account.User.Id,
								Username:    account.User.Username,
								Wallet:      account.Wallet,
							},
						},
					},
				)},
			},
			SearchFields: &pb.SearchFields{
				DoubleArgs: map[string]float64{
					"minDuration": float64(duration),
					"maxDuration": float64(duration),
				},
				Tags: tags,
			},
		},
		MatchID:       matchID,
		Version:       "*",
		DiscordID:     account.CustomId,
		UserID:        account.User.Id,
		UserReady:     ready,
		CaptainsDraft: true,
	}

	if _, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "TicketStateCreate", Payload: string(Marshal(&TicketStateCreateRequest{
		UserID:      account.User.Id,
		TicketState: ticketState,
	}))}); err != nil {
		log.Error(err)
		return nil, err
	}

	if err := createOrUpdateLastUserData(cmdBuilder, account, &UserData{
		UserID:   account.User.Id,
		MatchID:  ticketState.MatchID,
		TicketID: ticketState.Ticket.Id,
	}); err != nil {
		log.Error(err)
		return nil, err
	}

	return ticketState, nil
}

func createTicket(cmdBuilder *commandsBuilder, cmd *cobra.Command, args []string, isCaptainsDraftMode bool) error {
	log.Infof("%+v\n", args)
	matchMode, _ := cmd.Flags().GetStringSlice("mode")
	if isCaptainsDraftMode {
		if !IsStringInSlice(matchMode[0], CAPTAIN_DRAFT_MODES) {
			return fmt.Errorf("Match mode %v is invalid. Available match modes: %+v", matchMode[0], CAPTAIN_DRAFT_MODES)
		}
	} else {
		if !IsStringInSlice(matchMode[0], CAPTAIN_DRAFT_MODES) {
			return fmt.Errorf("Match mode %v is invalid. Available match modes: %+v", matchMode[0], CAPTAIN_DRAFT_MODES)
		}
	}
	ready, _ := cmd.Flags().GetBool("ready")
	//startDate, _ := cmd.Flags().GetFloat64(SEARCH_MIN_DATE)
	//endDate, _ := cmd.Flags().GetFloat64(SEARCH_MAX_DATE)
	//minDuration, _ := cmd.Flags().GetFloat64(SEARCH_MIN_DURATION)
	//maxDuration, _ := cmd.Flags().GetFloat64(SEARCH_MAX_DURATION)
	//maxDuration, _ := cmd.Flags().GetFloat64(SEARCH_MAX_DURATION)

	duration, _ := cmd.Flags().GetInt("duration")
	if duration < MIN_MATCH_DURATION_HOURS {
		return fmt.Errorf(fmt.Sprintf("duration can not be less than %v hours", MIN_MATCH_DURATION_HOURS))
	}

	if duration > MAX_MATCH_DURATION_HOURS {
		return fmt.Errorf(fmt.Sprint("duration can not be more than %v hours - this is not a Kaggle", MAX_MATCH_DURATION_HOURS))
	}

	account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return err
	}

	lastTicketState, err := getLastUserTicketState(cmdBuilder, account)
	if err != nil {
		log.Error(err)
		return err
	}

	if lastTicketState != nil {
		fmt.Fprintf(cmd.OutOrStdout(),
			fmt.Sprintf("<@%v> already has a ticket. Please cancel the following ticket or finish the following match:\n"+
				PrintTicketState(lastTicketState), account.CustomId))
		return nil
	}

	if isCaptainsDraftMode {
		user, _ := cmd.Flags().GetString("user")

		if user == "" && len(args) > 0 {
			user = args[0]
		} else {
			return fmt.Errorf("Please specify the opponent user ID to challenge him in the Captains Draft mode.")
		}

		opponentAccount, err := getAccount(cmdBuilder, user)
		if err != nil {
			log.Error(err)
			return err
		}
		if opponentAccount == nil {
			return fmt.Errorf("Opponent account not found")
		}
		if opponentAccount.User.Id == account.User.Id {
			return fmt.Errorf("You have selected yourself as your opponent's account. Please select a different opponent account.")
		}

		if user != "" {
			log.Infof("%+v", cmdBuilder.nakamaCtx.DiscordMsg)
		}
		lastOpponentTicketState, err := getLastUserTicketState(cmdBuilder, opponentAccount)
		if err != nil {
			log.Error(err)
			return err
		}

		if lastOpponentTicketState != nil {
			fmt.Fprintf(cmd.OutOrStdout(),
				fmt.Sprintf("<@%v> already has a ticket. Please cancel the following ticket or finish the following match:\n"+
					PrintTicketState(lastOpponentTicketState), opponentAccount.CustomId))
			return nil
		}

		matchID := uuid.Must(uuid.NewV4()).String()

		ticketState, err := createCaptainsDraftTicketState(cmdBuilder, cmd, account, matchID, ready)
		if err != nil {
			log.Error(err)
			return err
		}
		opponentTicketState, err := createCaptainsDraftTicketState(cmdBuilder, cmd, opponentAccount, matchID, false)
		if err != nil {
			log.Error(err)
			return err
		}

		var tickets []*pb.Ticket
		tickets = append(tickets, ticketState.Ticket)
		tickets = append(tickets, opponentTicketState.Ticket)

		match := &pb.Match{
			MatchId:      matchID,
			MatchProfile: matchMode[0],
			Tickets:      tickets,
			Extensions: map[string]*anypb.Any{
				MATCH_EXTENSION_MATCH_TYPE: &anypb.Any{Value: Marshal(MATCH_TYPE_CAPTAINS_DRAFT)},
			},
		}

		result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "MatchCreate", Payload: string(Marshal(match))})
		if err != nil {
			log.Error(err)
			return err
		}
		log.Infof("%+v\n", MarshalIndent(result))

		fmt.Fprintf(cmd.OutOrStdout(), PrintTicketState(ticketState))

	} else {
		userData, err := getLastUserData(cmdBuilder, account)
		if err != nil {
			log.Error(err)
			return err
		}
		//ticketStringArgs := make(map[string]string)
		//ticketDoubleArgs := make(map[string]float64)

		doubleArgs := make(map[string]float64)
		//doubleArgs[SEARCH_MIN_DATE] = startDate
		//doubleArgs[SEARCH_END_DATE] = endDate
		//doubleArgs[SEARCH_MIN_DURATION] = minDuration
		//doubleArgs[SEARCH_MAX_DURATION] = maxDuration
		doubleArgs[SEARCH_MIN_DURATION] = float64(duration)
		doubleArgs[SEARCH_MAX_DURATION] = float64(duration)
		payload, _ := json.Marshal(&pb.CreateTicketRequest{
			Ticket: &pb.Ticket{
				SearchFields: &pb.SearchFields{
					Tags: matchMode,
					//StringArgs: ,
					DoubleArgs: doubleArgs,
				},
				Extensions: map[string]*anypb.Any{
					TICKET_EXTENSION_USER: &anypb.Any{Value: Marshal(
						&TeamUser{
							User: &User{
								Discord: &DiscordUser{
									AuthorID:      account.CustomId,
									Username:      strings.Split(account.User.Username, "#")[0],
									ChannelID:     userData.DiscordChannelID,
									Discriminator: strings.Split(account.User.Username, "#")[1],
									GuildID:       userData.DiscordGuildID,
								},
								Nakama: &NakamaUser{
									CustomID:    account.CustomId,
									DisplayName: account.User.DisplayName,
									ID:          account.User.Id,
									Username:    account.User.Username,
									Wallet:      account.Wallet,
								},
							},
						},
					)},
				},
			},
		})
		log.Infof("%+v\n", string(payload))

		result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "OpenMatchFrontendTicketCreate", Payload: string(payload)})
		if err != nil {
			log.Error(err)
			return err
		}
		var ticket *pb.Ticket
		json.Unmarshal([]byte(result.Payload), &ticket)

		ticketState := &TicketState{
			Ticket:    ticket,
			MatchID:   "",
			Version:   "*",
			DiscordID: account.CustomId,
			UserID:    account.User.Id,
			UserReady: ready,
		}

		if _, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "TicketStateCreate", Payload: string(Marshal(&TicketStateCreateRequest{
			UserID:      account.User.Id,
			TicketState: ticketState,
		}))}); err != nil {
			log.Error(err)
			return err
		}

		if err := createOrUpdateLastUserData(cmdBuilder, account, &UserData{
			UserID:   account.User.Id,
			MatchID:  PATCH_NULL_VALUE,
			TicketID: ticket.Id,
		}); err != nil {
			log.Error(err)
			return err

		}
		fmt.Fprintf(cmd.OutOrStdout(), PrintTicketState(ticketState))
	}

	return nil
}

func setupTicketFlags(cmd *cobra.Command, isCaptainsDraft bool) {
	defaultMatchProfile := MATCH_PROFILE_1_VS_1
	if isCaptainsDraft {
		defaultMatchProfile = MATCH_PROFILE_2_VS_2
	}
	cmd.Flags().BoolP("ready", "r", true, "Indicate an early readiness for a new match")
	//cmd.Flags().Float64P(SEARCH_MIN_DATE, "s", 1, "start duration")
	//cmd.Flags().Float64P(SEARCH_MAX_DATE, "e", 24, "end duration")
	//cmd.Flags().Float64P(SEARCH_MIN_DURATION, "", 0, "min duration in hours")
	//cmd.Flags().Float64P(SEARCH_MAX_DURATION, "", 48, "max duration in hours")
	cmd.Flags().IntP("duration", "d", DEFAULT_MATCH_DURATION_HOURS, fmt.Sprintf("duration in hours, maximum duration is %v hours", MAX_MATCH_DURATION_HOURS))
	if isCaptainsDraft {
		cmd.Flags().StringP("user", "u", "", "**Challenge** a specific user by the discord username#1234, @username or <@discord_user_id> (https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID-)")
		cmd.Flags().StringSliceP("mode", "m", []string{defaultMatchProfile}, fmt.Sprintf("Captains Draft mode. Available modes: %+v", CAPTAIN_DRAFT_MODES))
	} else {
		cmd.Flags().StringSliceP("mode", "m", []string{defaultMatchProfile}, fmt.Sprintf("Match Maker mode. Available modes: %+v", MATCH_MAKER_MODES))
	}
}

func getCmdTicketCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "go ",
		Aliases: cmdTicketAliases,
		Short:   "Search for a **New Quick Match** and receive a ticket ID",
		Long:    `Search for a **New Quick Match** and receive a ticket ID`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createTicket(cmdBuilder, cmd, args, false)
		},
	}
	setupTicketFlags(cmd, false)
	return cmd
}

func getCmdCaptainsDraftCreate(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "challenge [user]",
		Aliases: []string{"chal", "chall", "c"},
		Short:   "**Challenge** a specific user in the Captains Draft mode.",
		Long:    `**Challenge** a specific user in the Captains Draft mode.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createTicket(cmdBuilder, cmd, args, true)
		},
	}
	setupTicketFlags(cmd, true)
	return cmd
}

func getTicketStateList(cmdBuilder *commandsBuilder, account *api.Account) ([]*TicketState, error) {

	objects, err := listUserStorageObjects(cmdBuilder, TICKET_COLLECTION, account.User.Id, "")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var ticketStateList []*TicketState
	for _, v := range objects {
		ticketState, err := getTicketState(cmdBuilder, v.Key, account)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		ticketStateList = append(ticketStateList, ticketState)
	}
	return ticketStateList, nil
}

func getCmdTicketGet(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket [id]",
		Aliases: cmdTicketAliases,
		Short:   "Get ticket state",
		Long:    `Get ticket state`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			all, _ := cmd.Flags().GetBool("all")
			account, err := cmdBuilder.nakamaCtx.Client.GetAccount(cmdBuilder.nakamaCtx.Ctx, &emptypb.Empty{})
			if err != nil {
				log.Error(err)
				return err
			}
			if all {
				ticketStateList, err := getTicketStateList(cmdBuilder, account)
				if err != nil {
					log.Error(err)
					return err
				}

				for _, ticketState := range ticketStateList {
					if err != nil {
						log.Error(err)
						return err
					}

					fmt.Fprintf(cmd.OutOrStdout(), PrintTicketState(ticketState))
				}
				return nil
			}

			var ticketState *TicketState
			if id != "" {
				ticketState, err = getTicketState(cmdBuilder, id, account)
			} else {
				ticketState, err = getLastUserTicketState(cmdBuilder, account)
			}
			if err != nil {
				log.Error(err)
				return err
			}

			log.Infof("%+v", ticketState)
			if ticketState != nil {
				fmt.Fprintf(cmd.OutOrStdout(), PrintTicketState(ticketState))
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), fmt.Sprintf("No tickets found for <@%v>", account.CustomId))
			}

			return nil
		},
	}
	cmd.Flags().StringP("id", "i", "", "usage")
	cmd.Flags().BoolP("all", "a", false, "-all=true to get all active matches")
	return cmd
}

/*
func getCmdTicketGet2(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket [id]",
		Aliases: cmdTicketAliases,
		Short:   "Short",
		Long:    `Long`,
		Args:    matchAll(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			payload, _ := json.Marshal(&pb.GetTicketRequest{
				TicketId: id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "OpenMatchFrontendTicketGet", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			var ticket pb.Ticket
			json.Unmarshal([]byte(result.Payload), &ticket)

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(ticket))
			return nil
		},
	}
	cmd.Flags().StringP("id", "i", "", "usage")
	return cmd
}

func getCmdTicketDelete(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket [id]",
		Aliases: cmdTicketAliases,
		Short:   "Delete game ticket by id",
		Long:    `Long`,
		Args:    matchAll(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			payload, _ := json.Marshal(&pb.DeleteTicketRequest{
				TicketId: id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "OpenMatchFrontendTicketDelete", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			var ticket pb.Ticket
			json.Unmarshal([]byte(result.Payload), &ticket)

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(ticket))
			return nil
		},
	}
	cmd.Flags().StringP("id", "i", "", "usage")
	return cmd
}
*/

func getCmdTicketWatchAssignments(cmdBuilder *commandsBuilder) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket [id]",
		Aliases: cmdTicketAliases,
		Short:   "Short",
		Long:    `Long`,
		Args:    matchAll(cobra.MinimumNArgs(1)),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("%+v\n", args)
			id, _ := cmd.Flags().GetString("id")
			if id == "" && len(args) > 0 {
				id = args[0]
			}
			payload, _ := json.Marshal(&pb.WatchAssignmentsRequest{
				TicketId: id,
			})
			log.Infof("%+v\n", string(payload))

			result, err := cmdBuilder.nakamaCtx.Client.RpcFunc(cmdBuilder.nakamaCtx.Ctx, &api.Rpc{Id: "OpenMatchFrontendTicketWatchAssignments", Payload: string(payload)})
			if err != nil {
				log.Error(err)
				return err
			}
			var watchAssignmentsResponse pb.WatchAssignmentsResponse
			json.Unmarshal([]byte(result.Payload), &watchAssignmentsResponse)

			fmt.Fprintf(cmd.OutOrStdout(), MarshalIndent(watchAssignmentsResponse))
			return nil
		},
	}
	cmd.Flags().StringP("id", "i", "", "usage")
	return cmd
}
