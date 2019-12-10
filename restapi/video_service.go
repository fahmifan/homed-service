package restapi

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"gitlab.com/homed/homed-service/utils"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/homed/homed-service/model"
	"gitlab.com/homed/homed-service/repository"
)

const maxUploadSize = 1 * 1024 * 1024 // 2 mb

// VideoService :nodoc:
type VideoService struct {
	videoRepo repository.VideoRepository
}

// NewVideo :nodoc:
func NewVideo(v repository.VideoRepository) *VideoService {
	return &VideoService{
		videoRepo: v,
	}
}

// ServeHLSM3U8 :nodoc:
func (s *VideoService) ServeHLSM3U8(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	mediaFile := fmt.Sprintf("videos/%s/%s.m3u8", videoID, videoID)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "application/x-mpegURL")
}

// ServeHLSTs :nodoc:
func (s *VideoService) ServeHLSTs(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	ts := chi.URLParam(r, "ts")
	mediaFile := fmt.Sprintf("videos/%s/%s", videoID, ts)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "video/MP2T")
}

// Find find all videos
func (s *VideoService) Find(w http.ResponseWriter, r *http.Request) {
	var videos []*model.Video
	var err error
	titleVals, titleOK := r.URL.Query()["title"]

	ctx := context.Background()
	switch {
	case titleOK && len(titleVals) > 0 && titleVals[0] != "":
		videos, err = s.videoRepo.FindByTitle(ctx, titleVals[0])
	default:
		videos, err = s.videoRepo.FindAll(ctx)
	}

	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if len(videos) == 0 {
		writeError(w, errors.New("video not found"), http.StatusNotFound)
		return
	}

	writeJSON(w, videos)
}

// FindByID :nodoc:
func (s *VideoService) FindByID(w http.ResponseWriter, r *http.Request) {
	var err error
	videoID := utils.String2Int64(chi.URLParam(r, "id"))

	ctx := context.Background()
	video, err := s.videoRepo.FindByID(ctx, videoID)
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if video == nil {
		writeError(w, err, http.StatusNotFound)
		return
	}

	writeJSON(w, video)
}

// Create :nodoc:
func (s *VideoService) Create(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	video := &model.Video{}
	err = s.videoRepo.Create(context.Background(), reader, video)
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, &video)
}

// Recreate :nodoc:
func (s *VideoService) Recreate(w http.ResponseWriter, r *http.Request) {
	videoID := utils.String2Int64(chi.URLParam(r, "id"))
	if videoID <= 0 {
		writeError(w, errors.New("invalid id"), http.StatusBadRequest)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	video := model.Video{}
	err = s.videoRepo.Recreate(context.Background(), reader, videoID, &video)
	log.Println(video)
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	writeJSON(w, &video)
}

// DeleteByID :nodoc:
func (s *VideoService) DeleteByID(w http.ResponseWriter, r *http.Request) {
	videoID := utils.String2Int64(chi.URLParam(r, "id"))
	if videoID <= 0 {
		writeError(w, errors.New("invalid id"), http.StatusBadRequest)
		return
	}

	video, err := s.videoRepo.DeleteByID(context.Background(), videoID)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJSON(w, video)
}

// Update :nodoc:
func (s *VideoService) Update(w http.ResponseWriter, r *http.Request) {
	videoID := utils.String2Int64(chi.URLParam(r, "id"))
	if videoID <= 0 {
		writeError(w, errors.New("invalid id"), http.StatusBadRequest)
		return
	}

	var video model.Video
	json.NewDecoder(r.Body).Decode(&video)

	err := s.videoRepo.Update(context.Background(), videoID, &video)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	writeJSON(w, video)
}

// UploadCover :nodoc:
func (s *VideoService) UploadCover(w http.ResponseWriter, r *http.Request) {
	// validate file size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, errors.New("file too big"), http.StatusBadRequest)
		return
	}

	// parse and validate file and post parameters
	file, _, err := r.FormFile("cover")
	if err != nil {
		writeError(w, errors.New("invalid file"), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		writeError(w, errors.New("invalid file"), http.StatusBadRequest)
		return
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "image/jpeg", "image/jpg":
	case "image/gif", "image/png":
	case "application/pdf":
		break
	default:
		writeError(w, errors.New("invalid file"), http.StatusBadRequest)
		return
	}
	fileName := randToken(18)
	fileEndings, err := mime.ExtensionsByType(detectedFileType)
	if err != nil {
		writeError(w, errors.New("can't read file type"), http.StatusBadRequest)
		return
	}
	coverName := fileName + fileEndings[0]
	newPath := filepath.Join("cover", coverName)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		writeError(w, errors.New("can't write file"), http.StatusBadRequest)
		return
	}

	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		writeError(w, errors.New("can't write file"), http.StatusBadRequest)
		return
	}

	resp := map[string]string{
		"cover": coverName,
	}

	writeJSON(w, resp)
}

// ServeCover :nodoc:
func (s *VideoService) ServeCover(w http.ResponseWriter, r *http.Request) {
	cover := chi.URLParam(r, "cover")
	mediaFile := filepath.Join("cover", cover)
	http.ServeFile(w, r, mediaFile)
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
