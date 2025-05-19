package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"io"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var (
	tmdbMovieGenreURL string
	tmdbTvGenreURL    string
	strapiTvURL       string
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
    strapiTvURL = os.Getenv("STRAPI_URL") + "/api/genre-tv-shows"

	// Schedule both movie and tv genre sync at midnight daily
	c := cron.New()
	_, err := c.AddFunc("0 0 * * 0", func() {
		log.Println("🚀 Exécution de SyncMovieGenres et SyncTvGenres chaque dimanche")
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

	log.Println("🔄 SyncMovieGenres start")
	syncGenres(tmdbMovieGenreURL, strapiTvURL)
	log.Println("✅ SyncMovieGenres done")
}


func SyncTvGenres() {
	log.Println("🔄 SyncTvGenres commencé ")
	syncGenres(tmdbTvGenreURL, strapiTvURL)
	log.Println("✅ SyncTvGenres terminé")

}

// syncGenres is shared logic for TMDB -> Strapi
func syncGenres(tmdbURL, strapiURL string) {
	strapiToken := os.Getenv("STRAPI_TOKEN")

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
    for _, g := range tmdbRes.Genres {
	

		payload := map[string]interface{}{"data": map[string]interface{}{"id_genre": g.ID, "nom_genre": g.Name}}
		body, _ := json.Marshal(payload)
       
		req, _ := http.NewRequest("POST", strapiURL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+strapiToken)

			// Faire la requête
			res, err := http.DefaultClient.Do(req)
			if err != nil {
			log.Printf("❌ erreur de POST  Strapi pour %s: %v", g.Name, err)
			continue
			}
			defer res.Body.Close()

			// Lire tout le corps
			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				log.Printf("❌ Lecture réponse pour %s: %v", g.Name, err)
				continue
			}

			if res.StatusCode >= 400 {
				// Tenter un décodage générique
				var data map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &data); err == nil {
					// Chercher "error" puis "message"
					if errObj, ok := data["error"].(map[string]interface{}); ok {
						if msg, ok := errObj["message"].(string); ok {
						log.Printf("⚠️ Strapi a renvoyé le code %d pour %s : %s", res.StatusCode, g.Name, msg)
							continue
						}
					}
				}

				log.Printf("⚠️ Strapi a renvoyé le code %d pour %s : %s", res.StatusCode, g.Name, string(bodyBytes))
			} else {
				log.Printf("✅ inserted genre: %s (%d)", g.Name, g.ID)
			}


	}

}

// GenreTVShowHandler triggers manual sync of both movie and tv genres
func GenreTVShowHandler(w http.ResponseWriter, r *http.Request) {
	go SyncMovieGenres()
	go SyncTvGenres()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sync triggered")
}

