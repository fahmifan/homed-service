package restapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/boltdb/bolt"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
	"gitlab.com/homed/homed-service/config"
	"gitlab.com/homed/homed-service/db"
	"gitlab.com/homed/homed-service/repository"
)

// Server :nodoc:
type Server struct {
	router *chi.Mux

	boltdb *bolt.DB

	videoRepo    repository.VideoRepository
	videoService *VideoService
}

// NewServer :nodoc:
func NewServer() *Server {
	s := &Server{}
	s.boltdb = db.NewBoltDB()
	s.videoRepo = repository.NewVideo(s.boltdb)
	s.videoService = NewVideo(s.videoRepo)
	s.router = chi.NewRouter()
	return s
}

func (s *Server) initRouter() {
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r := s.router
	r.Use(cors.Handler)
	r.Get("/ping", ping)
	r.Get("/api/videos", s.videoService.Find)
	r.Post("/api/videos", s.videoService.Create)
	r.Post("/api/videos/{id}/recreate", s.videoService.Recreate)
	r.Delete("/api/videos/{id}", s.videoService.DeleteByID)
	r.Put("/api/videos/{id}", s.videoService.Update)
	r.Get("/media/{id:[0-9]+}/{ts:[0-9]+.ts}", s.videoService.ServeHLSTs)
	r.Get("/media/{id:[0-9]+}/stream", s.videoService.ServeHLSM3U8)

	FileServer(r, "/videos")
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string) {
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "videos")

	root := http.Dir(filesDir)
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

// Run run the server
func (s *Server) Run() {
	s.initRouter()

	log.Println("http service listening on :" + config.Port())
	http.ListenAndServe(":"+config.Port(), s.router)
}

func ping(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"ping": "pong"})
}
