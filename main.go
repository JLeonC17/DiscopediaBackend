package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Album struct {
	ArtistaTitulo string `json:"title"`
	Año           string `json:"year"`
	Image         string `json:"cover_image"`
	MasterId      int    `json:"master_id"`
}

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5500"}
	router.Use(cors.New(config))

	router.GET("/api/discogs", discogsController)

	router.Run("localhost:8080")
}

func goDotEnvVariable(key string) string {

	err := godotenv.Load(".env")

	if err != nil {
		fmt.Printf("Error loading .env file")
	}

	return os.Getenv(key)
}

func filterAlbums(resultsArray []interface{}) []interface{} {
	filteredResults := make([]interface{}, 0, len(resultsArray))

	for _, element := range resultsArray {
		if resultData, ok := element.(map[string]interface{}); ok {

			formatArray, formatArrayExists := resultData["format"].([]interface{})
			formatStrings := make([]string, 0)
			if formatArrayExists {
				for _, formatElement := range formatArray {
					if formatStr, isString := formatElement.(string); isString {
						formatStrings = append(formatStrings, formatStr)
					}
				}
			}

			genreArray, genreArrayExists := resultData["genre"].([]interface{})
			genreStrings := make([]string, 0)
			if genreArrayExists {
				for _, genreElement := range genreArray {
					if genreStr, isString := genreElement.(string); isString {
						genreStrings = append(genreStrings, genreStr)
					}
				}
			}

			if !strings.Contains(strings.Join(formatStrings, ""), "Unofficial Release") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Box Set") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Reissue") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Compilation") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Limited Edition") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Remastered") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Mixed") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Cassette") &&
				!strings.Contains(strings.Join(formatStrings, ""), "White Label") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Promo") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Single") &&
				!strings.Contains(strings.Join(formatStrings, ""), "Maxi-Single") &&
				(strings.Contains(strings.Join(formatStrings, ""), "Album") ||
					strings.Contains(strings.Join(formatStrings, ""), "LP") ||
					strings.Contains(strings.Join(formatStrings, ""), "EP") ||
					strings.Contains(strings.Join(genreStrings, ""), "Stage & Screen")) &&
				!strings.Contains(strings.Join(genreStrings, ""), "Non-Music") {

				filteredResults = append(filteredResults, element)
			}
		}
	}
	cleanedResults := make([]interface{}, 0, len(filteredResults))
	for _, element := range filteredResults {
		if _, ok := element.(map[string]interface{}); ok {
			cleanedResults = append(cleanedResults, element)
		}
	}

	return cleanedResults
}

func getStringValue(data interface{}) string {
	if str, ok := data.(string); ok {
		return str
	}
	return ""
}

func processAlbums(resultsArray []interface{}) []Album {

	albums := make([]Album, 0)

	for _, result := range resultsArray {

		jsonResults, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			fmt.Printf("Error al convertir a JSON: %v\n", err)
		}

		fmt.Println(string(jsonResults))

		if resultData, ok := result.(map[string]interface{}); ok && result != nil {
			title := getStringValue(resultData["title"])
			year := getStringValue(resultData["year"])
			coverImage := getStringValue(resultData["cover_image"])
			masterID := int(resultData["master_id"].(float64))
			{
				album := Album{
					ArtistaTitulo: title,
					Año:           year,
					Image:         coverImage,
					MasterId:      masterID,
				}

				albums = append(albums, album)
			}
		}
	}

	return albums
}

func fetchGetData(url string, resultsArray *[]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creando la solicitud: %v\n", err)
		return
	}

	key := goDotEnvVariable("DISCOGS_KEY")
	secret := goDotEnvVariable("DISCOGS_SECRET")
	headerName := "Authorization"
	headerValue := fmt.Sprintf("Discogs key=%s, secret=%s", key, secret)

	req.Header.Add(headerName, headerValue)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error obteniendo datos de %s: %v\n", url, err)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error leyendo el cuerpo de la respuesta: %v\n", err)
		return
	}

	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("Error decodificando JSON: %v\n", err)
		return
	}

	if results, ok := responseData["results"].([]interface{}); ok && len(results) > 0 {
		mutex := &sync.Mutex{}
		mutex.Lock()
		defer mutex.Unlock()

		for _, result := range results {
			if result != nil {
				*resultsArray = append(*resultsArray, result)
			}
		}

	}

}

func getDiscogsData(titulo, artista string, resultsArray *[]interface{}) []Album {
	var wg sync.WaitGroup

	formatos := []string{"album", "ep", "mini-album", ""}

	for _, formato := range formatos {
		wg.Add(1)
		go fetchGetData(fmt.Sprintf("https://api.discogs.com/database/search?release_title=%s&artist=%s&type=master&format=%s&per_page=100", titulo, artista, formato), resultsArray, &wg)
	}

	wg.Wait()

	return processAlbums(filterAlbums(*resultsArray))
}

func discogsController(c *gin.Context) {

	err := godotenv.Load()

	if err != nil {
		fmt.Printf("Error cargando el archivo .env")
		return
	}

	artista := c.Query("artista")
	titulo := c.Query("titulo")

	var albumsArray []interface{}

	fmt.Printf("Titutlo: %s \nArtista: %s \n", titulo, artista)

	c.IndentedJSON(http.StatusOK, getDiscogsData(titulo, artista, &albumsArray))

}
