// cmd/server/main.go
package main

import (
    "log"
    "mon-projet/internal/handlers"
    "net/http"
    "fmt"
    "os"
)

func main() {
    // Routes HTTP
     // Route racine pour décrire l'API
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        fmt.Fprintln(w, "Bienvenue sur le serveur de Maxil TMDB !")
        fmt.Fprintln(w, "Routes disponibles :")
        fmt.Fprintln(w, "/Genre                  → Récuprèrer les Genres de film ou série TV")
        fmt.Fprintln(w, "/Films                  → Récuprèrer les films")
        fmt.Fprintln(w, "/TvShows                → Récuprèrer les séries TV")
        fmt.Fprintln(w, "/FilmRecommendations    → Récuprèrer les Recommandations de films")
        fmt.Fprintln(w, "/TvShowsRecommendations → Récuprèrer les Recommandations de séries TV")
        fmt.Fprintln(w, "/Configurations         → Récuprèrer la Configuration TMDB")
    })

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
