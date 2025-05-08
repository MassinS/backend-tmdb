// cmd/server/main.go
package main

import (
	"log"
	"mon-projet/internal/handlers" // Assurez-vous que le chemin d'import correspond à votre structure de projet
	"net/http"
)

func main() {
	// Associer le handler Accueil à la route "/Genre"
	http.HandleFunc("/Genre", handlers.GenreTVShowHandler)
	http.HandleFunc("/Films", handlers.MovieHandler)
	http.HandleFunc("/TvShows", handlers.TvShowHandler)
	http.HandleFunc("/FilmRecommendations", handlers.FilmRecommendationHandler)
	http.HandleFunc("/TvShowsRecommendations", handlers.TvShowRecommendationHandler)
	http.HandleFunc("/Configurations", handlers.ConfigurationHandler)

	log.Println("Serveur démarré sur le port 8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}

}
