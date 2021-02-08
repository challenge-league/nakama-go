package context

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/api"

	"github.com/heroiclabs/nakama/v2/apigrpc"
	log "github.com/micro/go-micro/v2/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	address            = "nakama.dataleague.svc.cluster.local:7349" // gRPC
	socketServerKey    = "defaultkey"
	NakamaSystemUserID = "00000000-0000-0000-0000-000000000000"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// Context
// can be used to retrieve context-specific args and
// parsed command-line options.
type Context struct {
	Conn       *grpc.ClientConn
	Client     apigrpc.NakamaClient
	Session    *api.Session
	Ctx        context.Context
	DiscordMsg *discordgo.Message
}

func GenerateString() string {
	return uuid.Must(uuid.NewV4()).String()
}

func NewBasicContext() (*Context, error) {
	ctx := context.Background()
	outgoingCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(socketServerKey+":")),
	}))
	conn, err := grpc.DialContext(outgoingCtx, address, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client := apigrpc.NewNakamaClient(conn)
	return &Context{Conn: conn, Client: client, Session: nil, Ctx: outgoingCtx}, nil
}

func NewCustomSession(authenticateCustomRequest *api.AuthenticateCustomRequest) (*Context, error) {
	nakamaCtx, err := NewBasicContext()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	session, err := nakamaCtx.Client.AuthenticateCustom(nakamaCtx.Ctx, authenticateCustomRequest)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	nakamaCtx.Session = session
	return nakamaCtx, nil
}

func NewEmailSession(authenticateEmailRequest *api.AuthenticateEmailRequest) (*Context, error) {
	nakamaCtx, err := NewBasicContext()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	session, err := nakamaCtx.Client.AuthenticateEmail(nakamaCtx.Ctx, authenticateEmailRequest)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	nakamaCtx.Session = session
	return nakamaCtx, nil
}

func NewEmailAuthenticatedSession(email string, password string) (*Context, error) {
	return NewEmailSession(&api.AuthenticateEmailRequest{
		Account: &api.AccountEmail{
			Email:    email,
			Password: password,
		},
		Username: email,
	})
}

func NewCustomIdSession(customId string) (*Context, error) {
	return NewCustomSession(
		&api.AuthenticateCustomRequest{
			Account: &api.AccountCustom{
				Id: customId,
			},
			Username: customId,
		})
}

func FixUsername(username string) string {
	return strings.ReplaceAll(username, " ", "_")
}

func GetAccountCustomVarsFromDiscordMessage(discordMsg *discordgo.Message) map[string]string {
	return map[string]string{
		"ChannelID":            discordMsg.ChannelID,
		"GuildID":              discordMsg.GuildID,
		"Author.ID":            discordMsg.Author.ID,
		"Author.Email":         discordMsg.Author.Email,
		"Author.Username":      FixUsername(discordMsg.Author.Username),
		"Author.Locale":        discordMsg.Author.Locale,
		"Author.Discriminator": discordMsg.Author.Discriminator,
		"Author.Verified":      strconv.FormatBool(discordMsg.Author.Verified),
		"Author.MFAEnbled":     strconv.FormatBool(discordMsg.Author.MFAEnabled),
		"Author.Bot":           strconv.FormatBool(discordMsg.Author.Bot),
		"Author.AvatarUrl":     discordMsg.Author.AvatarURL(""),
		"Content":              discordMsg.Content,
	}
}

func NewCustomDiscordSession(discordMsg *discordgo.Message) (*Context, error) {
	return NewCustomSession(&api.AuthenticateCustomRequest{
		Account: &api.AccountCustom{
			Id:   discordMsg.Author.ID,
			Vars: GetAccountCustomVarsFromDiscordMessage(discordMsg),
		},
		Username: FixUsername(discordMsg.Author.Username) + "#" + discordMsg.Author.Discriminator,
	})
}

func RestoreAuthenticatedAPIClient(nakamaCtx *Context, discordMessage *discordgo.Message) (*Context, error) {
	outgoingCtx := metadata.NewOutgoingContext(nakamaCtx.Ctx, metadata.New(map[string]string{
		"authorization": "Bearer " + nakamaCtx.Session.Token,
	}))
	conn, err := grpc.DialContext(outgoingCtx, address, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	client := apigrpc.NewNakamaClient(conn)
	_, err = client.Healthcheck(nakamaCtx.Ctx, &emptypb.Empty{})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Infof("Session restored")
	return &Context{conn, client, nakamaCtx.Session, outgoingCtx, discordMessage}, nil
}

func NewCustomAuthenticatedAPIClient(customId string) (*Context, error) {
	nakamaCtx, err := NewCustomIdSession(customId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	nakamaCtx, err = RestoreAuthenticatedAPIClient(nakamaCtx, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return nakamaCtx, nil

}

func NewCustomAuthenticatedAdminAPIClient() (*Context, error) {
	return NewCustomAuthenticatedAPIClient("administrator")
}

func NewCustomAuthenticatedDiscordAPIClient(discordMsg *discordgo.Message) (*Context, error) {
	nakamaCtx, err := NewCustomDiscordSession(discordMsg)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	nakamaCtx, err = RestoreAuthenticatedAPIClient(nakamaCtx, discordMsg)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return nakamaCtx, nil
}

func UserDataFromSession(session *api.Session) (map[string]interface{}, error) {
	parts := strings.Split(session.Token, ".")
	content, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, 0)
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}
	log.Infof("DECODE %v", data)
	return data, err
}
