package models

import (
	"fmt"
	"strconv"
	"time"
)

type WorkerControllerState string

const (
	WorkerControllerUninitialised WorkerControllerState = WorkerControllerState("Uninitialised")
	WorkerControllerRunning WorkerControllerState = WorkerControllerState("Running")
	WorkerControllerPaused WorkerControllerState = WorkerControllerState("Paused")
	WorkerControllerStopped WorkerControllerState = WorkerControllerState("Stopped")
)

//WorkerController is where the main stuff happens. We can pause and resume work from here and it also houses the worker poo.
type WorkerController struct {
	IsPaused   *bool
	pause      chan bool
	MaxWorkers int
	RetryQueue chan Job
	JobQueue chan Job
	RetryCount int
	WorkerPool chan chan Job
	WorkersCreatedChannel chan bool
	quit       chan bool
	MessageChan chan string
	State WorkerControllerState
}

func NewWorkerController(maxWorkers int) *WorkerController {
	f := false
	return &WorkerController{
		&f,
		make(chan bool),
		maxWorkers,
		make(chan Job),
		make(chan Job),
		0,
		make(chan chan Job, maxWorkers),
		make(chan bool),
		make(chan bool),
		make(chan string),
		WorkerControllerUninitialised,
	}
}

//Run starts the worker controller
func (wc *WorkerController) Run() {


	for i := 0; i < wc.MaxWorkers; i++ {
		//create and start new workers
		fmt.Println("creating workers...")
		worker := newWorker(*wc)
		worker.Start()

	}
	wc.WorkersCreatedChannel <- true

	wc.State = WorkerControllerRunning

	go func() {
		for {
			select {
			case j := <- wc.JobQueue:
				fmt.Println("received Job")
				go func(j Job) {
					if *wc.IsPaused == false {
						JobChannel := <-wc.WorkerPool

						JobChannel <- j
					} else {
						fmt.Println("hereee")
						wc.RetryCount++
						wc.RetryQueue <- j

					}

					fmt.Println("Job has been sent to worker")
				}(j)

			case <-wc.quit:
				fmt.Println("quitting")
				if wc.State == WorkerControllerUninitialised || wc.State == WorkerControllerStopped{
					wc.MessageChan <- "WorkerController cannot be stopped. State: "+wc.GetState()
				} else{
					wc.State = WorkerControllerStopped
					return
				}



			case <-wc.pause:
				if *wc.IsPaused == false {
					fmt.Println("Pausing for a minute...")
					*wc.IsPaused = true
					wc.State = WorkerControllerPaused
					go func() {
						select {
						case <-time.After(1 * time.Minute):
							*wc.IsPaused = false
							wc.State = WorkerControllerRunning
							fmt.Println("Resume work")
							wc.MessageChan <-  "Resuming work!, Retry Count: "+strconv.Itoa(wc.RetryCount)

							for j := range wc.RetryQueue {
								wc.JobQueue <- j
								wc.RetryCount--
							}
						}

					}()
				} else {
					wc.MessageChan <- "Already paused"
				}

			}

		}
	}()
}

//Pause pauses the worker controller
func (wc *WorkerController) Pause() {

	go func() {
		wc.pause <- true
	}()

	wc.MessageChan <- "Pausing work!"
}

func (wc *WorkerController) Stop() {
	go func() {
		wc.quit <- true
	}()
}

func (wc *WorkerController) GetState() string{
	return string(wc.State)
}