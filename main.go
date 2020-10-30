package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bot-api/telegram"
	"github.com/gin-gonic/gin"
	"github.com/gozaddy/findtreasure/models"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gozaddy/findtreasure/types"
)

var (

	telegramAPI  *telegram.API
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("error loading .env file ", err)
	}
	telegramAPI = telegram.New(os.Getenv("TELEGRAM_TOKEN"))

}

func main() {
	var chatID int64
	ci, err := strconv.Atoi(os.Getenv("CHAT_ID"))
	if err != nil{
		chatID = 1042382451
	} else {
		chatID = int64(ci)
	}

	maxWorkers, err := strconv.Atoi(os.Getenv("MAX_WORKERS"))
	if err != nil{
		maxWorkers = 50
	}


	wc := models.NewWorkerController(maxWorkers)

	//spawn goroutine for sending messages through telegram
	go func() {
		for{
			select {
			case message := <-wc.MessageChan:
				msg := telegram.NewMessage(chatID, message)
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}

			}
		}
	}()




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
				if wc.State == models.WorkerControllerRunning{
					msg := telegram.NewMessage(chatID, "WorkerController already running!")
					_, err := telegramAPI.SendMessage(context.Background(), msg)
					if err != nil {
						log.Println(err)
					}
					return
				}

				startWork(wc, chatID)



			case "stop":
				msg := telegram.NewMessage(chatID, "Stopping work!")
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
				wc.Stop()

			case "state":
				msg := telegram.NewMessage(chatID, "State: "+wc.GetState())
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
			default:
				msg := telegram.NewMessage(chatID, "Hello, please send a recognised command /run or /stop or /state")
				_, err := telegramAPI.SendMessage(context.Background(), msg)
				if err != nil {
					log.Println(err)
				}
			}

		}
	})

	r.Run()

}

func startWork(wc *models.WorkerController, chatID int64){
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
	case isWorkersCreated := <- wc.WorkersCreatedChannel:
		if isWorkersCreated {
			go func() {
				fmt.Println("workers have been created")
				for _, v := range resp.Paths {
					job := models.Job{resp.Encryption, v}
					wc.JobQueue <- job
				}
			}()

		}

	}
}

func start() (*types.NodeResponse, error) {
	fmt.Println("starting...")
	client := http.DefaultClient

	body1 := gin.H{
		"email": "faruqyusuff437@gmail.com" ,
	}

	bs, err := json.Marshal(&body1)
	if err != nil{
		return nil, err
	}
	//get token for accessing endpoints
	req1, err := http.NewRequest("POST", "https://findtreasure.app/api/v1/contestants/refresh", bytes.NewBuffer(bs))
	req1.Header.Set("Content-Type", "application/json")
	if err != nil{
		return nil, err
	}
	resp, err := client.Do(req1)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	if err != nil {
		return nil, err
	}

	var tokenInfo struct{
		Token string `json:"token"`
	}

	err = json.Unmarshal(respBody, &tokenInfo)
	if err != nil {
		return nil, err
	}

	fmt.Println("Token: ", tokenInfo.Token)
	_ = os.Setenv("FIND_TREASURE_TOKEN", tokenInfo.Token)


	//hit start endpoint
	req, err := http.NewRequest("GET", os.Getenv("BASE_URL")+"start", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenInfo.Token)
	req.Header.Set("accountId", "faru")

	resp, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err = ioutil.ReadAll(resp.Body)
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
