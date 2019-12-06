package model

import (
	"encoding/json"
	"time"
)

// videoBucket boltdb bucket for videos
var videoBucket = []byte("videos")

// Video :nodoc:
type Video struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Ext       string     `json:"ext"`
	Cover     string     `json:"cover"`
	Category  string     `json:"category"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

// Marshall video into json
func (v *Video) Marshall() []byte {
	b, _ := json.Marshal(v)
	return b
}

// NewVideoFromBytes :nodoc:
func NewVideoFromBytes(b []byte) *Video {
	video := &Video{}
	_ = json.Unmarshal(b, video)

	return video
}

// VideoBucket :nodoc:
func VideoBucket() []byte {
	return videoBucket
}
