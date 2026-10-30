package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)


func getVideoAspectRatio (filePath string) (string , error) {
	cmd :=  exec.Command("ffprobe" , "-v" , "error" , "-print_format" , "json" , "-show_streams" , filePath )

	var dataBytes bytes.Buffer

	cmd.Stdout = &dataBytes

	err := cmd.Run()
	if err != nil {
		return "" , err 
	}

	type VideoInfo struct {
		Streams []struct {
			Width int `json:"width,omitempty"`
			Height int `json:"height,omitempty"`
		} `json:"streams"`
	}

	var vidInfo VideoInfo

	err = json.Unmarshal(dataBytes.Bytes() , &vidInfo)
	if err != nil {
		return "" , err
	}

	var ratio string 

	

	for _ , elem := range vidInfo.Streams {
		if elem.Width == 0 && elem.Height == 0 {
			continue 
		}
		ratioFloat := float64(elem.Width) / float64(elem.Height)
		if ratioFloat >= 1.76 && ratioFloat <= 1.79 {
			ratio = "16:9"
		}else if ratioFloat >= 0.55 && ratioFloat <= 0.57 {
			ratio = "9:16"
		}else {
			ratio = "other"
		}
		fmt.Println(elem , ratioFloat)
	}

	fmt.Println(ratio)

	return ratio , nil 

}