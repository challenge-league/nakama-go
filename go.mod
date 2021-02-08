module github.com/challenge-league/nakama-go/v2

go 1.14

require (
	contrib.go.opencensus.io/exporter/prometheus v0.1.0 // indirect
	contrib.go.opencensus.io/resource v0.1.1 // indirect
	github.com/RoaringBitmap/roaring v0.4.21 // indirect
	github.com/blevesearch/bleve v0.8.2 // indirect
	github.com/blevesearch/snowballstem v0.0.0-20200325004757-48afb64082dd // indirect
	github.com/bwmarrin/discordgo v0.20.3
	github.com/challenge-league/nakama-go/commands v0.0.0-00010101000000-000000000000
	github.com/challenge-league/nakama-go/context v0.0.0-00010101000000-000000000000
	github.com/couchbase/vellum v0.0.0-20190829182332-ef2e028c01fd // indirect
	github.com/dgrijalva/jwt-go v3.2.1-0.20200107013213-dc14462fd587+incompatible // indirect
	github.com/envoyproxy/go-control-plane v0.9.4 // indirect
	github.com/etcd-io/bbolt v1.3.3 // indirect
	github.com/glycerine/go-unsnap-stream v0.0.0-20190901134440-81cf024a9e0a // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gobuffalo/syncx v0.0.0-20190224160051-33c29581e754 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.3 // indirect
	github.com/hako/durafmt v0.0.0-20200710122514-c0fb7b4da026 // indirect
	github.com/heroiclabs/nakama-common v1.5.1
	github.com/heroiclabs/nakama/v2/apigrpc v0.0.0-00010101000000-000000000000 // indirect
	github.com/jackc/pgx v3.5.0+incompatible // indirect
	github.com/m3db/prometheus_client_golang v0.8.1 // indirect
	github.com/m3db/prometheus_client_model v0.1.0 // indirect
	github.com/m3db/prometheus_common v0.1.0 // indirect
	github.com/m3db/prometheus_procfs v0.8.1 // indirect
	github.com/markbates/safe v1.0.1 // indirect
	github.com/micro/go-micro/v2 v2.7.0
	github.com/rubenv/sql-migrate v0.0.0-20190902133344-8926f37f0bc1 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb // indirect
	go.opencensus.io v0.22.3 // indirect
	go.uber.org/zap v1.14.1 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	google.golang.org/appengine v1.6.5
	google.golang.org/grpc v1.27.1
	gopkg.in/yaml.v2 v2.2.8 // indirect
	open-match.dev/open-match v1.0.0 // indirect
)

replace (
	github.com/challenge-league/nakama-go/commands => ./commands
	github.com/challenge-league/nakama-go/context => ./context
	github.com/heroiclabs/nakama/v2/apigrpc => ./apigrpc
)
