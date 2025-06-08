package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unable to parse form file", err)
		return
	}

	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "video not found", err)
		return
	}

	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user not video author", err)
		return
	}

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid content type", err)
		return
	}

	if mediatype != "image/jpeg" && mediatype != "image/png" {
		respondWithError(w, http.StatusBadRequest, "unsupported media type", nil)
		return
	}

	ext := strings.Split(mediatype, "/")
	if len(ext) != 2 {
		respondWithError(w, http.StatusBadRequest, "invalid media type", nil)
		return
	}

	thumbnailPath := filepath.Join(cfg.assetsRoot, video.ID.String()+"."+ext[1])

	thumbnailFile, err := os.Create(thumbnailPath)
	if err != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to create file", err)
		return
	}

	defer thumbnailFile.Close()

	_, err1 := io.Copy(thumbnailFile, file)
	if err1 != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to copy data to file", err)
		return
	}

	err3 := thumbnailFile.Sync()
	if err3 != nil {
		respondWithError(w, http.StatusNotImplemented, "unable to store file", err)
		return
	}

	thumbnailURL := "/" + thumbnailPath

	video.UpdatedAt = time.Now()
	video.ThumbnailURL = &thumbnailURL

	err2 := cfg.db.UpdateVideo(video)
	if err2 != nil {
		respondWithError(w, http.StatusBadRequest, "unable to update video", err2)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
