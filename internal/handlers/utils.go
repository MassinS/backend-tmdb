package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	strapiToken string
)

func init() {
	strapiToken = os.Getenv("STRAPI_TOKEN")
}

func Exists(tmdbID int, endpoint string) (bool, error) {
	url := fmt.Sprintf("%s=%d", endpoint, tmdbID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+strapiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
      log.Printf("⚠️ Erreur lors de la vérification d'existence du film %d : %v", tmdbID, err)
		return false, err
	}
	defer res.Body.Close()

	var data struct {
		Data []interface{} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		log.Printf("⚠️ Erreur lors du décodage de la réponse d'existence : %v", err)
		return false, err
	}
	return len(data.Data) > 0, nil
}


// getLastFetchedPage interroge Strapi pour la plus grande page_fetched_from existante
// la fonction getLastFetchedPage renvoie la dernière page ou le serveur a arrêté de récupérer les films lors de dernier appel
// Alors l'idée ici est que j'ai ajouté pour chaque film un attribut page_fetched_from qui est incrémenté à chaque fois que je fais une requête vers TMDB
// et que je l'enregistre dans Strapi. Donc si je fais une requête vers TMDB et que je récupère 20 films, je vais incrémenter page_fetched_from de 1
// Franchement jai eu cette idée dans le mitro lorsque une veille femme qu'été à coté de moi  a mets un petit papier dans son livre lorsque elle a terminée de lire
// pourqu'elle puisse savoir dans la prochaine lecture où elle s'est arrêté de lire que je me suis inspiré de l'idée page_fetched_from

func getLastFetchedPage(tmdbUrl string) int {

	url := fmt.Sprintf("%s?sort=page_fetched_from:desc&pagination[limit]=1", tmdbUrl)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("⚠️ Erreur création requête pagination: %v", err)
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+strapiToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("⚠️ Erreur requête pagination: %v", err)
		return 0
	}
	defer res.Body.Close()

	var resp struct {
		Data []struct {
			PageFetchedFrom string `json:"page_fetched_from"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		log.Printf("⚠️ Décodage pagination Strapi: %v", err)
		return 0
	}

	if len(resp.Data) == 0 {
		log.Printf("📦 Aucune donnée dans la réponse")
		return 0
	}

	// Convertir la chaîne en int
	page, err := strconv.Atoi(resp.Data[0].PageFetchedFrom)
	if err != nil {
	log.Printf("⚠️ Échec de la conversion de page_fetched_from en entier : %v", err)
		return 0
	}

	log.Printf("📦 Dernière page récupérée: %d", page)
	return page

}

func getLastFetchedPageFilmStrapi(url string) int {
	url = fmt.Sprintf("%s?sort=page_fetched_from_strapi_film:desc&pagination[limit]=1", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("⚠️ Erreur création requête pagination: %v", err)
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+strapiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("⚠️ Erreur requête pagination: %v", err)
		return 0
	}
	defer res.Body.Close()

	var resp struct {
		Data []struct {
			PageFetchedFrom string `json:"page_fetched_from_strapi_film"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		log.Printf("⚠️ Décodage pagination Strapi: %v", err)
		return 0
	}

	if len(resp.Data) == 0 {
		log.Printf("📦 Aucune donnée dans la réponse")
		return 0
	}

	// Convertir la chaîne en int
	page, err := strconv.Atoi(resp.Data[0].PageFetchedFrom)
	if err != nil {
     log.Printf("⚠️ Échec de la conversion de « page_fetched_from » en entier : %v", err)
		return 0
	}

	log.Printf("📦 Dernière page récupérée: %d", page)
	return page

}

func getLastFetchedPageTvShowsStrapi(url string) int {
	url = fmt.Sprintf("%s?sort=page_fetched_from_strapi_TvShow:desc&pagination[limit]=1", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("⚠️ Erreur création requête pagination: %v", err)
		return 0
	}
	req.Header.Set("Authorization", "Bearer "+strapiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("⚠️ Erreur requête pagination: %v", err)
		return 0
	}
	defer res.Body.Close()

	var resp struct {
		Data []struct {
			PageFetchedFrom string `json:"page_fetched_from_strapi_TvShow"`
		} `json:"data"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		log.Printf("⚠️ Décodage pagination Strapi: %v", err)
		return 0
	}

	if len(resp.Data) == 0 {
		log.Printf("📦 Aucune donnée dans la réponse")
		return 0
	}

	// Convertir la chaîne en int
	page, err := strconv.Atoi(resp.Data[0].PageFetchedFrom)
	if err != nil {
    	log.Printf("⚠️ Échec de la conversion de « page_fetched_from » en entier : %v", err)
		return 0
	}

	log.Printf("📦 Dernière page récupérée: %d", page)
	return page

}

type FilmStrapi struct {
	IDFilm int `json:"id_film,string"` // <- string car id_film est une string dans le JSON
}

type ResponseStrapi struct {
	Data []FilmStrapi `json:"data"`
}

func getFilmsByPageStrapi(page int) ([]int, error) {
	strapiFilmURLWithPage := strapiFilmURL + "?filters[page_fetched_from][$eq]=" + strconv.Itoa(page)

	req, err := http.NewRequest("GET", strapiFilmURLWithPage, nil)
	if err != nil {
		return nil, fmt.Errorf("❌ erreur création requête GET: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+strapiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("❌ erreur d'appel HTTP: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("❌ réponse invalide: code %d", res.StatusCode)
	}

	var resp ResponseStrapi
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("❌ erreur décodage JSON: %w", err)
	}

	var filmIDs []int
	for _, film := range resp.Data {
		//log.Printf("ID film: %d", film.IDFilm)
		filmIDs = append(filmIDs, film.IDFilm)
	}

	return filmIDs, nil
}

type TvShowStrapi struct {
	IDFilm int `json:"id_TvShow,string"` // <- string car id_film est une string dans le JSON
}

type ResponseTvShowStrapi struct {
	Data []TvShowStrapi `json:"data"`
}

func getTvShowsByPageStrapi(page int) ([]int, error) {
	strapiFilmURLWithPage := strapiTvShowURL + "?filters[page_fetched_from][$eq]=" + strconv.Itoa(page)

	req, err := http.NewRequest("GET", strapiFilmURLWithPage, nil)
	if err != nil {
		return nil, fmt.Errorf("❌ erreur création requête GET: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+strapiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("❌ erreur appel HTTP: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("❌ réponse invalide: code %d", res.StatusCode)
	}

	var resp ResponseTvShowStrapi
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("❌ erreur décodage JSON: %w", err)
	}

	var filmIDs []int
	for _, film := range resp.Data {
		//log.Printf("ID film: %d", film.IDFilm)
		filmIDs = append(filmIDs, film.IDFilm)
	}

	return filmIDs, nil
}
