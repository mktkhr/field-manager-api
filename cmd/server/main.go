package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("サーバーを起動しています...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("サーバーがポート :8080 で起動しました")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
