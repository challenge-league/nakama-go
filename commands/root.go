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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"os"
	"sync"

	nakama "github.com/challenge-league/nakama-go/context"
	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cmdBuilder *commandsBuilder
	once       sync.Once
	cfgFile    string
)

type commandsBuilder struct {
	nakamaCtx *nakama.Context
	rootCmd   *cobra.Command
}

func NewCommandsBuilderSingleton() *commandsBuilder {
	once.Do(func() {
		cmdBuilder = &commandsBuilder{
			rootCmd:   NewRootCmd(),
			nakamaCtx: nil,
		}
	})
	return cmdBuilder
}

func NewCommandsBuilder() *commandsBuilder {
	return &commandsBuilder{
		rootCmd:   NewRootCmd(),
		nakamaCtx: nil,
	}
}

func (b *commandsBuilder) SetContext(nakamaCtx *nakama.Context) *commandsBuilder {
	b.nakamaCtx = nakamaCtx
	return b
}

func (b *commandsBuilder) GetContext() *nakama.Context {
	return b.nakamaCtx
}

func (b *commandsBuilder) SetRootCmd(rootCmd *cobra.Command) *commandsBuilder {
	b.rootCmd = rootCmd
	return b
}

func (b *commandsBuilder) GetRootCmd() *cobra.Command {
	return b.rootCmd
}

func checkPermission(b *commandsBuilder) bool {
	// Check if user = dkozlov
	return b.nakamaCtx.DiscordMsg != nil && b.nakamaCtx.DiscordMsg.Author.ID == "554195751274807297"
}

func (b *commandsBuilder) SetCommandsAndFlags() *commandsBuilder {
	b.rootCmd.ResetCommands()
	b.rootCmd.ResetFlags()

	/*
		if checkPermission(b) {
			cmdBan := getCmdBan(b)
			cmdBan.AddCommand(getCmdGroupUsersBan(b))
			b.rootCmd.AddCommand(cmdBan)

			cmdKick := getCmdKick(b)
			cmdKick.AddCommand(getCmdGroupUsersKick(b))
			b.rootCmd.AddCommand(cmdKick)
		}

			cmdCreate := getCmdCreate(b)
			//cmdCreate.AddCommand(getCmdGroupCreate(b))
			//cmdCreate.AddCommand(getCmdGroupUsersAdd(b))
			cmdCreate.AddCommand(getCmdTicketCreate(b))

			if checkPermission(b) {
				cmdCreate.AddCommand(getCmdTournamentCreate(b))
				cmdCreate.AddCommand(getCmdLeaderboardCreate(b))
				cmdCreate.AddCommand(getCmdLeaderboardRecordAdd(b))
				cmdCreate.AddCommand(getCmdMatchCreate(b))
				cmdCreate.AddCommand(getCmdTournamentRecordCreate(b))
			}
			b.rootCmd.AddCommand(cmdCreate)

		cmdDelete := getCmdDelete(b)
		//cmdDelete.AddCommand(getCmdGroupDelete(b))
		cmdDelete.AddCommand(getCmdTicketDelete(b))
		if checkPermission(b) {
			cmdDelete.AddCommand(getCmdTournamentDelete(b))
			cmdDelete.AddCommand(getCmdLeaderboardDelete(b))
			cmdDelete.AddCommand(getCmdLeaderboardRecordDelete(b))
		}
		b.rootCmd.AddCommand(cmdDelete)
	*/

	cmdGet := getCmdGet(b)
	//cmdGet.AddCommand(getCmdTournamentGet(b))
	//cmdGet.AddCommand(getCmdGroupGet(b))
	//cmdGet.AddCommand(getCmdGroupUsersGet(b))
	cmdGet.AddCommand(getCmdMatchGet(b))
	cmdGet.AddCommand(getCmdLeaderboardRecordsGet(b))
	cmdGet.AddCommand(getCmdTopLeaderboardRecordsGet(b))
	//cmdGet.AddCommand(getCmdLeaderboardRecordAroundOwnerGet(b))
	cmdGet.AddCommand(getCmdTicketGet(b))
	//cmdGet.AddCommand(getCmdTournamentRecordGet(b))
	//cmdGet.AddCommand(getCmdTournamentRecordAroundOwnerGet(b))
	cmdGet.AddCommand(getCmdUserGet(b))
	//cmdGet.AddCommand(getCmdUserGroupsGet(b))
	b.rootCmd.AddCommand(cmdGet)

	//cmdGo := getCmdGo(b)
	//cmdGo.AddCommand(getCmdTicketCreate(b))
	//b.rootCmd.AddCommand(cmdGo)

	cmdGetLeaderboard := getCmdLeaderboardRecordsGet(b)
	b.rootCmd.AddCommand(cmdGetLeaderboard)

	cmdGetTopLeaderboard := getCmdTopLeaderboardRecordsGet(b)
	b.rootCmd.AddCommand(cmdGetTopLeaderboard)

	cmdTicketCreate := getCmdTicketCreate(b)
	b.rootCmd.AddCommand(cmdTicketCreate)

	cmdSubmit := getCmdSubmit(b)
	b.rootCmd.AddCommand(cmdSubmit)

	cmdJoinMatchPool := getCmdJoinMatchPool(b)
	b.rootCmd.AddCommand(cmdJoinMatchPool)

	cmdPickUserFromMatchPool := getCmdPickUserFromMatchPool(b)
	b.rootCmd.AddCommand(cmdPickUserFromMatchPool)

	cmdAddUserToMatchPool := getCmdAddUserToMatchPool(b)
	b.rootCmd.AddCommand(cmdAddUserToMatchPool)

	/*

		cmdLeave := getCmdLeave(b)
		//cmdLeave.AddCommand(getCmdGroupLeave(b))
		cmdLeave.AddCommand(getCmdTicketDelete(b))
		b.rootCmd.AddCommand(cmdLeave)

		cmdPromote := getCmdPromote(b)
		cmdPromote.AddCommand(getCmdGroupUsersPromote(b))
		b.rootCmd.AddCommand(cmdPromote)

		cmdUpdate := getCmdUpdate(b)
		cmdUpdate.AddCommand(getCmdGroupUpdate(b))
		b.rootCmd.AddCommand(cmdUpdate)
	*/
	cmdLogin := getCmdLogin(b)
	b.rootCmd.AddCommand(cmdLogin)

	cmdReady := getCmdReady(b)
	b.rootCmd.AddCommand(cmdReady)

	cmdCancel := getCmdCancel(b)
	b.rootCmd.AddCommand(cmdCancel)

	cmdResultWin := getCmdResultWinCreate(b)
	b.rootCmd.AddCommand(cmdResultWin)

	cmdResultDraw := getCmdResultDrawCreate(b)
	b.rootCmd.AddCommand(cmdResultDraw)

	cmdResultLoss := getCmdResultLossCreate(b)
	b.rootCmd.AddCommand(cmdResultLoss)

	cmdChallengeCreate := getCmdCaptainsDraftCreate(b)
	b.rootCmd.AddCommand(cmdChallengeCreate)

	/*
		cmdWatch := getCmdWatch(b)
		cmdWatch.AddCommand(getCmdTicketWatchAssignments(b))
		b.rootCmd.AddCommand(cmdWatch)
	*/

	return b
}

func matchAll(checks ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, check := range checks {
			if err := check(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

func ExecuteCommandWithContext(ctx context.Context, root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)
	return buf.String(), err
}

func ExecuteCommandC(b *commandsBuilder, args ...string) (output string, err error) {
	bufferString := bytes.NewBufferString("")
	b.SetCommandsAndFlags()
	rootCmd := b.GetRootCmd()
	rootCmd.SetOut(bufferString)
	rootCmd.SetErr(bufferString)
	rootCmd.SetArgs(args)

	rootCmd.ExecuteC()
	out, err := ioutil.ReadAll(bufferString)
	fmt.Printf("%+v %+v", err, out)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	return string(out), err
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(b *commandsBuilder) {
	b.SetCommandsAndFlags()
	if err := b.GetRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	return &cobra.Command{
		Use:   "dl",
		Short: "Data League CLI",
		Long:  ``,
		// Uncomment the following line if your bare application
		// has an action associated with it:
		//      Run: func(cmd *cobra.Command, args []string) { },
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//cmdBuilder.rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dataleague.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".commands" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".dataleague")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())

	}
}
