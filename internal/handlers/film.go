package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Film représente la structure d'un film renvoyé par l'API
type Film struct {
	ID          int    `json:"id"`
	DocumentID  string `json:"documentId"`
	Name        string `json:"name"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	PublishedAt string `json:"publishedAt"`
}

// APIResponse représente la réponse complète de l'API
type APIResponse struct {
	Data []Film `json:"data"`
	Meta struct {
		Pagination struct {
			Page      int `json:"page"`
			PageSize  int `json:"pageSize"`
			PageCount int `json:"pageCount"`
			Total     int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
}

// FilmsHandler gère la requête pour récupérer les films depuis l'API externe
func FilmsHandler(w http.ResponseWriter, r *http.Request) {
	
	url := "https://tmdb-database-strapi.onrender.com/api/films"

	// Ici j'envoie une requete HTTP GET  
	resp, err := http.Get(url)
	// err au cas où il y'a avais un erreur lors de traitement de la requete
	// resp est la réponse de la requete GET 

	
	if err != nil {
		http.Error(w, "Erreur lors de la requête GET", http.StatusInternalServerError)
		return
	}


	defer resp.Body.Close()

	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du corps de réponse", http.StatusInternalServerError)
		return
	}

	// Parser correctement le JSON
	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		http.Error(w, "Erreur lors du parsing JSON", http.StatusInternalServerError)
		return
	}

	// Afficher les films pour debug
	for _, film := range apiResponse.Data {
		fmt.Printf("ID: %d, Nom: %s, DocumentID: %s\n", film.ID, film.Name, film.DocumentID)
	}

	// Convertir en JSON et envoyer comme réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiResponse.Data)
}
