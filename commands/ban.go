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
	"github.com/spf13/cobra"
)

func getCmdBan(cmdBuilder *commandsBuilder) *cobra.Command {
	cmdBan := &cobra.Command{
		Use:     "ban",
		Aliases: []string{"b"},
		Short:   "short description",
		Long:    `long description`,
		/*RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("create called")
			return nil
		},
		*/
	}
	return cmdBan
}
