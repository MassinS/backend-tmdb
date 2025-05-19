# backend-tmdb

## Pour démarrer le serveur, exécuter le script :
Nous avons ajouté le fichier .env afin de faciliter le démarrage du serveur. Cependant, dans les projets réels, on ne partage jamais ce fichier afin de protéger les clés sensibles.

./start.sh

## Structure du projet :

<pre> ``` 
backend-tmdb/ 
|
├── cmd/
   │ └── server/ 
   │ └── main.go 
├── internal/ 
   │ └── handlers/ 
   │ ├── ConfigurationTMDB.go 
   │ ├── Genre.go 
   │ ├── Movie.go 
   │ ├── RecommendationFilms.go 
   │ ├── RecommendationTvShows.go 
   │ ├── TvShow.go 
   │ └── utils.go 
├── .env 
├── go.mod 
├── go.sum 
├── README.md 
``` </pre>
