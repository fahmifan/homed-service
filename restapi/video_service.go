package restapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"gitlab.com/homed/homed-service/utils"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/homed/homed-service/model"
	"gitlab.com/homed/homed-service/repository"
)

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

// FindAll find all videos
func (s *VideoService) FindAll(w http.ResponseWriter, r *http.Request) {
	videos, err := s.videoRepo.FindAll(context.Background())
	if err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, videos); err != nil {
		log.Error(err)
		writeError(w, err, http.StatusInternalServerError)
	}
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
