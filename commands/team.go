/*
Copyright © 2020 Dmitry Kozov dmitry.f.kozlov@gmail.com

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

type Team struct {
	ID              int
	Name            string
	TeamUsers       []*TeamUser
	DiscordChannels []*DiscordChannel
}

type TeamUser struct {
	User     *User
	TicketID string
	Reward   float64
	Captain  bool
}

func PrintTeams(teams []*Team) string {
	msg := "> Teams:\n"
	for _, v := range teams {
		msg = msg + PrintTeam(v)
	}
	return msg
}

func PrintTeam(team *Team) string {
	msg := ExecuteTemplate(
		ExecuteTemplate("> **{{.ID}}** {{if .Name}}**{{.Name}}**{{end}}: ", team)+
			`{{range $index, $element := .TeamUsers}} <@{{.User.Nakama.CustomID}}> {{.User.Nakama.Username}} {{if .User.Nakama.Wallet }}({{.User.Nakama.Wallet | getCoinsFromWallet }}){{end}}{{ if .Reward | isFloatPositive }}**+{{.Reward}}**{{end}}{{ if .Reward | isFloatNegative }}**-{{.Reward}}**{{end}}{{end}}`+"\n",
		team)
	return msg
}

func GetUsersFromTeam(team *Team) []*User {
	var users []*User
	for _, v := range team.TeamUsers {
		users = append(users, v.User)
	}
	return users
}

func GetTeamUsersFromTeams(teams []*Team) []*TeamUser {
	var teamUsers []*TeamUser
	for _, team := range teams {
		for _, teamUser := range team.TeamUsers {
			teamUsers = append(teamUsers, teamUser)
		}
	}
	return teamUsers

}

func GetUsersFromTeamUsers(teamUsers []*TeamUser) []*User {
	var users []*User
	for _, v := range teamUsers {
		users = append(users, v.User)
	}
	return users
}

func GetTeamNumberWithFunctionOutputMap(teams []*Team, f func(*TeamUser) string) map[int][]string {
	teamMap := make(map[int][]string)
	for teamNumber, team := range teams {
		var slice []string
		for _, v := range team.TeamUsers {
			slice = append(slice, f(v))
		}
		teamMap[teamNumber] = slice
	}
	return teamMap
}

func GetDiscordIDsByTeamNumber(teamNumber int, matchState *MatchState) []string {
	teamMap := GetTeamNumberWithFunctionOutputMap(
		matchState.Teams,
		func(teamUser *TeamUser) string {
			return teamUser.User.Nakama.CustomID
		},
	)
	if discordIDs, ok := teamMap[teamNumber]; ok {
		return discordIDs
	}
	return []string{}
}

func GetMaxUserCountPerTeam(matchState *MatchState) int {
	max := -1
	for _, team := range matchState.Teams {
		if len(team.TeamUsers) > max {
			max = len(team.TeamUsers)
		}
	}
	return max
}

func GetUserIDsByTeamNumber(teamNumber int, matchState *MatchState) []string {
	teamMap := GetTeamNumberWithFunctionOutputMap(
		matchState.Teams,
		func(teamUser *TeamUser) string {
			return teamUser.User.Nakama.ID
		},
	)
	if userIDs, ok := teamMap[teamNumber]; ok {
		return userIDs
	}
	return []string{}
}

func GetUserAndTeamNumberByUserID(userID string, matchState *MatchState) (*TeamUser, int) {
	for teamNumber, team := range matchState.Teams {
		for _, v := range team.TeamUsers {
			if v.User.Nakama.ID == userID {
				return v, teamNumber
			}
		}
	}
	return nil, -1
}

func GetUsersReady(matchState *MatchState) map[int][]*UserReady {
	teamUsersReadyMap := make(map[int][]*UserReady)
	for k, team := range matchState.Teams {
		for _, teamUser := range team.TeamUsers {
			if IsStringInSlice(teamUser.User.Nakama.ID, matchState.ReadyUserIDs) {
				teamUsersReadyMap[k] = append(teamUsersReadyMap[k], &UserReady{
					Ready:     true,
					UserID:    teamUser.User.Nakama.ID,
					DiscordID: teamUser.User.Nakama.CustomID,
				})
			} else {
				teamUsersReadyMap[k] = append(teamUsersReadyMap[k], &UserReady{
					Ready:     false,
					UserID:    teamUser.User.Nakama.ID,
					DiscordID: teamUser.User.Nakama.CustomID,
				})
			}
		}
	}

	return teamUsersReadyMap
}

func GetTeamNumberFromUserAndMatch(userID string, matchState *MatchState) int {
	teamMap := GetTeamNumberWithFunctionOutputMap(
		matchState.Teams,
		func(teamUser *TeamUser) string {
			return teamUser.User.Nakama.ID
		},
	)
	for teamNumber, userList := range teamMap {
		if IsStringInSlice(userID, userList) {
			return teamNumber
		}
	}
	return -1
}

func IsUserIDInMatch(userID string, matchState *MatchState) bool {
	return -1 != GetTeamNumberFromUserAndMatch(userID, matchState)
}
