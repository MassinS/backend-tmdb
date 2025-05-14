// cmd/server/main.go
package main

import (
    "log"
    "mon-projet/internal/handlers"
    "net/http"
    "os"
)

func main() {
    // Routes HTTP
    http.HandleFunc("/Genre", handlers.GenreTVShowHandler)
    http.HandleFunc("/Films", handlers.MovieHandler)
    http.HandleFunc("/TvShows", handlers.TvShowHandler)
    http.HandleFunc("/FilmRecommendations", handlers.FilmRecommendationHandler)
    http.HandleFunc("/TvShowsRecommendations", handlers.TvShowRecommendationHandler)
    http.HandleFunc("/Configurations", handlers.ConfigurationHandler)

    // Port dynamique (Render injecte la variable $PORT)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8081" // fallback local
    }

    log.Printf("Serveur démarré sur le port %s…", port)
    if err := http.ListenAndServe(":" + port, nil); err != nil {
        log.Fatalf("Erreur au démarrage du serveur: %v", err)
    }
}
