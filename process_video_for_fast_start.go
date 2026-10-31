package main

import (
	"os/exec"
)

func processVideoForFastStart(filepath string) (string , error ) {
	outputFilePath := filepath + ".processing"

	cmd := exec.Command("ffmpeg" , "-i" , filepath , "-c" , "copy" , "-movflags" , "faststart" , "-f" , "mp4" , outputFilePath )

	err := cmd.Run()
	if err != nil {
		return "" , err
	}

	return outputFilePath , nil 
}