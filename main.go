package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// ログ設定
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}
	initDB()

	// ルーティング
	mux := http.NewServeMux()
	mux.HandleFunc("/api/orders", corsMiddleware(handleOrders))
	mux.HandleFunc("/api/orders/", corsMiddleware(handleOrderDetail))

	fmt.Println("サーバー起動: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func writeLog(req interface{}, res interface{}) {
	f, _ := os.OpenFile("logs/order.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	entry := map[string]interface{}{
		"time":     "2026-05-13 11:26:43", // 本日の日付
		"request":  req,
		"response": res,
	}
	jsonBytes, _ := json.Marshal(entry)
	f.WriteString(string(jsonBytes) + "\n")
}