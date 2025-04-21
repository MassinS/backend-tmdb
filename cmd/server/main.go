// cmd/server/main.go
package main

import (
	"log"
	"mon-projet/internal/handlers" // Assurez-vous que le chemin d'import correspond à votre structure de projet
	"net/http"
)

func main() {
	// Associer le handler Accueil à la route "/"
	http.HandleFunc("/Genre", handlers.GenreTVShowHandler)

	log.Println("Serveur démarré sur le port 8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Erreur lors du démarrage du serveur: %v", err)
	}

}
