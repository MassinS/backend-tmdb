package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var (
	tmdbMovieURL  string
	strapiFilmURL string
)

// TMDBMovie représente un film renvoyé par TMDB
// GenreIDs est un slice d'entiers car TMDB renvoie [id1, id2, ...]
type TMDBMovie struct {
	ID               int     `json:"id"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	OriginalTitle    string  `json:"original_title"`
	OriginalLanguage string  `json:"original_language"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Popularity       float64 `json:"popularity"`
	GenreIDs         []int   `json:"genre_ids"`
}

// MovieResponse enveloppe la réponse TMDB pour discover/movie
type MovieResponse struct {
	Page       int         `json:"page"`
	Results    []TMDBMovie `json:"results"`
	TotalPages int         `json:"total_pages"`
}

func init() {

	tmdbMovieURL = "https://api.themoviedb.org/3/discover/movie"
	strapiFilmURL = os.Getenv("STRAPI_URL") + "/api/films"

	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env non chargé: %v", err)
	}

	c := cron.New()
	_, err := c.AddFunc("0 * * * *", func() {
		log.Println("🚀 Lancement planifié: SyncMovies")
		SyncMovies()
	})
	if err != nil {
		log.Fatalf("Erreur cron SyncMovies: %v", err)
	}
	c.Start()

}

// la fonction SyncMovies est la responsable de recuperer les données
// vérifier si les données existe pas dans la base de données
// recueprer pour chaque film les genres qui le correspond
// et enfin les stocker dans la table films
func SyncMovies() {
	lastPage := getLastFetchedPage(strapiFilmURL)
	nextPage := lastPage + 1
	log.Printf("🔄 SyncMovies: fetching TMDB page %d", nextPage)

	tmdbURL := fmt.Sprintf("%s?api_key=%s&language=fr-FR&page=%d", tmdbMovieURL, os.Getenv("API_KEY"), nextPage)
	resp, err := http.Get(tmdbURL)
	if err != nil {
		log.Printf("❌ Erreur TMDB GET: %v", err)
		return
	}
	defer resp.Body.Close()

	var mr MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		log.Printf("❌ JSON decode TMDB: %v", err)
		return
	}

	/* J'ai ajouté cette condition pour vérifier si on a atteint la fin de l'API. */
	if len(mr.Results) == 0 {
		log.Printf("✅ Plus de films à synchroniser. Toutes les pages TMDB sont terminées.")
		return
	}

	log.Printf("📦 TMDB page %d: %d films, total pages %d", mr.Page, len(mr.Results), mr.TotalPages)

	allSuccess := true
	endpoint := strapiFilmURL + "?filters[id_film][$eq]"

	for _, m := range mr.Results {
		exists, err := Exists(m.ID, endpoint)
		if err != nil {
			log.Printf("⚠️ check exists error for %d: %v", m.ID, err)
			continue
		}
		if exists {
			log.Printf("ℹ️ Film existant, skip: %s (%d)", m.Title, m.ID)
			continue
		}

		payload := map[string]interface{}{"data": map[string]interface{}{
			"id_film":              m.ID,
			"title":                m.Title,
			"original_title":       m.OriginalTitle,
			"original_language":    m.OriginalLanguage,
			"overview":             m.Overview,
			"Backdrop_path":        m.BackdropPath,
			"poster_path":          m.PosterPath,
			"release_date":         m.ReleaseDate,
			"Video":                m.Video,
			"vote_average_tmdb":    m.VoteAverage,
			"vote_count_tmdb":      m.VoteCount,
			"popularity_tmdb":      m.Popularity,
			"genre_tv_films":       m.GenreIDs,
			"adult":                m.Adult,
			"popularity_website":   0.0,
			"vote_average_website": 0.0,
			"vote_count_website":   0.0,
			"page_fetched_from":    nextPage,
		}}

		b, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", strapiFilmURL, bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+strapiToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("❌ POST Strapi film %d: %v", m.ID, err)
			allSuccess = false
			continue
		}
		res.Body.Close()

		if res.StatusCode >= 400 {
			log.Printf("⚠️ Strapi returned %d for film %d", res.StatusCode, m.ID)
			allSuccess = false
		} else {
			log.Printf("✅ Film inséré: %s (%d)", m.Title, m.ID)
		}
	}

	// Si tous les films ont été correctement insérés, on peut dire que la page est traitée
	if !allSuccess {
		log.Printf("⚠️ Tous les films de la page %d n'ont pas été insérés. On retentera plus tard.", nextPage)
	} else {
		log.Printf("✅ Tous les films de la page %d ont été insérés avec succès.", nextPage)
	}

}

// MovieHandler déclenche manuellement la sync
func MovieHandler(w http.ResponseWriter, r *http.Request) {
	go SyncMovies()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "SyncMovies triggered")
}
