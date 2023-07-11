package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	//O endpoint necessário gerado pelo server.go para este desafio será: /cotacao
	//e a porta a ser utilizada pelo servidor HTTP será a 8080:
	http.HandleFunc("/cotacao", cotacao)
	http.ListenAndServe(":8080", nil)
}

type Cotacoes struct {
	USDBRL USDBRL `gorm:"-"`
}

type USDBRL struct {
	Bid string `json:"bid"`
	gorm.Model
}

func cotacao(w http.ResponseWriter, r *http.Request) {
	//Timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms:
	ctxAPI, cancelAPI := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancelAPI()

	var urlCotacao = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	request, err := http.NewRequestWithContext(ctxAPI, "GET", urlCotacao, nil)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(os.Stderr, "Erro ao realizar o request na API externa: %v\n", err)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Println("Erro ao realizar o request a URL externa:", err)
		panic(err)
	}
	defer response.Body.Close()

	retorno, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println("Erro ao ler o response:", err)
		fmt.Fprintf(os.Stderr, "Erro ao ler o response: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "JSON retornado: ", string(retorno))

	var cotacao Cotacoes

	err = json.Unmarshal(retorno, &cotacao)
	if err != nil {
		log.Println("Erro ao ler o JSON retornado:", err)
		fmt.Fprintf(os.Stderr, "Erro ao ler o JSON retornado: %v\n", err)
	}

	//O client.go precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON):
	w.Write([]byte(cotacao.USDBRL.Bid))

	db, err := gorm.Open(sqlite.Open("db_cotacao.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&USDBRL{Bid: cotacao.USDBRL.Bid})

	db_cotacao := USDBRL{Bid: cotacao.USDBRL.Bid}

	//Timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms:
	ctxDB, cancelDB := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancelDB()

	result := db.WithContext(ctxDB).Create(&db_cotacao)

	if result.Error != nil {
		log.Println(result.Error)
		fmt.Fprintf(os.Stderr, "Erro ao gravar no banco de dados: %v\n", result.Error)
	}

	fmt.Println("Info do BD:", cotacao.USDBRL.Bid)
}
