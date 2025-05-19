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
	tmdbRecommendationFilmURL  string
	strapiRecommendationFilmURL string
)


func init () {
  
	tmdbRecommendationFilmURL = "https://api.themoviedb.org/3/movie/"
	strapiRecommendationFilmURL = os.Getenv("STRAPI_URL") + "/api/recommendation-films"


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

func SyncFilmsRecommendation() {

  // Ici on va recupèrer la page de film de strapi 
  lastpage := getLastFetchedPageFilmStrapi(strapiRecommendationFilmURL)
  nextPage := lastpage + 1
  log.Printf("🔄 Sync Film Recommendation : fetching TMDB page %d", nextPage)

  FilmsStrapiPage,err := getFilmsByPageStrapi(nextPage)

  if err != nil {	
	log.Printf("⚠️ Erreur lors de la récupération de la page %d: %v", nextPage, err)
	return
  }

  for _, tmdbID := range FilmsStrapiPage {
	log.Printf("🔄 Synchronisation des recommandations de films : récupération du film TMDB %d", tmdbID)

	var recommendedIDs []int
	page := 1

	for {
		url := fmt.Sprintf("%s%d/recommendations?api_key=%s&language=fr-FR&page=%d",
			tmdbRecommendationFilmURL, tmdbID, os.Getenv("API_KEY"), page)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("⚠️ Erreur lors de la récupération des recommandations (page %d) pour le film %d: %v", page, tmdbID, err)
			break
		}
		defer resp.Body.Close()

		var mr MovieResponse
		if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
			log.Printf("❌ JSON decode TMDB (page %d) pour film %d: %v", page, tmdbID, err)
			break
		}

		// Si aucun résultat sur cette page
		if len(mr.Results) == 0 {
			log.Printf("✅ Fin des recommandations pour le film %d (page %d vide)", tmdbID, page)
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
		log.Printf("ℹ️ Aucune recommandation trouvée pour le film %d", tmdbID)
		continue
	}

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"id_film":           tmdbID,
			"id_films_recommendations": recommendedIDs,
			"page_fetched_from_strapi_film": nextPage,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Erreur encodage JSON pour film %d: %v", tmdbID, err)
		continue
	}

	req, err := http.NewRequest("POST", strapiRecommendationFilmURL, bytes.NewBuffer(body))
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
		log.Printf("⚠️ Strapi a retourné %d pour film %d", res.StatusCode, tmdbID)
	} else {
		log.Printf("✅ Recommandations insérées pour film %d", tmdbID)
	}
}
}

func FilmRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	go SyncFilmsRecommendation()
	w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Synchronisation des recommandations de film TV déclenchée")
}

