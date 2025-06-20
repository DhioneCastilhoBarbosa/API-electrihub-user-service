package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type InstallerData struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	CPF          string `json:"cpf"`
	CNPJ         string `json:"cnpj"`
	CompanyName  string `json:"company_name"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Complement   string `json:"complement"`
	CEP          string `json:"cep"`
	BirthDate    string `json:"birth_date"`
	Reference    string `json:"reference"`
}

func NotifyNewInstaller(data InstallerData) error {
	url := "https://mail.api-castilho.com.br/send-email-new-installer"
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("erro ao serializar dados para JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("erro ao enviar Post: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("erro na resposta da API de e-mail: %s", resp.Status)
	}
	return nil
}
