package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	statisticsrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"

	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
	"github.com/noellimx/hepmilserver/src/utils/bytes"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	TGBOT_TOKEN := os.Getenv("TGBOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(TGBOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Optional: shows all bot activity in logs

	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Start interacting with the bot",
		},
		{
			Command:     "help",
			Description: "Get help and usage information",
		},
		{
			Command:     "now",
			Description: "Get help and usage information",
		},
	}

	cfg := tgbotapi.NewSetMyCommands(commands...)
	_, err = bot.Request(cfg)
	if err != nil {
		log.Fatalf("failed to set bot commands: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			switch update.Message.Command() {
			case "":
				var username string
				if update.Message.Chat != nil {
					username = update.Message.Chat.UserName
				}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`Hello %s 👋 Please use commands to chat with me~ Start your message with "/" \n /hello `, username))
				bot.Send(msg)
			case "start", "help":
				button1 := tgbotapi.NewInlineKeyboardButtonData("Add Task", "opt_1")
				button2 := tgbotapi.NewInlineKeyboardButtonData("Option 2", "opt_2")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(button1, button2),
				)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "*Hello*, /start click [here](https://example.com).")
				msg.ParseMode = "Markdown"

				//msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, world! See menu for commands")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case "now":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Working on it.... %v", update.Message.Time()))
				bot.Send(msg)
				client := StatsClient{
					Host: "http://localhost:8080",
				}

				resp, err := client.Get("memes", "top", "day", 3)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Something went wrong.... %v %v", resp.Error, err))
					bot.Send(msg)
					break
				}
				bb := convert(resp.Data.Posts)

				//postCh := reddit_miner.SubRedditPosts("memes", reddit_miner.CreatedWithinPastDay, reddit_miner.OrderByAlgoTop, false)
				//var posts []reddit_miner.Post
				//for post := range postCh {
				//	posts = append(posts, post)
				//}
				//var bb = toCsv(posts)
				b, err := bytes.TwoDStringAsBytes(bb)
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Something went wrong executing the command now")
					bot.Send(msg)
					return
				}
				message := tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FileBytes{
					Name:  "output.csv",
					Bytes: b.Bytes(),
				})
				bot.Send(message)
			default:
			}
		}
	}
}

func convert(posts []Post) [][]string {
	return nil
}

func toCsv(posts []reddit_miner.Post) [][]string {
	header := []string{
		"title",
		"perma_link_path",
		"data_ks_id",
		"subreddit_id",
		"subreddit_prefix_name",
		//"score",
		//"comment_count",
		"author_id",
		"author_name",
		//"created_timestamp",
	}
	rows := [][]string{header}
	for _, p := range posts {
		rows = append(rows, []string{
			p.Title,
			p.PermaLinkPath,
			p.DataKsId,
			p.SubredditId,
			p.SubredditPrefixedName,
			//p.Score,
			//p.CommentCount,
			p.AuthorId,
			p.AuthorName,
			//p.CreatedTimestamp,
		})
	}
	return rows
}

type StatsClient struct {
	Host string
}

type Response[T any] struct {
	Data  T       `json:"data"`
	Error *string `json:"error"`
}

type Post struct {
	Title                         string                           `json:"title"`
	PermaLinkPath                 string                           `json:"perma_link_path"`
	DataKsId                      string                           `json:"data_ks_id"`
	Score                         int32                            `json:"score"`
	SubredditId                   string                           `json:"subreddit_id"`
	CommentCount                  int32                            `json:"comment_count"`
	SubredditName                 string                           `json:"subreddit_name"`
	PolledTime                    time.Time                        `json:"polled_time"`
	AuthorId                      string                           `json:"author_id"`
	AuthorName                    string                           `json:"author_name"`
	PolledTimeRoundedMinute       time.Time                        `json:"polled_time_rounded_min"`
	Rank                          int32                            `json:"rank"`
	RankOrderType                 statisticsrepo.OrderByAlgo       `json:"rank_order_type"`
	RankOrderForCreatedWithinPast statisticsrepo.CreatedWithinPast `json:"rank_order_created_within_past"`
	Id                            int64                            `json:"id"`
}

type GetStatisticsResponseBodyData struct {
	Posts []Post `json:"posts"`
}

type GetStatisticsResponseBody = Response[GetStatisticsResponseBodyData]

func (s *StatsClient) Get(subRedditName string, rankOrderAlgoType string, rankPast string, granularity int) (GetStatisticsResponseBody, error) {
	baseUrl := s.Host + "/statistics"
	params := url.Values{}
	params.Add("subreddit_name", subRedditName)
	params.Add("rank_order_type", rankOrderAlgoType)
	params.Add("rank_order_created_within_past", rankPast)
	params.Add("granularity", strconv.Itoa(granularity))

	// Final URL with encoded params
	fullURL := baseUrl + "?" + params.Encode()

	resp, err := http.Get(fullURL)
	if err != nil {
		// handle error
		fmt.Println("Error:", err)
		return GetStatisticsResponseBody{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return GetStatisticsResponseBody{}, err
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))

	var b GetStatisticsResponseBody
	err = json.Unmarshal(body, &b)
	return b, nil
}
