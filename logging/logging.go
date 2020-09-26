package logging

import (
	"encoding/json"
	"io/ioutil"
	"os"
	pathLib "path"
	"strconv"
)

//LogSuccesfulHit logs details of a successful hit back to me with a telegram bot or something
func LogHitsToFile(pathName string) error {
	initLogFile()

	//get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	//make secret.ly directory
	path := pathLib.Join(wd, "results.json")

	var content map[string]string
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	json.Unmarshal(data, &content)
	pos := strconv.Itoa(len(content) + 1)

	content[pos] = pathName

	bs, err := json.Marshal(content)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, bs, 0660)
	if err != nil {
		return err
	}
	return nil
}

//local stuff sha
//cam remove later
func initLogFile() error {
	/*home, err := homedir.Dir()
	if err != nil {
		return Vault{}, err
	}*/

	//get working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	//make secret.ly directory
	path := pathLib.Join(wd, "results.json")
	err = os.MkdirAll(path, 0660) //creates path if it doesn't exist
	if err != nil {
		return err
	}

	var data []byte

	for {
		data, err = ioutil.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				_, err = os.Create(path)
				if err != nil {
					return err
				}
				err = writeEmptyMapToFile(path)
				if err != nil {
					return err
				}
			} else {
				return err
			}

		} else {
			break
		}
	}

	//write empty map to file if file's empty
	if string(data) == "" {
		err = writeEmptyMapToFile(path)
		if err != nil {
			return err
		}
	}

	return nil

}

func writeEmptyMapToFile(path string) error {
	bs, err := json.Marshal(map[string]string{})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, bs, 0660)
	if err != nil {
		return err
	}
	return nil
}
