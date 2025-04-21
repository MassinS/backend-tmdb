package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"bytes"
	"time"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var (
	strapiURL = "https://tmdb-database-strapi.onrender.com/api/genre-tv-shows"
)

type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type GenreResponse struct {
	Genres []TMDBGenre `json:"genres"`
}

func init() {
	// Initialisation du scheduler
	c := cron.New()
	
	// Planification à minuit chaque jour
	_, err := c.AddFunc("11 0 * * *", func() {
		log.Println("🚀 Démarrage de la tâche planifiée...")
		SyncGenres()
	})
	
	if err != nil {
		log.Fatalf("Erreur de planification : %v", err)
	}
	
	c.Start()
	
	// Chargement des variables d'environnement
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Erreur .env : %v", err)
	}
}

// SyncGenres contient la logique principale à exécuter
func SyncGenres() {
	startTime := time.Now()
	
	baseURL := os.Getenv("BASE_URL")
	apiKey := os.Getenv("API_KEY")
	
	// Récupération des données TMDB
	url := fmt.Sprintf("%s/genre/movie/list?api_key=%s", baseURL, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ Erreur de requête TMDB : %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("⚠️ Code TMDB inattendu : %d", resp.StatusCode)
		return
	}

	var tmdbGenres GenreResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbGenres); err != nil {
		log.Printf("❌ Erreur décodage TMDB : %v", err)
		return
	}

	// Envoi vers Strapi
	for _, genre := range tmdbGenres.Genres {
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"id_genre": genre.ID,
				"nom_genre": genre.Name,
			
			},
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", strapiURL, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("❌ Erreur envoi %s: %v", genre.Name, err)
			continue
		}
		defer res.Body.Close()

		if res.StatusCode >= 400 {
			log.Printf("⚠️ Échec envoi %s: %d", genre.Name, res.StatusCode)
		} else {
			log.Printf("✅ Succès envoi %s", genre.Name)
		}
	}

	log.Printf("🎉 Synchronisation terminée en %s", time.Since(startTime))
}

// Handler HTTP (optionnel)
func GenreTVShowHandler(w http.ResponseWriter, r *http.Request) {
	SyncGenres()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Synchronisation déclenchée manuellement")
}


