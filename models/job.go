package models

import (
	"encoding/json"
	"fmt"
	"github.com/gozaddy/findtreasure/customerrors"
	"github.com/gozaddy/findtreasure/mycrypto"
	"github.com/gozaddy/findtreasure/types"
	"io/ioutil"
	"net/http"
	"os"
)

type Job struct {
	EncryptionDetails types.EncryptionDetails
	Path              types.Path
}

func (j *Job) Run() (types.NodeResponse, string, error) {
	fmt.Println("Job running")

	//add Jobs to a retry  queue or something if worker controller is currently paused

	newPath, err := mycrypto.DecryptWithRounds(
		j.EncryptionDetails.Key,
		&j.Path.CipherID,
		j.Path.Rounds,
	)

	if err != nil {
		return types.NodeResponse{}, "", err
	}

	req, err := http.NewRequest("GET", os.Getenv("BASE_URL")+newPath, nil)
	if err != nil {
		return types.NodeResponse{}, "", err
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("FIND_TREASURE_TOKEN"))
	req.Header.Set("accountId", "faru")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.NodeResponse{}, "", err
	}
	defer resp.Body.Close()

	var message string

	if resp.StatusCode == 208 {
		fmt.Println("treasure has already been claimed...")
		message = ""
	} else if resp.StatusCode == 302 {
		message = "Successfully discovered treasure at node "+newPath
	} else if resp.StatusCode == 429 {
		return types.NodeResponse{}, "", customerrors.ErrTooManyRequests
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return types.NodeResponse{}, "", err
	}

	fmt.Println(string(respBody))

	var nodeResponse types.NodeResponse

	err = json.Unmarshal(respBody, &nodeResponse)
	if err != nil {
		return types.NodeResponse{}, "", err
	}

	return nodeResponse, message, nil

}
