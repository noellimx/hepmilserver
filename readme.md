# Reddit Miner
Applications for mining and storing data from Reddit.

# Requirements

Refer to `go.mod`

# Docs
- [Slides](https://docs.google.com/presentation/d/1m9q9mTnbAbfRuMqWFxH7f2i_wa83noimrTO5_gDSg8I/edit?slide=id.g36c2cab8088_0_0#slide=id.g36c2cab8088_0_0)

# Deployment
## http server
Package: `cmd/server/http`\
`run server`: `DATABASE_URL=<connstring> OPTIONAL_LOAD_ENV_FILE=TRUE LISTENING_PORT=<port> go run ./cmd/server/http`

## tgbot server
Package: `cmd/server/tgbot`\
`run server`: `API_SERVER_ADDRESS=<token> TGBOT_TOKEN=<token> go run cmd/server/tgbot/main.go`

# Swagger Docs Generation
`swag init --parseDependency --dir ./src/controller/mux/statistics,./src/controller/mux/task,./src/controller/mux/ping`
