package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const maxMemory = 1 << 30 

	r.Body = http.MaxBytesReader(w , r.Body , maxMemory)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	videoData , err  := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w , 500 , "Error fetching metadata" , err)
		return 
	}

	if userID != videoData.UserID {
		respondWithError(w , 403 , "Unauthorized" , nil)
		return 
	}

	data , header , err := r.FormFile("video")
	if err != nil {
		respondWithError(w , http.StatusBadRequest , "Error fetching file data and headers" , err)
		return 
	}

	defer data.Close()

	mediaType , _ , err  := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w , 400 , "Bad Input for Content-Type Header" , err)
		return 
	}

	if mediaType != "video/mp4" {
		respondWithError(w , 400 , "Error uploading the File. Only mp4 is accepted" , err)
		return 
	}

	tempFile , err := os.CreateTemp("" , "tubely-upload-*.mp4" )
	if err != nil {
		respondWithError(w , 500 , "Error Creating a Temp Storage File" , err)
		return 
	}

	
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_ , err = io.Copy(tempFile , data)
	if err != nil {
		respondWithError(w , 500 , "Error copying the multipart.File to the new file" , err)
		return 
	}

	_ , err = tempFile.Seek(0 , io.SeekStart)
	if err != nil {
		respondWithError(w , 500 , "Error resetting the temp files pointer" , err)
		return 
	}

	vidKeyBytes := make([]byte , 32)
	rand.Read(vidKeyBytes)

	vidKey := base64.RawURLEncoding.EncodeToString(vidKeyBytes)

	aspectRatio , err := getVideoAspectRatio(tempFile.Name())


	var aspectPrefix string

	switch aspectRatio {
	case "16:9":
		aspectPrefix = "landscape"
	case "9:16":
		aspectPrefix = "portrait"
	default:
		aspectPrefix = "other"
	}




	vidKeyExtension :=  aspectPrefix + "/" + vidKey + ".mp4" 

	_ , err = cfg.s3Client.PutObject(r.Context() , &s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &vidKeyExtension,
		Body: tempFile,
		ContentType: &mediaType,
	})
	if err != nil {
		respondWithError(w , 500 , "Server side Issue" , err)
		return 
	}

	vidUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s" , cfg.s3Bucket , cfg.s3Region , vidKeyExtension)

	videoData.VideoURL = &vidUrl 

	fmt.Println(vidUrl)



	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w , 500 , "Error Updating the video" , err)
		return
	}

	fmt.Println(videoData)

	respondWithJSON(w , 200 , videoData)
}
