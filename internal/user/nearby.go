package user

import (
	"math"
	"net/http"
	"strconv"
	"user-service/internal/database"
	"user-service/internal/user/models"

	"github.com/gin-gonic/gin"
)

func calcularDistanciaKm(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func ListNearbyInstallers(c *gin.Context) {
	lat := c.Query("lat")
	lng := c.Query("lng")

	latF, err1 := strconv.ParseFloat(lat, 64)
	lngF, err2 := strconv.ParseFloat(lng, 64)

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetros latitude e longitude são obrigatórios e válidos"})
		return
	}

	var users []models.User
	if err := database.DB.Where("role = ?", "instalador").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar instaladores"})
		return
	}

	const limiteKm = 150.0
	var proximos []UserInstalerResponse

	for _, user := range users {
		if user.Latitude == 0 && user.Longitude == 0 {
			continue
		}
		dist := calcularDistanciaKm(latF, lngF, user.Latitude, user.Longitude)
		if dist <= limiteKm {
			proximos = append(proximos, UserInstalerResponse{
				ID:                    user.ID,
				Name:                  user.Name,
				CompanyName:           user.CompanyName,
				AverageRating:         user.AverageRating,
				TotalServicesAccepted: user.TotalServicesAccepted,
				ServicesNotExecuted:   user.ServicesNotExecuted,
				Role:                  user.Role,
				Photo:                 user.Photo,
			})
		}
	}

	c.JSON(http.StatusOK, proximos)
}
