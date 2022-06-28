module github.com/nakji-network/connector/gocryptotrader

go 1.13

replace github.com/thrasher-corp/gocryptotrader => ./

require (
	github.com/d5/tengo/v2 v2.8.0
	github.com/gofrs/uuid v4.0.0+incompatible
	github.com/google/go-querystring v1.1.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.5.0
	github.com/kat-co/vala v0.0.0-20170210184112-42e1d8b61f12
	github.com/lib/pq v1.10.2
	github.com/mattn/go-sqlite3 v1.14.8
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/shopspring/decimal v1.2.0
	github.com/spf13/viper v1.8.1
	github.com/thrasher-corp/gct-ta v0.0.0-20200623072738-f2b55b7f9f41
	github.com/thrasher-corp/gocryptotrader v0.0.0-00010101000000-000000000000
	github.com/thrasher-corp/goose v2.7.0-rc4.0.20191002032028-0f2c2a27abdb+incompatible
	github.com/thrasher-corp/sqlboiler v1.0.1-0.20191001234224-71e17f37a85e
	github.com/urfave/cli/v2 v2.3.0
	github.com/volatiletech/null v8.0.0+incompatible
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced
	google.golang.org/grpc v1.39.0
	google.golang.org/protobuf v1.27.1
)
