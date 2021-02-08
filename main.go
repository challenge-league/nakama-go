package main

import (
	"fmt"
	"log"
	"os"

	nakamaCommands "github.com/challenge-league/nakama-go/commands"
	nakamaContext "github.com/challenge-league/nakama-go/context"
)

func main() {
	nakamaCtx, err := nakamaContext.NewCustomAuthenticatedAdminAPIClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer nakamaCtx.Conn.Close()
	log.Printf("NewSession %v", nakamaCtx)

	cmdBuilder := nakamaCommands.NewCommandsBuilderSingleton()
	cmdBuilder.SetContext(nakamaCtx)
	nakamaCommands.Execute(cmdBuilder)
}
