package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	statisticsrepo "github.com/noellimx/hepmilserver/src/infrastructure/repositories/statistics"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/noellimx/hepmilserver/src/infrastructure/reddit_miner"
)

func main() {
	godotenv.Load()
	TGBOT_TOKEN := os.Getenv("TGBOT_TOKEN")
	SERVER_ADDRESS := os.Getenv("API_SERVER_ADDRESS")

	if SERVER_ADDRESS == "" {
		log.Fatal("API_SERVER_ADDRESS environment variable not set")
	}

	bot, err := tgbotapi.NewBotAPI(TGBOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // Optional: shows all bot activity in logs

	log.Printf("Authorized on account %s", bot.Self.UserName)

	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "your journey begins here",
		},
		{
			Command:     "report",
			Description: "download historical dataset over time",
		},
		{
			Command:     "now",
			Description: "view current trending posts in subreddit",
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

flush_old: // prevent thundering herd
	for {
		select {
		case <-updates:
		case <-time.NewTicker(1 * time.Second).C:
			break flush_old
		}
	}
	for update := range updates {
		processUpdate(update, bot, SERVER_ADDRESS)
	}
}

func processUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI, serverAddress string) {
	website := "https://liger-social-crane.ngrok-free.app/"

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()
	if update.Message != nil {
		switch command := update.Message.Command(); command {
		case "":
			userName := getUsername(update)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`Hello %s 👋 Please use commands to chat with me~ Start your message with "/" \n`, userName))
			bot.Send(msg)
		case "start", "help":
			userName := getUsername(update)

			/*
						Command:     "report",
					Description: "download historical dataset over time",
				},
				{
					Command:     "now",
					Description: "view current trending posts in subreddit",
			*/
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`
Hello %s 👋, Welcome to Reddit Miner.
				
Links:
website: %s
design / live demo: <a href="https://docs.google.com/presentation/d/1v3m7omQCMDQCkXm_0DNJfBWddqaGq0aNfmLxQwVp4F8/edit?slide=id.g35e6300cbb9_0_47#slide=id.g35e6300cbb9_0_47"> slides </a>
API Docs: <a href="https://liger-social-crane.ngrok-free.app/api/swagger/index.html"> swagger </a>
Github: <a href="https://github.com/noellimx/mk-fe.git"> backend </a> |  <a href="https://github.com/noellimx/mk-fe.git"> frontend </a> 

/report 	download historical dataset over time
/now		view current trending posts in subreddit
				`, userName, website))
			msg.ParseMode = "HTML"
			bot.Send(msg)

		case "report":
			client := TaskClient{
				Host: serverAddress,
			}
			resp, err := client.GetList()
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Something went wrong.... %v %v", resp.Error, err))
				bot.Send(msg)
				break
			}

			if len(resp.Data.Tasks) == 0 {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("No tasks found. Please visit %s to create a task.", website))
				bot.Send(msg)
				break
			}

			var buttons []tgbotapi.InlineKeyboardButton
			for _, task := range resp.Data.Tasks {
				buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(task.SubRedditName, "report"+"_"+task.SubRedditName))
			}
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(buttons...),
			)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Which subreddit would you like to obtain the dataset?")
			msg.ReplyMarkup = keyboard
			bot.Send(msg)

		default:
			switch {
			case strings.HasPrefix(command, "now"):
				client := TaskClient{
					Host: serverAddress,
				}
				resp, err := client.GetList()
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Something went wrong.... %v %v", resp.Error, err))
					bot.Send(msg)
					break
				}

				var buttons []tgbotapi.InlineKeyboardButton
				for _, task := range resp.Data.Tasks {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(task.SubRedditName, "now"+"_"+task.SubRedditName))
				}
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Which subreddit would you like to query?")
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
				break
			}
		}
	} else if update.CallbackQuery != nil {
		callback := update.CallbackQuery
		commandargs := strings.Split(callback.Data, "_")
		if len(commandargs) == 0 {
			return
		}
		//args := strings.Split(command, "_")

		// len 2 -> [now, subreddit] -> request order
		// len 3 -> [now, subreddit, order] -> request past
		// len 4 -> [now, subreddit, order, past] -> get statistics
		cmd := commandargs[0]
		chatId := update.CallbackQuery.Message.Chat.ID
		switch cmd {
		case "now":
			switch len(commandargs) {

			case 2:
				var buttons []tgbotapi.InlineKeyboardButton
				for _, order := range []string{"top"} {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(order, callback.Data+"_"+order))
				}
				msg := tgbotapi.NewMessage(chatId, "Select Sort By:")

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case 3:
				var buttons []tgbotapi.InlineKeyboardButton
				for _, past := range []string{"day"} {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(past, callback.Data+"_"+past))
				}
				msg := tgbotapi.NewMessage(chatId, "Select Post Created Time")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case 4:
				sName := commandargs[1]
				order := commandargs[2]
				past := commandargs[3]

				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`SubReddit: <a href="https://reddit.com/%s">r/%s</a> Order: %s T: %s
Working on it...`, sName, sName, order, past))
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true

				bot.Send(msg)

				go func() {
					ch := reddit_miner.SubRedditPosts(sName, reddit_miner.CreatedWithinPast(past), reddit_miner.OrderByAlgo(order), false)
					var ps []reddit_miner.Post
					for p := range ch {
						ps = append(ps, p)
					}

					log.Printf("len(ps): %d", len(ps))
					if len(ps) == 0 {
						msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`SubReddit: <a href="https://reddit.com/%s">r/%s</a> Order: %s T: %s
Result: No Data...`, sName, sName, order, past))
						msg.DisableWebPagePreview = true
						msg.ParseMode = "HTML"

						bot.Send(msg)
						return
					}

					slices.SortFunc(ps, func(a, b reddit_miner.Post) int {
						if a.Rank < b.Rank {
							return -1
						} else if a.Rank > b.Rank {
							return 1
						}
						return 0
					})

					var ss = strings.Builder{}
					if len(ps) > 20 {
						ps = ps[:20]
					}

					ss.WriteString(fmt.Sprintf(`SubReddit: <a href="https://reddit.com/%s">r/%s</a> Order: %s T: %s`, sName, sName, order, past))
					for _, p := range ps {
						logging := fmt.Sprintf(`
Rank %02d: <a href="https://reddit.com%s">%s</a> `, p.Rank, p.PermaLinkPath, p.Title)
						ss.WriteString(logging)
					}

					log.Println(ss.String())
					msg = tgbotapi.NewMessage(chatId, ss.String())
					//msg.ParseMode = "HTML"
					msg.ParseMode = "HTML"
					bot.Send(msg)
				}()
			}
		case "report":
			switch len(commandargs) {
			case 2:
				var buttons []tgbotapi.InlineKeyboardButton
				for _, order := range []string{"top"} {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(order, callback.Data+"_"+order))
				}
				msg := tgbotapi.NewMessage(chatId, "Select Sort By:")

				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case 3:
				var buttons []tgbotapi.InlineKeyboardButton
				for _, past := range []string{"day"} {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(past, callback.Data+"_"+past))
				}
				msg := tgbotapi.NewMessage(chatId, "Select Post Created Time:")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)

			case 4:
				var buttons []tgbotapi.InlineKeyboardButton
				for _, past := range []string{"Past 24 Hours"} {
					buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(past, callback.Data+"_"+past))
				}
				msg := tgbotapi.NewMessage(chatId, "Select Time Range:")
				keyboard := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(buttons...),
				)
				msg.ReplyMarkup = keyboard
				bot.Send(msg)
			case 5:
				sName := commandargs[1]
				order := commandargs[2]
				past := commandargs[3]
				trange := commandargs[4]

				var f, tt time.Time
				switch trange {
				case "Past 24 Hours":
					tt = time.Now()
					f = tt.Add(-24 * time.Hour)
				}

				msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`SubReddit: <a href="https://reddit.com/%s">r/%s</a> Order: %s T: %s Time Range: Past 24 Hours
Working on report...`, sName, sName, order, past))
				msg.ParseMode = "HTML"
				msg.DisableWebPagePreview = true
				bot.Send(msg)

				go func() {
					s := StatsClient{
						Host: serverAddress,
					}

					respBody, fileName, err := s.GetCSV(sName, order, past, 3, f, tt)
					if err != nil {
						msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Something went wrong.... %v", err))
						bot.Send(msg)
						return
					}

					if fileName == "" {
						fileName = "output.csv"
					}
					message := tgbotapi.NewDocument(chatId, tgbotapi.FileBytes{
						Name:  fileName,
						Bytes: respBody,
					})
					bot.Send(message)

					msg := tgbotapi.NewMessage(chatId, fmt.Sprintf("Report Download Success. Visit website for more options."))
					bot.Send(msg)
				}()
			}
		}
	}
}

func formatTable(rows [][]string) string {
	var result string
	result += "Post Name         | Rank\n"
	result += "------------------|-----\n"
	for _, row := range rows {
		result += fmt.Sprintf("%-18s| %s\n", row[0], row[1])
	}
	return result
}

func getUsername(update tgbotapi.Update) string {
	var username string
	if update.Message.Chat != nil {
		username = update.Message.Chat.UserName
	}

	return username
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

func (s *StatsClient) GetJSON(subRedditName string, rankOrderAlgoType string, rankPast string, granularity int) (GetStatisticsResponseBody, error) {
	baseUrl := s.Host + "/statistics"
	params := url.Values{}
	params.Add("subreddit_name", subRedditName)
	params.Add("rank_order_type", rankOrderAlgoType)
	params.Add("rank_order_created_within_past", rankPast)
	params.Add("granularity", strconv.Itoa(granularity))

	// Final URL with encoded params
	fullURL := baseUrl + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return GetStatisticsResponseBody{}, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
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

func (s *StatsClient) GetCSV(subRedditName string, rankOrderAlgoType string, rankPast string, granularity int, _fromTime, _toTime time.Time) ([]byte, string, error) {
	baseUrl := s.Host + "/statistics"
	fromTime := _fromTime.Format("2006-01-02T15:04:05.000Z")
	toTime := _toTime.Format("2006-01-02T15:04:05.000Z")

	params := url.Values{}
	params.Add("subreddit_name", subRedditName)
	params.Add("rank_order_type", rankOrderAlgoType)
	params.Add("rank_order_created_within_past", rankPast)
	params.Add("granularity", strconv.Itoa(granularity))
	params.Add("to_time", toTime)
	params.Add("from_time", fromTime)

	// Final URL with encoded params
	fullURL := baseUrl + "?" + params.Encode()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return []byte{}, "", err
	}
	req.Header.Set("Accept", "text/csv")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		fmt.Println("Error:", err)
		return []byte{}, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []byte{}, "", errors.New(resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		var b GetStatisticsResponseBody
		err = json.Unmarshal(body, &b)
		fmt.Println("Read error:", err)
		em := "error"
		if b.Error != nil {
			em = *b.Error
		}
		return []byte{}, "", fmt.Errorf(em)
	}

	return body, extractFilename(resp.Header.Get("Content-Disposition")), nil

}

func extractFilename(contentDisposition string) string {
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		log.Println("Error:", err, contentDisposition)
		return ""
	}
	filename, ok := params["filename"]
	log.Printf("filename: %v, %v\n", filename, ok)
	if ok {
		return filename
	}

	// Handle filename*=UTF-8'' format
	filenameUTF8, ok := params["filename*"]
	if ok && strings.HasPrefix(filenameUTF8, "UTF-8''") {
		return strings.TrimPrefix(filenameUTF8, "UTF-8''")
	}
	return ""
}

type Task struct {
	Id                     int64  `json:"id"`
	SubRedditName          string `json:"subreddit_name"`
	MinItemCount           int64  `json:"min_item_count"`
	Interval               string `json:"interval"`
	OrderBy                string `json:"order_by"`
	PostsCreatedWithinPast string `json:"posts_created_within_past"`
}

type GetTaskResponseBodyData struct {
	Tasks []Task `json:"tasks"`
}
type GetTaskResponseBody = Response[GetTaskResponseBodyData]

type TaskClient struct {
	Host string
}

func (s *TaskClient) GetList() (GetTaskResponseBody, error) {
	baseUrl := s.Host + "/tasks"
	fullURL := baseUrl

	resp, err := http.Get(fullURL)
	if err != nil {
		// handle error
		fmt.Println("Error:", err)
		return GetTaskResponseBody{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Read error:", err)
		return GetTaskResponseBody{}, err
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Body:", string(body))

	var b GetTaskResponseBody
	err = json.Unmarshal(body, &b)
	return b, nil
}
