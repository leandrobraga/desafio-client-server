package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	ID         string `json:"id"`
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", CotacaoHandler)
	http.ListenAndServe(":8080", mux)
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		http.Error(w, "Request canceled by client.", http.StatusRequestTimeout)
		fmt.Println("Request canceled by client")
		return
	case <-time.After(300 * time.Millisecond):
		http.Error(w, "Request timeout.", http.StatusRequestTimeout)
		fmt.Println("Request timeout")
		return
	default:
		ctxCotacao, cancelCotacao := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancelCotacao()
		cotacao, err := BuscaCotacao(ctxCotacao)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		db, err := sql.Open("sqlite3", "desafio.db")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("Error happened in OPen sqlite. Err: %s", err)
			return
		}
		defer db.Close()

		// 	const create string = `
		// CREATE TABLE IF NOT EXISTS cotacao (
		// 	id TEXT NOT NULL PRIMARY KEY,
		// 	code TEXT,
		// 	codein TEXT,
		// 	name TEXT,
		// 	high TEXT,
		// 	low TEXT,
		// 	varBid TEXT,
		// 	pctChange TEXT,
		// 	bid TEXT,
		// 	ask TEXT,
		// 	timestamp TEXT,
		// 	create_date TEXT
		// );`
		// 	_, err = db.Exec(create)
		// 	if err != nil {
		// 		w.WriteHeader(http.StatusInternalServerError)
		// 		log.Fatalf("Error happened in Create Table Cotacao. Err: %s", err)
		// 		return
		// 	}

		cotacao.ID = uuid.New().String()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err = insertCotacao(ctx, db, cotacao)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"bid": cotacao.Bid})

	}
}

func insertCotacao(ctx context.Context, db *sql.DB, cotacao *Cotacao) error {
	select {
	case <-ctx.Done():
		fmt.Println("Cotação not save in database. Timeout reached")
		return errors.New("cotação not save in database")
	default:
		stmt, err := db.Prepare("insert into cotacao(id, code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) values(?,?,?,?,?,?,?,?,?,?,?,?)")
		if err != nil {
			return err
		}
		_, err = stmt.Exec(
			cotacao.ID,
			cotacao.Code,
			cotacao.Codein,
			cotacao.Name,
			cotacao.High,
			cotacao.Low,
			cotacao.VarBid,
			cotacao.PctChange,
			cotacao.Bid,
			cotacao.Ask,
			cotacao.Timestamp,
			cotacao.CreateDate)
		if err != nil {
			return err
		}
	}
	return nil
}

func BuscaCotacao(ctx context.Context) (*Cotacao, error) {
	var c Cotacao
	select {
	case <-ctx.Done():
		fmt.Println("Cotação not requested. Timeout reached")
		return nil, errors.New("cotação not requested. Timeout reached")
	default:
		resp, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body[10:len(body)-1], &c)
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
}
