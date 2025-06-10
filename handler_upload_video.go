package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	http.MaxBytesReader(w, r.Body, 1 << 30)

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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user not video author", err)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}

	contentType := header.Header.Get("Content-Type")

	defer file.Close()

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid content type", err)
		return
	}

	if mediatype != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "unsupported media type", nil)
		return
	}

	ext := strings.Split(mediatype, "/")
	if len(ext) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid media type", nil)
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to create file", err)
		return
	}

	defer os.Remove(tmpFile.Name())

	_, err1 := io.Copy(tmpFile, file)
	if err1 != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to copy data to file", err1)
		return
	}

	err3 := tmpFile.Sync()
	if err3 != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to store file", err)
		return
	}

	_, err2 := tmpFile.Seek(0, io.SeekStart)
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, "unable to set seek offset", err2)
		return
	}
	
	defer tmpFile.Close()

	aspectRatio, err := getVideoAspectRatio(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to get aspect ratio", err)
		return
	}


	randNum := make([]byte, 32)
	rand.Read(randNum)
	videoKey := base64.StdEncoding.EncodeToString(randNum)+"."+ext[1]
	awsVideoKey := "other/"+videoKey

	if aspectRatio == "16:9" {
		awsVideoKey = "landscape/"+videoKey
	}
	if aspectRatio == "9:16" {
		awsVideoKey = "portrait/"+videoKey
	}

	_, err4 := cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
							Bucket: aws.String("tubely-65731"),
							Key: aws.String(awsVideoKey),
							Body: tmpFile,
							ContentType: aws.String(mediatype),
						})
	if err4 != nil {
		respondWithError(w, http.StatusBadRequest, "unable to put ojbect in bucket", err3)
		return
	}

	videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, awsVideoKey)

	video.VideoURL = &videoURL
	video.UpdatedAt = time.Now()

	err5 := cfg.db.UpdateVideo(video)
	if err5 != nil {
		respondWithError(w, http.StatusBadRequest, "unable to update video", err5)
		return
	}

	respondWithJSON(w, http.StatusOK, video)

}