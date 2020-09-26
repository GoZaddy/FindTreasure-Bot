package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bot-api/telegram"

	"github.com/gozaddy/findtreasure/mycrypto"
	"github.com/gozaddy/findtreasure/types"
)

const (
	maxWorkers int = 40
)

var (
	//ErrTooManyRequests is returned when a worker controller is currently paused
	ErrTooManyRequests error = errors.New("The worker controller is currently paused")
)

type worker struct {
	workerController workerController
	jobChannel       chan job
	quit             chan bool
}

func newWorker(wc workerController) worker {
	return worker{
		wc,
		make(chan job),
		make(chan bool),
	}
}

func (w worker) Start() {
	go func() {
		for {
			//register current work into worker pool
			w.workerController.WorkerPool <- w.jobChannel

			select {
			case job := <-w.jobChannel:
				if *w.workerController.IsPaused == false {
					_, err := job.Run()
					if err != nil {
						if errors.Is(err, ErrTooManyRequests) {
							fmt.Println("heree")
							if *w.workerController.IsPaused == false {
								w.workerController.Pause()
							}
							wc.RetryCount++
							wc.RetryQueue <- job

						} else {
							log.Println(err)
						}
					}
					fmt.Println("ran job")

				} else {
					fmt.Println("hereeeee")
					wc.RetryCount++
					wc.RetryQueue <- job
				}
			case <-w.quit:
				return
			}
		}

	}()
}

func (w worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

//workerController is where the main stuff happens. We can pause and resume work from here and it also houses the worker poo.
type workerController struct {
	IsPaused   *bool
	pause      chan bool
	MaxWorkers int
	RetryQueue chan job
	RetryCount int
	WorkerPool chan chan job
	quit       chan bool
}

func newWorkerController() *workerController {
	f := false
	return &workerController{
		&f,
		make(chan bool),
		maxWorkers,
		make(chan job),
		0,
		make(chan chan job, maxWorkers),
		make(chan bool),
	}
}

//Run starts the worker controller
func (wc *workerController) Run() {
	for i := 0; i < wc.MaxWorkers; i++ {
		//create and start new workers
		fmt.Println("creating workers...")
		worker := newWorker(*wc)
		worker.Start()

	}
	workersCreatedChannel <- true

	go func() {
		for {
			select {
			case j := <-jobQueue:
				fmt.Println("received job")
				go func(j job) {
					if *wc.IsPaused == false {
						jobChannel := <-wc.WorkerPool

						jobChannel <- j
					} else {
						fmt.Println("hereee")
						wc.RetryCount++
						wc.RetryQueue <- j

					}

					fmt.Println("job has been sent to worker")
				}(j)

			case <-wc.quit:
				fmt.Println("quitting")
				return

			case <-wc.pause:
				if *wc.IsPaused == false {
					fmt.Println("Pausing for a minute...")
					*wc.IsPaused = true
					go func() {
						select {
						case <-time.After(1 * time.Minute):
							*wc.IsPaused = false
							fmt.Println("Resume work")
							msg := telegram.NewMessage(chatID, "Resuming work!, Retry Count: "+strconv.Itoa(wc.RetryCount))
							_, err := telegramAPI.SendMessage(context.Background(), msg)
							if err != nil {
								log.Println(err)
							}

							for j := range wc.RetryQueue {

								jobQueue <- j

								time.Sleep(500 * time.Millisecond)

							}
						}

					}()
				} else {
					msg := telegram.NewMessage(chatID, "Already paused")
					_, err := telegramAPI.SendMessage(context.Background(), msg)
					if err != nil {
						log.Println(err)
					}
				}

			}

		}
	}()
}

//Pause pauses the worker controller
func (wc *workerController) Pause() {

	go func() {
		wc.pause <- true
	}()

	msg := telegram.NewMessage(chatID, "Pausing work!")
	_, err := telegramAPI.SendMessage(context.Background(), msg)
	if err != nil {
		log.Println(err)
	}

}

func (wc *workerController) Stop() {
	go func() {
		wc.quit <- true
	}()
}

type job struct {
	EncryptionDetails types.EncryptionDetails
	Path              types.Path
}

func (j *job) Run() (*types.NodeResponse, error) {
	fmt.Println("job running")

	if *wc.IsPaused {
		log.Println("Worker controller is currently paused")
		return nil, ErrTooManyRequests
	}

	//add jobs to a retry queue or something if worker controller is currently paused

	newPath, err := mycrypto.DecryptWithRounds(
		j.EncryptionDetails.Key,
		&j.Path.CipherID,
		j.Path.Rounds,
	)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", os.Getenv("BASE_URL")+newPath, nil)
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
	msg := telegram.NewMessage(chatID, strconv.Itoa(resp.StatusCode))
	_, err = telegramAPI.SendMessage(context.Background(), msg)
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode == 208 {
		fmt.Println("treasure has already been claimed...")
	} else if resp.StatusCode == 302 {
		msg := telegram.NewMessage(chatID, "Successfully discovered treasure at node "+newPath)
		_, err = telegramAPI.SendMessage(context.Background(), msg)
		if err != nil {
			log.Println(err)
		}
	} else if resp.StatusCode == 429 {
		return nil, ErrTooManyRequests
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeResponse types.NodeResponse

	err = json.Unmarshal(respBody, &nodeResponse)
	if err != nil {
		return nil, err
	}

	for _, v := range nodeResponse.Paths {
		//add them to the worker pool
		job := job{nodeResponse.Encryption, v}
		jobQueue <- job
	}

	return nil, nil
}
