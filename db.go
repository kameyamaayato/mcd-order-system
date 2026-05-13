package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./order.db")
	if err != nil {
		log.Fatal(err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS order_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_no TEXT NOT NULL,
		terminal_no TEXT NOT NULL,
		order_status TEXT NOT NULL,
		item_no INTEGER NOT NULL,
		menu_name TEXT NOT NULL,
		unit_price INTEGER NOT NULL,
		quantity INTEGER NOT NULL,
		subtotal INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func generateOrderNo() (string, error) {
	today := time.Now().Format("0102")
	var count int
	query := "SELECT COUNT(DISTINCT order_no) FROM order_items WHERE order_no LIKE ?"
	err := db.QueryRow(query, today+"-%").Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%03d", today, count+1), nil
}