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
	tmdbRecommendationTvShowsURL   string
	strapiRecommendationTvShowsURL string
)

func init() {

	tmdbRecommendationTvShowsURL = "https://api.themoviedb.org/3/tv/"
	strapiRecommendationTvShowsURL = os.Getenv("STRAPI_URL") + "/api/recommendation-tv-shows"

	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env non chargé: %v", err)
	}

	c := cron.New()
	_, err := c.AddFunc("0 0 * * *", func() {
		log.Println("🚀 Lancement planifié: SyncMovies chaque 24h")
		SyncMovies()
	})
	if err != nil {
		log.Fatalf("Erreur cron SyncMovies: %v", err)
	}
	c.Start()

}

func SyncTvShowsRecommendation() {


	// Ici on va recupèrer la page de film de strapi
	lastpage := getLastFetchedPageTvShowsStrapi(strapiRecommendationTvShowsURL)
	nextPage := lastpage + 1
    log.Printf("🔄 Synchronisation des recommandations de séries TV : récupération de la page %d depuis TMDB", nextPage)

	// Je suis arrivé ici , on doit ajouter une foncion getTvShowsPageStrapi
	TvShowsStrapiPage, err := getTvShowsByPageStrapi(nextPage)

	if err != nil {
		log.Printf("⚠️ Erreur lors de la récupération de la page %d: %v", nextPage, err)
		return
	}

	for _, tmdbID := range TvShowsStrapiPage {
      log.Printf("🔄 Sync TV shows recommendation : récupération de la page %d depuis TMDB", nextPage)

		var recommendedIDs []int
		page := 1

		for {
			url := fmt.Sprintf("%s%d/recommendations?api_key=%s&language=fr-FR&page=%d",
				tmdbRecommendationTvShowsURL, tmdbID, os.Getenv("API_KEY"), page)

			resp, err := http.Get(url)
			if err != nil {
				log.Printf("⚠️ Erreur lors de la récupération des recommandations (page %d) pour le Tv Show %d: %v", page, tmdbID, err)
				break
			}
			defer resp.Body.Close()

			var mr TvShowResponse
			if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
				log.Printf("❌ Erreur de décodage JSON depuis TMDB (page %d) pour la série TV %d : %v", page, tmdbID, err)
				break
			}

			// Si aucun résultat sur cette page
			if len(mr.Results) == 0 {
				log.Printf("✅ Fin des recommandations pour le Tv-Shows %d (page %d vide)", tmdbID, page)
				break
			}

			for _, rec := range mr.Results {
				recommendedIDs = append(recommendedIDs, rec.ID)
			}

			if page >= mr.TotalPages {
				break
			}
			page++
		}

		if len(recommendedIDs) == 0 {
			log.Printf("ℹ️ Aucune recommandation trouvée pour le Tv Show %d", tmdbID)
			continue
		}

		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id_TvShow":                       tmdbID,
				"id_TvShow_recommendations":       recommendedIDs,
				"page_fetched_from_strapi_TvShow": nextPage,
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("❌ Erreur encodage JSON pour Tv Show %d: %v", tmdbID, err)
			continue
		}

		req, err := http.NewRequest("POST", strapiRecommendationTvShowsURL, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("❌ Erreur création requête POST Strapi: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+os.Getenv("STRAPI_TOKEN"))

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("❌ Erreur envoi POST à Strapi: %v", err)
			continue
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			log.Printf("⚠️ Strapi a retourné %d pour Tv Show %d", res.StatusCode, tmdbID)
		} else {
			log.Printf("✅ Recommandations insérées pour film %d", tmdbID)
		}
	}

}

func TvShowRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	go SyncTvShowsRecommendation()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Sync Tv-show recommendations triggered")
}
