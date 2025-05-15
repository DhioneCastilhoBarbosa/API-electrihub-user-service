package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func BuscarCoordenadas(enderecoCompleto string) (float64, float64, error) {
	apiKey := os.Getenv("LOCATIONIQ_API_KEY")
	baseURL := "https://us1.locationiq.com/v1/search.php"

	query := fmt.Sprintf("%s?q=%s&key=%s&country=Brazil&format=json",
		baseURL,
		url.QueryEscape(enderecoCompleto),
		apiKey,
	)

	resp, err := http.Get(query)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}
	if len(results) == 0 {
		return 0, 0, fmt.Errorf("Endereço não encontrado")
	}

	lat, _ := strconv.ParseFloat(results[0].Lat, 64)
	lng, _ := strconv.ParseFloat(results[0].Lon, 64)
	return lat, lng, nil
}
