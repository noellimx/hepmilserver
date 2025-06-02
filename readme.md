# Design Docs

- [Slides](https://docs.google.com/presentation/d/1v3m7omQCMDQCkXm_0DNJfBWddqaGq0aNfmLxQwVp4F8/edit?slide=id.g35e6300cbb9_0_47#slide=id.g35e6300cbb9_0_47)

# Generate swagger docs
`swag init --parseDependency --dir ./src/controller/mux/statistics,./src/controller/mux/task,./src/controller/mux/ping`


## http server
Package: `cmd/server/http`

`run server`: `API_SERVER_ADDRESS=<token> TGBOT_TOKEN=<token> go run cmd/server/tgbot/main.go`



## http server
Package: `cmd/server/tgbot`
`run server`: `DATABASE_URL=<connstring> OPTIONAL_LOAD_ENV_FILE=TRUE LISTENING_PORT=<port> go run ./cmd/server/http`

