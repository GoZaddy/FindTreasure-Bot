package models

import (
	"errors"
	"fmt"
	"github.com/gozaddy/findtreasure/cache"
	"github.com/gozaddy/findtreasure/customerrors"
	"log"
	"time"
)

type worker struct {
	workerController WorkerController
	jobChannel       chan Job
	quit             chan bool
}

func newWorker(wc WorkerController) worker {
	return worker{
		wc,
		make(chan Job),
		make(chan bool),
	}
}

func (w worker) Start() {
	go func() {
		for {
			//register current work into worker pool
			w.workerController.WorkerPool <- w.jobChannel

			select {
			case currentJob := <-w.jobChannel:
				fmt.Println("worker has received job")
				if *w.workerController.IsPaused == false {
					if _, err := cache.Get(currentJob.Path.CipherID); err != nil{
						//check cache to see if we've run this job in the past
						if errors.Is(err, customerrors.ErrNilRedisValue){
							nodeResponse, message, err := currentJob.Run()
							if err != nil {
								if errors.Is(err, customerrors.ErrTooManyRequests) {
									fmt.Println("heree")
									if *w.workerController.IsPaused == false {
										w.workerController.Pause()
									}
									w.workerController.RetryCount++
									w.workerController.RetryQueue <- currentJob

								} else {
									log.Println(err)
								}
							}
							fmt.Println("ran Job")

							//add job cipher id to cache so we won't call it again
							err = cache.Set(currentJob.Path.CipherID, time.Now().Unix(), "86400")
							if err != nil{
								w.workerController.MessageChan <- "An error just occurred: "+err.Error()
								log.Fatalln(err)
							}

							if message != ""{
								w.workerController.MessageChan <- message
							}
							for _, v := range nodeResponse.Paths {
								//add them to the worker pool
								newJob := Job{nodeResponse.Encryption, v}
								w.workerController.JobQueue <- newJob
							}
						} else {
							w.workerController.MessageChan <- "An error occurred: "+err.Error()
						}

					}


				} else {
					fmt.Println("hereeeee")
					w.workerController.MessageChan <- "testing 1"
					w.workerController.RetryCount++
					w.workerController.RetryQueue <- currentJob
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

