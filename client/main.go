package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	cotacao := map[string]string{}
	err = json.NewDecoder(res.Body).Decode(&cotacao)
	if err != nil {
		panic(err)
	}
	f, err := os.Create("cotacao.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write([]byte(fmt.Sprintf("Dolar:{%s}", cotacao["bid"])))
	if err != nil {
		panic(err)
	}
	fmt.Println("Cotação gravada no arquivo com sucesso")

}
