package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
	"github.com/noellimx/hepmilserver/src/utils/bytes"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
				fallthrough
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, world! See menu for commands")
				bot.Send(msg)
			case "now":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Working on it.... %v", update.Message.Time()))
				bot.Send(msg)

				postCh := reddit_miner.SubRedditPosts("memes", reddit_miner.CreatedWithinPastDay, reddit_miner.OrderByColumnTop, false)
				var posts []reddit_miner.Post
				for post := range postCh {
					posts = append(posts, post)
				}
				toCsv(posts)
				b, err := bytes.TwoDStringAsBytes(toCsv(posts))
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
			}
		}
	}
}

func toCsv(posts []reddit_miner.Post) [][]string {
	header := []string{
		"title",
		"perma_link_path",
		"data_ks_id",
		"subreddit_id",
		"subreddit_prefix_name",
		"score",
		"comment_count",
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
			p.Score,
			p.CommentCount,
			p.AuthorId,
			p.AuthorName,
			//p.CreatedTimestamp,
		})
	}
	return rows
}
