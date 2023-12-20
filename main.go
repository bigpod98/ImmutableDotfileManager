package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type config struct {
	Locations	[]location
}

type location struct {
	name	string	`json:"name"`
	dstLocation	string	`json:"dstlocation"`
	srcLocation	string	`json:"srclocation"`
}

func main() {
	homedir := os.Getenv("HOME")
	filedata, err := os.ReadFile(fmt.Sprintf("%s/.idotfiles.json", homedir))
	check(err)
	conf := config{}
	json.Unmarshal([]byte(filedata), &conf)

	for _,i := range conf.Locations {
		finalCommand := fmt.Sprintf("-t overlay overlay -o lowerdir=%s,workdir=%s/.local/%s-workdir,upper=%s/.local/%s-mutable %s", i.srcLocation,homedir,i.name,homedir,i.name, i.dstLocation)
		exet := exec.Command("mount", finalCommand)
		exet.Wait()
		output, err := exet.Output()
		check(err)

		fmt.Println(string(output))
	}
}
func check(err error) {
	if(err != nil) {
		fmt.Println(err)
	}
}

