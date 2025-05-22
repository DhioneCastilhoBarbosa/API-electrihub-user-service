package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func BuscarCoordenadas(enderecoCompleto string) (float64, float64, error) {
	apiKey := os.Getenv("LOCATIONIQ_API_KEY")
	if apiKey == "" {
		return 0, 0, fmt.Errorf("API key do LocationIQ não encontrada")
	}

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

	// Lê todo o corpo da resposta
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	// Tenta decodificar como array
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}
	if err := json.Unmarshal(bodyBytes, &results); err == nil && len(results) > 0 {
		lat, _ := strconv.ParseFloat(results[0].Lat, 64)
		lng, _ := strconv.ParseFloat(results[0].Lon, 64)
		return lat, lng, nil
	}

	// Tenta decodificar como erro
	var errorResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Error != "" {
		return 0, 0, fmt.Errorf("Erro do LocationIQ: %s", errorResp.Error)
	}

	// Caso não consiga interpretar a resposta de nenhuma forma
	return 0, 0, fmt.Errorf("Resposta inesperada do serviço de geolocalização")
}
