package main

import (
	"database/spl"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type OrderRequest struct {
	TerminalNo  string `json:"terminalNo"`
	MessageType string `json:"messageType"`
	TotalAmount int    `json:"totalAmount"`
	Items       []Item `json:"items"`
}

type Item struct {
	MenuName  string `json:"menuName"`
	UnitPrice int    `json:"unitPrice"`
	Quantity  int    `json:"quantity"`
}

// 注文登録 API
func handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	if r.Method == http.MethodPost {
		var req OrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 入力チェック
		if req.TerminalNo == "" || req.MessageType != "ORDER_CONFIRM" || req.TotalAmount < 1 || len(req.Items) < 1 || len(req.Items) > 5 {
			http.Error(w, "Validation Error", http.StatusBadRequest)
			return
		}

		calcTotal := 0
		menuNames := make(map[string]bool)
		for _, item := range req.Items {
			if item.MenuName == "" || item.UnitPrice < 1 || item.Quantity < 1 || item.Quantity > 5 || menuNames[item.MenuName] {
				http.Error(w, "Item Validation Error", http.StatusBadRequest)
				return
			}
			menuNames[item.MenuName] = true
			calcTotal += item.UnitPrice * item.Quantity
		}

		if calcTotal != req.TotalAmount {
			http.Error(w, "Total Amount Mismatch", http.StatusBadRequest)
			return
		}

		// 保存処理
		orderNo, _ := generateOrderNo()
		status := "オーダー受信"
		for i, item := range req.Items {
			subtotal := item.UnitPrice * item.Quantity
			_, err := db.Exec(`INSERT INTO order_items (order_no, terminal_no, order_status, item_no, menu_name, unit_price, quantity, subtotal) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				orderNo, req.TerminalNo, status, i+1, item.MenuName, item.UnitPrice, item.Quantity, subtotal)
			if err != nil {
				http.Error(w, "DB Error", http.StatusInternalServerError)
				return
			}
		}

		res := map[string]interface{}{
			"result":      "OK",
			"orderNo":     orderNo,
			"orderStatus": status,
			"totalAmount": req.TotalAmount,
			"message":     "注文を受け付けました",
		}
		writeLog(req, res)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)

	} else if r.Method == http.MethodGet {
		// 一覧取得
		status := r.URL.Query().Get("status")
		query := "SELECT order_no, order_status, SUM(subtotal) FROM order_items"
		var args []interface{}
		if status != "" {
			query += " WHERE order_status = ?"
			args = append(args, status)
		}
		query += " GROUP BY order_no"

		rows, _ := db.Query(query, args...)
		defer rows.Close()
		var results []map[string]interface{}
		for rows.Next() {
			var oNo, oStat string
			var total int
			rows.Scan(&oNo, &oStat, &total)
			results = append(results, map[string]interface{}{"orderNo": oNo, "orderStatus": oStat, "totalAmount": total})
		}
		json.NewEncoder(w).Encode(results)
	}
}

// 詳細取得・更新 API
func handleOrderDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	orderNo := parts[3]

	if r.Method == http.MethodGet {
		rows, _ := db.Query("SELECT menu_name, unit_price, quantity, subtotal FROM order_items WHERE order_no = ?", orderNo)
		defer rows.Close()
		var items []map[string]interface{}
		for rows.Next() {
			var name string
			var up, q, st int
			rows.Scan(&name, &up, &q, &st)
			items = append(items, map[string]interface{}{"menuName": name, "unitPrice": up, "quantity": q, "subtotal": st})
		}
		json.NewEncoder(w).Encode(items)

	} else if r.Method == http.MethodPut {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		newStatus := body["orderStatus"]
		_, err := db.Exec("UPDATE order_items SET order_status = ? WHERE order_no = ?", newStatus, orderNo)
		if err != nil {
			http.Error(w, "Update Error", http.StatusInternalServerError)
			return
		}
		writeLog("Update Status: "+orderNo, newStatus)
		json.NewEncoder(w).Encode(map[string]string{"result": "OK", "newStatus": newStatus})
	}
}