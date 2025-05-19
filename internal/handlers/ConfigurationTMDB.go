package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/joho/godotenv"
	cron "github.com/robfig/cron/v3"
)

var (
	tmdbConfigurationURL   string
	strapiConfigurationURL string
)

type TmdbImageConfig struct {
	BaseURL       string   `json:"base_url"`
	SecureBaseURL string   `json:"secure_base_url"`
	BackdropSizes []string `json:"backdrop_sizes"`
	LogoSizes     []string `json:"logo_sizes"`
	PosterSizes   []string `json:"poster_sizes"`
	ProfileSizes  []string `json:"profile_sizes"`
	StillSizes    []string `json:"still_sizes"`
}

type TmdbConfigResponse struct {
	Images     TmdbImageConfig `json:"images"`
	ChangeKeys []string        `json:"change_keys"`
}

func init() {
	// Charge .env avant d'utiliser os.Getenv
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️ .env non chargé: %v", err)
	}

	tmdbConfigurationURL = "https://api.themoviedb.org/3/configuration"
	strapiConfigurationURL = os.Getenv("STRAPI_URL") + "/api/configurations"

	c := cron.New()
	_, err := c.AddFunc("0 0 * * 0", func() {
		log.Println("🚀 Lancement planifié: SyncConfiguration chaque une semaine ")
		SyncConfiguration()
	})
	if err != nil {
		log.Fatalf("Erreur planification cron: %v", err)
	}
	c.Start()
}

func SyncConfiguration() {
	// Étape 1: récupère la config TMDB
	tmdbURL := fmt.Sprintf("%s?api_key=%s", tmdbConfigurationURL, os.Getenv("API_KEY"))
	tmdbRespRaw, err := http.Get(tmdbURL)
	if err != nil {
		log.Printf("⚠️ Erreur fetch configuration TMDB: %v", err)
		return
	}
	defer tmdbRespRaw.Body.Close()

	var tmdbResp TmdbConfigResponse
	if err := json.NewDecoder(tmdbRespRaw.Body).Decode(&tmdbResp); err != nil {
		log.Printf("⚠️ Erreur décodage JSON TMDB: %v", err)
		return
	}

	// Étape 2: récupère la config Strapi
	req, err := http.NewRequest("GET", strapiConfigurationURL, nil)
	if err != nil {
		log.Printf("⚠️ Erreur création requête GET Strapi: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+os.Getenv("STRAPI_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("⚠️ Erreur récupération configuration Strapi: %v", err)
		return
	}
	defer resp.Body.Close()

	// Vérifie le status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("⚠️ GET config Strapi returned %d: %s", resp.StatusCode, string(body))
		return
	}

	// Décodage avec les champs à la racine de chaque data[]
	var strapiResponse struct {
		Data []struct {
			ID            string   `json:"documentId"`
			BaseURL       string   `json:"base_url"`
			SecureBaseURL string   `json:"secure_base_url"`
			BackdropSizes []string `json:"backdrop_sizes"`
			LogoSizes     []string `json:"logo_sizes"`
			PosterSizes   []string `json:"poster_sizes"`
			ProfileSizes  []string `json:"profile_sizes"`
			StillSizes    []string `json:"still_sizes"`
			ChangeKeys    []string `json:"change_keys"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&strapiResponse); err != nil {
		log.Printf("⚠️ Erreur décodage configuration Strapi: %v", err)
		return
	}

	// Si aucune entrée, on POST
	if len(strapiResponse.Data) == 0 {
		log.Println(" Aucune configuration trouvée, création via POST")
		payload := map[string]interface{}{
			"data": map[string]interface{}{
				"base_url":        tmdbResp.Images.BaseURL,
				"secure_base_url": tmdbResp.Images.SecureBaseURL,
				"backdrop_sizes":  tmdbResp.Images.BackdropSizes,
				"logo_sizes":      tmdbResp.Images.LogoSizes,
				"poster_sizes":    tmdbResp.Images.PosterSizes,
				"profile_sizes":   tmdbResp.Images.ProfileSizes,
				"still_sizes":     tmdbResp.Images.StillSizes,
				"change_keys":     tmdbResp.ChangeKeys,
			},
		}
		body, _ := json.Marshal(payload)
		postReq, _ := http.NewRequest("POST", strapiConfigurationURL, bytes.NewReader(body))
		postReq.Header.Set("Content-Type", "application/json")
		postReq.Header.Set("Authorization", "Bearer "+os.Getenv("STRAPI_TOKEN"))

		postRes, err := http.DefaultClient.Do(postReq)
		if err != nil {
			log.Printf("⚠️ Erreur exécution POST: %v", err)
			return
		}
		defer postRes.Body.Close()
		if postRes.StatusCode >= 200 && postRes.StatusCode < 300 {
			log.Println("✅ Configuration créée avec succès via POST")
		} else {
			log.Printf("⚠️ POST échoué - Code: %d", postRes.StatusCode)
		}
		return
	}

	// Reçoit la première entrée existante
	entry := strapiResponse.Data[0]
	strapiID := entry.ID
	strapiConfig := TmdbImageConfig{
		BaseURL:       entry.BaseURL,
		SecureBaseURL: entry.SecureBaseURL,
		BackdropSizes: entry.BackdropSizes,
		LogoSizes:     entry.LogoSizes,
		PosterSizes:   entry.PosterSizes,
		ProfileSizes:  entry.ProfileSizes,
		StillSizes:    entry.StillSizes,
	}
	// et change_keys à part
	strapiChangeKeys := entry.ChangeKeys

	// Log des deux JSON pour debug
	strapiJSON, _ := json.MarshalIndent(strapiConfig, "", "  ")
	tmdbJSON, _ := json.MarshalIndent(tmdbResp.Images, "", "  ")
	log.Printf("🔍 strapiConfig: %s", string(strapiJSON))
	log.Printf("🔍 tmdbResp.Images: %s", string(tmdbJSON))

	// Étape 3: comparer changement
	if reflect.DeepEqual(strapiConfig, tmdbResp.Images) && reflect.DeepEqual(strapiChangeKeys, tmdbResp.ChangeKeys) {
		log.Println("✅ Configuration TMDB inchangée")
		return
	}

	log.Println("⚠️ Différence détectée, on va mettre à jour…")

	// Étape 4: Construction du payload pour PUT
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"base_url":        tmdbResp.Images.BaseURL,
			"secure_base_url": tmdbResp.Images.SecureBaseURL,
			"backdrop_sizes":  tmdbResp.Images.BackdropSizes,
			"logo_sizes":      tmdbResp.Images.LogoSizes,
			"poster_sizes":    tmdbResp.Images.PosterSizes,
			"profile_sizes":   tmdbResp.Images.ProfileSizes,
			"still_sizes":     tmdbResp.Images.StillSizes,
			"change_keys":     tmdbResp.ChangeKeys,
		},
	}
	configJSON, _ := json.Marshal(payload)

	log.Printf("➡️ Tentative PUT sur %s/%s", strapiConfigurationURL, strapiID)
	putReq, _ := http.NewRequest("PUT", fmt.Sprintf("%s/%s", strapiConfigurationURL, strapiID), bytes.NewReader(configJSON))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+os.Getenv("STRAPI_TOKEN"))

	putRes, err := http.DefaultClient.Do(putReq)
	if err != nil {
		log.Printf("⚠️ Erreur exécution PUT: %v", err)
		return
	}
	defer putRes.Body.Close()

	if putRes.StatusCode >= 200 && putRes.StatusCode < 300 {
		log.Println("🔄 Configuration mise à jour avec succès")
	} else {
		log.Printf("⚠️ PUT échoué - Code: %d", putRes.StatusCode)
	}

}

func ConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	go SyncConfiguration()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Synchronisation de la configuration déclenchée")
}
