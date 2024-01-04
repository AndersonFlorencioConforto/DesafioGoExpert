package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	err := BuscarCotacao()
	if err != nil {
		log.Println(err)
	}

}

func gravarCotacao(response *ResponseCotacao) error {
	file, err := os.OpenFile("cotacao.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	defer file.Close()
	if err != nil {
		return err
	}

	_, err = file.WriteString(fmt.Sprintf("Dólar: %s\n", response.USDBRL.Bid))
	if err != nil {
		return err
	}
	return nil

}

func BuscarCotacao() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancelFunc()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println(err.Error())
		return errors.New("Falha ao buscar a cotação, tempo excedido")
	}
	defer resp.Body.Close()

	resultBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response ResponseCotacao
	err = json.Unmarshal(resultBody, &response)
	if err != nil {
		return err
	}
	err = gravarCotacao(&response)

	if err != nil {
		return err
	}
	return nil
}

type ResponseCotacao struct {
	USDBRL USDBRL `json:"USDBRL"`
}

type USDBRL struct {
	Bid string `json:"bid"`
}
