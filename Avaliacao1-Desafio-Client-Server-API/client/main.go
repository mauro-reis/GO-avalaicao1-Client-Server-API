package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	//O client.go terá um timeout máximo de 300ms para receber o resultado do server.go:
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	//O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar:
	request, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Println("Erro ao instanciar o request ao servidor:", err)
		fmt.Fprintf(os.Stderr, "Erro ao instanciar o request ao servidor: %v\n", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("Erro ao realizar o request ao servidor:", err)
		panic(err)
	}
	defer response.Body.Close()

	retorno, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Resposta exibida pelo client:", string(retorno))

	arquivo, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer arquivo.Close()

	//O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
	arquivo.WriteString(fmt.Sprintf("Dólar: %v", string(retorno)))
}
