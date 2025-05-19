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
	tmdbTvShowURL   string
	strapiTvShowURL string
)

type TMDBTvShow struct {
	ID               int      `json:"id"`
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	OriginalName     string   `json:"original_name"`
	OriginalLanguage string   `json:"original_language"`
	Overview         string   `json:"overview"`
	PosterPath       string   `json:"poster_path"`
	ReleaseDate      string   `json:"release_date"`
	Name             string   `json:"name"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	GenreIDs         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
}

type TvShowResponse struct {
	Page       int          `json:"page"`
	Results    []TMDBTvShow `json:"results"`
	TotalPages int          `json:"total_pages"`
}

func init() {
	tmdbTvShowURL = "https://api.themoviedb.org/3/discover/tv"
	strapiTvShowURL = os.Getenv("STRAPI_URL") + "/api/tv-shows"

	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env non chargé: %v", err)
	}

	c := cron.New()
	_, err := c.AddFunc("0 * * * *", func() {
		log.Println("🚀 Lancement planifié: SyncTvShows chaque une heure")
		SyncTvShows()
	})
	if err != nil {
		log.Fatalf("Erreur planification cron: %v", err)
	}
	c.Start()
}

func SyncTvShows() {
	lastPage := getLastFetchedPage(strapiTvShowURL)
	nextPage := lastPage + 1
    log.Printf("🔄 Sync TV shows : récupération de la page %d depuis TMDB", nextPage)

	tmdbURL := fmt.Sprintf("%s?api_key=%s&language=fr-FR&page=%d", tmdbTvShowURL, os.Getenv("API_KEY"), nextPage)
	resp, err := http.Get(tmdbURL)
	if err != nil {
		log.Printf("❌ Erreur TMDB GET: %v", err)
		return
	}
	defer resp.Body.Close()

	var tsr TvShowResponse
	if err := json.NewDecoder(resp.Body).Decode(&tsr); err != nil {
		log.Printf("❌ JSON decode TMDB: %v", err)
		return
	}

	/* J'ai ajouté cette condition pour vérifier si on a atteint la fin de l'API. */
	if len(tsr.Results) == 0 {
		log.Printf("✅ Plus de Tv-show à synchroniser. Toutes les pages TMDB sont terminées.")
		return
	}

	log.Printf("📦 TMDB page %d: %d Tv-Show, total pages %d", tsr.Page, len(tsr.Results), tsr.TotalPages)

	allSuccess := true
	endpoint := strapiTvShowURL + "?filters[id_TvShow][$eq]"

	for _, m := range tsr.Results {
		exists, err := Exists(m.ID, endpoint)
		if err != nil {
           log.Printf("⚠️ Erreur lors de la vérification de l’existence pour l’ID %d : %v", m.ID, err)
			continue
		}
		if exists {
			log.Printf("ℹ️ Tv-Show existant, skip: %s (%d)", m.Name, m.ID)
			continue
		}
			firstAirDate := ""
			if m.FirstAirDate != "" && len(m.FirstAirDate) >= 10 {
			firstAirDate = m.FirstAirDate[:10]
			}

		payload := map[string]interface{}{"data": map[string]interface{}{
			"id_TvShow":            m.ID,
			"Name":                 m.Name,
			"original_Name":        m.OriginalName,
			"original_language":    m.OriginalLanguage,
			"overview":             m.Overview,
			"backdrop_path":        m.BackdropPath,
			"poster_path":          m.PosterPath,
			"Origin_country":       m.OriginCountry,
			"first_air_date":       firstAirDate,
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
		req, _ := http.NewRequest("POST", strapiTvShowURL, bytes.NewBuffer(b))
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
			log.Printf("⚠️ Strapi returned %d for Tv-Show %d", res.StatusCode, m.ID)
			allSuccess = false
		} else {
			log.Printf("✅ Tv-Show inséré: %s (%d)", m.Name, m.ID)
		}


	}

	// Si tous les Tv-Show ont été correctement insérés, on peut dire que la page est traitée
	if !allSuccess {
		log.Printf("⚠️ Tous les Tv Shows de la page %d n'ont pas été insérés. On retentera plus tard.", nextPage)
	} else {
		log.Printf("✅ Tous les Tv Shows de la page %d ont été insérés avec succès.", nextPage)
	}

}


func TvShowHandler(w http.ResponseWriter, r *http.Request) {
	go SyncTvShows()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Synchronisation des séries TV déclenchée")
}
