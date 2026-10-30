package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {



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


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w , 500 , "Couldnt Parse Max Memory" , err)
	}

	data , header , err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w , http.StatusBadRequest , "Error fetching file data and headers" , err)
		return 
	}

	mediaType , _ , err  := mime.ParseMediaType(header.Header.Get("Content-Type"))

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w , 400 , "Error uploading the File. Upload .jpeg / .png files only" , err)
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

	fileExtension := strings.Split(mediaType , "image/")
	fileExtensionString := strings.Join(fileExtension , "")

	thumbIDBytes := make([]byte , 32)
	rand.Read(thumbIDBytes)

	thumbID := base64.RawURLEncoding.EncodeToString(thumbIDBytes)


	pathElement := fmt.Sprintf("%s.%s" , thumbID , fileExtensionString)
	
	filePath := filepath.Join("."  , cfg.assetsRoot , pathElement)

	destFile , err := os.Create(filePath)
	if err != nil {
		respondWithError(w , 500 , "Error Creating a FilePath" , err)
		return 
	}

	_ , err = io.Copy(destFile , data)
	if err != nil {
		respondWithError(w , 500 , "Error copying the multipart.File to the new file" , err)
		return 
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/%s", cfg.port , filePath)

	videoData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w , 500 , "Error Updating the video" , err)
		return
	}

	defer data.Close()
	defer destFile.Close()

	respondWithJSON(w, http.StatusOK, videoData)
}

