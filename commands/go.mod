module commands

go 1.14

require (
	github.com/bwmarrin/discordgo v0.20.3
	github.com/challenge-league/nakama-go/context v0.0.0-00010101000000-000000000000
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.4.1
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026
	github.com/heroiclabs/nakama-common v1.5.1
	github.com/micro/go-micro/v2 v2.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	google.golang.org/protobuf v1.22.0
	open-match.dev/open-match v1.0.0
)

replace (
	github.com/challenge-league/nakama-go/context => ../context
)
