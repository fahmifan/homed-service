package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"
	"time"

	"gitlab.com/homed/homed-service/utils"

	"github.com/boltdb/bolt"

	log "github.com/sirupsen/logrus"
	"gitlab.com/homed/homed-service/model"
)

// ErrVideoNotFound :nodoc:
var ErrVideoNotFound = errors.New("video not found")

// VideoRepository :nodoc:
type VideoRepository interface {
	Create(ctx context.Context, reader *multipart.Reader, video *model.Video) error
	Recreate(ctx context.Context, reader *multipart.Reader, id int64, video *model.Video) error
	Update(ctx context.Context, videoID int64, video *model.Video) error
	FindAll(ctx context.Context) ([]*model.Video, error)
	DeleteByID(ctx context.Context, id int64) (*model.Video, error)
	FindByTitle(ctx context.Context, title string) ([]*model.Video, error)
	FindByID(ctx context.Context, id int64) (*model.Video, error)
}

type videoRepository struct {
	db *bolt.DB
}

// NewVideo :nodoc:
func NewVideo(db *bolt.DB) VideoRepository {
	return &videoRepository{db: db}
}

func (r *videoRepository) SaveVideo(ctx context.Context, reader *multipart.Reader, videoID int64) (fileName, path string, err error) {
	dir := fmt.Sprintf("videos/%d", videoID)

	for {
		var part *multipart.Part
		part, err = reader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" {
			continue
		}

		fileName = part.FileName()
		path = dir + "/" + part.FileName()
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				panic(err)
			}
		}

		var dst *os.File
		dst, err = os.Create(path)
		if err != nil {
			log.Error(err)
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, part); err != nil {
			log.Error(err)
			return
		}
	}

	return
}

func (r *videoRepository) Create(ctx context.Context, reader *multipart.Reader, video *model.Video) (err error) {
	video.ID = time.Now().Unix()
	dir := fmt.Sprintf("videos/%d", video.ID)
	fileName, path, err := r.SaveVideo(ctx, reader, video.ID)
	if err != nil && err != io.EOF {
		log.Error(err)
		return
	}

	fileNameOri := fileName[:len(fileName)-4]
	ext := fileName[len(fileName)-4:]

	video.Name = fileNameOri
	video.Ext = ext
	video.CreatedAt = time.Now()
	video.UpdatedAt = video.CreatedAt

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())

		videoBytes := video.Marshall()
		err = b.Put(utils.Int64ToBytes(video.ID), videoBytes)
		log.Info(video.ID, string(videoBytes))
		return err
	})

	if err != nil {
		log.WithFields(log.Fields{
			"source":  path,
			"context": ctx,
		}).Error(err)
		return err
	}

	playlistName := fmt.Sprintf("%d.m3u8", video.ID)
	playListPath := dir + "/" + playlistName
	go r.createHLS(path, playListPath)

	return nil
}

func (r *videoRepository) Recreate(ctx context.Context, reader *multipart.Reader, id int64, video *model.Video) (err error) {
	var fileName, sourcePath string
	dir := fmt.Sprintf("videos/%d", id)

	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())
		bid := utils.Int64ToBytes(id)
		v := b.Get(bid)

		if string(v) == "" {
			return ErrVideoNotFound
		}

		*video = *(model.NewVideoFromBytes(v))
		if video.DeletedAt != nil {
			return ErrVideoNotFound
		}

		// remove source video
		err = os.RemoveAll(dir)
		if err != nil {
			log.Error(err)
			return err
		}

		fileName, sourcePath, err = r.SaveVideo(ctx, reader, id)

		fileNameOri := fileName[:len(fileName)-4]
		ext := fileName[len(fileName)-4:]

		video.Name = fileNameOri
		video.Ext = ext
		video.UpdatedAt = time.Now()

		videoBytes := video.Marshall()
		err = b.Put(utils.Int64ToBytes(video.ID), videoBytes)
		log.Info(video.ID, string(videoBytes))
		return err
	})

	if err != nil {
		log.WithFields(log.Fields{
			"source":  sourcePath,
			"context": ctx,
		}).Error(err)
		return err
	}

	playlistName := fmt.Sprintf("%d.m3u8", video.ID)
	playListPath := dir + "/" + playlistName
	go r.createHLS(sourcePath, playListPath)

	return nil
}

func (r *videoRepository) Update(ctx context.Context, videoID int64, video *model.Video) (err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())

		bid := utils.Int64ToBytes(videoID)

		v := b.Get(bid)
		if string(v) == "" {
			return ErrVideoNotFound
		}

		currentVideo := model.NewVideoFromBytes(v)
		if currentVideo.DeletedAt != nil {
			return ErrVideoNotFound
		}

		video.ID = currentVideo.ID
		video.CreatedAt = currentVideo.CreatedAt
		video.UpdatedAt = time.Now()

		videoBytes := video.Marshall()
		err = b.Put(bid, videoBytes)
		log.Info(video.ID, string(videoBytes))
		return err
	})

	return
}

func (r *videoRepository) DeleteByID(ctx context.Context, id int64) (video *model.Video, err error) {
	err = r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())
		v := b.Get(utils.Int64ToBytes(id))

		if string(v) == "" {
			return ErrVideoNotFound
		}

		video = model.NewVideoFromBytes(v)
		if video.DeletedAt != nil {
			return ErrVideoNotFound
		}

		t := time.Now()
		video.DeletedAt = &t

		videoBytes := video.Marshall()
		log.Info(video.ID, videoBytes)
		err := b.Put(utils.Int64ToBytes(video.ID), videoBytes)
		return err
	})

	if err != nil {
		log.WithFields(log.Fields{
			"id": id,
		}).Error(err)
		return nil, err
	}

	return
}

func (r *videoRepository) FindAll(ctx context.Context) (videos []*model.Video, err error) {
	err = r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())
		b.ForEach(func(k, v []byte) error {
			if string(v) == "" {
				return nil
			}

			video := model.NewVideoFromBytes(v)
			if video.DeletedAt == nil {
				videos = append(videos, video)
			}

			return nil
		})

		return nil
	})

	if err != nil {
		log.Error(err)
	}

	return
}

func (r *videoRepository) FindByTitle(ctx context.Context, title string) (videos []*model.Video, err error) {
	r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())

		err = b.ForEach(func(k, v []byte) error {
			if string(v) == "" {
				return nil
			}

			video := model.NewVideoFromBytes(v)
			if strings.ToLower(video.Name) == strings.ToLower(title) {
				videos = append(videos, video)
				return nil
			}

			return nil
		})

		if err != nil {
			log.WithFields(log.Fields{"title": title}).Error(err)
			return err
		}

		return nil
	})

	return
}

func (r *videoRepository) createHLS(sourcePath, destPath string) {
	opt := []string{"-i", sourcePath, "-c:a", "aac", "-strict", "experimental", "-c:v", "libx264", "-f", "hls", "-hls_time", "60", "-hls_list_size", "0", destPath}

	now := time.Now()
	log.Infoln("ffmpeg", strings.Join(opt, " "))
	cmd := exec.Command("ffmpeg", opt...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err, ": ", string(output))
		return
	}
	log.Info(fmt.Sprintf("took: %.2f minutes | finished create playlist %s", time.Since(now).Minutes(), destPath))

	// remove source video
	if err := os.Remove(sourcePath); err != nil {
		log.Error(err)
		return
	}

	log.Info("success remove source: " + sourcePath)
}

// FindByID :nodoc:
func (r *videoRepository) FindByID(ctx context.Context, id int64) (*model.Video, error) {
	var video *model.Video
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(model.VideoBucket())
		v := b.Get(utils.Int64ToBytes(id))

		if string(v) == "" {
			return nil
		}

		video = model.NewVideoFromBytes(v)
		if video.DeletedAt != nil {
			return nil
		}

		return nil
	})

	return video, err
}
