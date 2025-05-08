package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var (
	tmdbMovieGenreURL string
	tmdbTvGenreURL    string
)

// TMDBGenre represents a genre from TMDB
type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GenreResponse is the envelope TMDB returns
type GenreResponse struct {
	Genres []TMDBGenre `json:"genres"`
}

func init() {
	// Load environment
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env non chargé: %v", err)
	}
	tmdbMovieGenreURL = "https://api.themoviedb.org/3/genre/movie/list"
	tmdbTvGenreURL = "https://api.themoviedb.org/3/genre/tv/list"
	// Schedule both movie and tv genre sync at midnight daily
	c := cron.New()
	_, err := c.AddFunc("2 * * * *", func() {
		log.Println("🚀 Running SyncMovieGenres and SyncTvGenres")
		SyncMovieGenres()
		SyncTvGenres()
	})
	if err != nil {
		log.Fatalf("Erreur planification cron: %v", err)
	}
	c.Start()
}


// SyncMovieGenres retrieves movie genres and syncs to Strapi
func SyncMovieGenres() {
	strapiTvURL := os.Getenv("STRAPI_URL") + "/api/genre-tv-shows"

	log.Println("🔄 SyncMovieGenres start")
	syncGenres(tmdbMovieGenreURL, strapiTvURL)
	log.Println("✅ SyncMovieGenres done")
}

// SyncTvGenres retrieves TV genres and syncs to Strapi
func SyncTvGenres() {
	strapiTvURL := os.Getenv("STRAPI_URL") + "/api/genre-tv-shows"
	log.Println("🔄 SyncTvGenres start")
	syncGenres(tmdbTvGenreURL, strapiTvURL)
	log.Println("✅ SyncTvGenres done")

}

// syncGenres is shared logic for TMDB -> Strapi
func syncGenres(tmdbURL, strapiURL string) {
	strapiToken := os.Getenv("STRAPI_TOKEN")

	start := time.Now()

	resp, err := http.Get(fmt.Sprintf("%s?api_key=%s&language=fr-FR", tmdbURL, os.Getenv("API_KEY")))
	if err != nil {
		log.Printf("❌ TMDB GET error: %v", err)
		return
	}
	defer resp.Body.Close()

	var tmdbRes GenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbRes); err != nil {
		log.Printf("❌ JSON decode error: %v", err)
		return
	}
	log.Printf("TMDB returned %d genres", len(tmdbRes.Genres))
    endpoint := os.Getenv("STRAPI_URL") + "/api/genre-tv-shows?" + "filters[id_genre][$eq]"
	for _, g := range tmdbRes.Genres {
		exists, err := Exists(g.ID,endpoint)
		if err != nil {
			log.Printf("⚠️ check exists error for %d: %v", g.ID, err)
			continue
		}
		if exists {
			log.Printf("ℹ️ skipped existing genre: %s (%d)", g.Name, g.ID)
			continue
		}

		payload := map[string]interface{}{"data": map[string]interface{}{"id_genre": g.ID, "nom_genre": g.Name}}
		body, _ := json.Marshal(payload)
       
		req, _ := http.NewRequest("POST", strapiURL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+strapiToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("❌ Strapi POST error for %s: %v", g.Name, err)
			continue
		}
		res.Body.Close()

		if res.StatusCode >= 400 {
			log.Printf("⚠️ Strapi returned %d for %s", res.StatusCode, g.Name)
		} else {
			log.Printf("✅ inserted genre: %s (%d)", g.Name, g.ID)
		}

	}

	log.Printf("Sync complete in %s", time.Since(start))
}

// GenreTVShowHandler triggers manual sync of both movie and tv genres
func GenreTVShowHandler(w http.ResponseWriter, r *http.Request) {
	go SyncMovieGenres()
	go SyncTvGenres()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sync triggered")
}

