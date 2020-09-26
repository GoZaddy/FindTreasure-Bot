package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/bot-api/telegram"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/gozaddy/findtreasure/types"
)

var (
	client                *http.Client      = http.DefaultClient
	jobQueue              chan job          = make(chan job)
	wc                    *workerController = newWorkerController()
	workersCreatedChannel chan bool         = make(chan bool)
	telegramAPI           *telegram.API
	chatID                int64 = 1042382451
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("error loading .env file", err)
	}
	telegramAPI = telegram.New(os.Getenv("TELEGRAM_TOKEN"))
	/*err = telegramAPI.SetWebhook(context.Background(), telegram.NewWebhook(os.Getenv("WEBHOOK_URL")))
	if err != nil {
		log.Fatalln("error setting webhook", err)
	}*/
}

func main() {

	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello!"})
	})

	r.POST("/", func(c *gin.Context) {
		var req types.TelegramResponse

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(req.Message.Text)
		fmt.Println(req.Message.Chat.ID)

		if req.Message.IsCommand() {
			com, _ := req.Message.Command()
			fmt.Println(com)
			switch com {
			case "run":
				msg := telegram.NewMessage(chatID, "Starting work!")
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
				resp, err := start()
				if err != nil {
					log.Fatalln(err)

				}
				go func() { wc.Run() }()

				select {
				case isWorkersCreated := <-workersCreatedChannel:
					if isWorkersCreated {
						go func() {
							fmt.Println("workers have been created")
							for _, v := range resp.Paths {
								job := job{resp.Encryption, v}
								jobQueue <- job
							}
						}()

					}
				}

			case "stop":
				msg := telegram.NewMessage(chatID, "Stopping work!")
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
				//wc.Stop()
			default:
				msg := telegram.NewMessage(chatID, "Hello, please send a recognised commmand /run or /stop")
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
			}

		}
	})

	r.Run()

}

func start() (*types.NodeResponse, error) {
	fmt.Println("starting...")
	client := http.DefaultClient

	req, err := http.NewRequest("GET", os.Getenv("BASE_URL")+"start", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("FIND_TREASURE_TOKEN"))
	req.Header.Set("gomoney", "09062289933")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeResponse types.NodeResponse

	err = json.Unmarshal(respBody, &nodeResponse)
	if err != nil {
		return nil, err
	}

	return &nodeResponse, nil

}
