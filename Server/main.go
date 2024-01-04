package main

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	db      *gorm.DB
	errorDB error
)

func migrateDatabase() {
	db, errorDB = gorm.Open(sqlite.Open("mydb.db"), &gorm.Config{})
	if errorDB != nil {
		log.Fatal("Erro ao abrir o banco de dados:", errorDB)
	}

	if err := db.AutoMigrate(&Cotacao{}); err != nil {
		log.Fatal("Erro ao migrar o banco de dados:", err)
	}
}

func main() {
	migrateDatabase() // Realiza a migração do banco de dados antes de iniciar o servidor

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", BuscaCotacaoHandler)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response, err := BuscaCotacao()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		message := map[string]string{
			"errorMessage": err.Error(),
		}
		json.NewEncoder(w).Encode(message)
		return
	}

	err = saveCotacao(&response.USDBRL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		message := map[string]string{
			"errorMessage": err.Error(),
		}
		json.NewEncoder(w).Encode(message)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}

func BuscaCotacao() (*ResponseCotacao, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelFunc()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err.Error())
		return nil, errors.New("Falha ao buscar a cotação, tempo excedido")
	}
	defer resp.Body.Close()

	resultBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response ResponseCotacao
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func saveCotacao(data *USDBRL) error {
	cotacao := Cotacao{
		Bid:  data.Bid,
		Data: time.Now(),
	}

	// Utilização de um contexto com timeout de 10ms para a operação de salvamento
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancelFunc()

	startTime := time.Now()
	// Aplicação do contexto na operação de banco de dados
	db = db.WithContext(ctx).
		Model(&Cotacao{}).
		Debug().
		Create(&cotacao)

	duration := time.Since(startTime)
	log.Println("Tempo de execução:", duration)

	if db.Error != nil {
		log.Println(db.Error)
		return db.Error
	}

	return nil

}

type Cotacao struct {
	ID   int `gorm:"primaryKey"`
	Data time.Time
	Bid  string
}

type ResponseCotacao struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type USDBRL struct {
	Bid string `json:"bid"`
}
