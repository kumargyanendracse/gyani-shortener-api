package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type LogEntry struct {
	Service   string                 `json:"service"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	RequestID string                 `json:"request_id"`
	UserID    string                 `json:"user_id"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func createLog(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		return
	}

	var logData LogEntry
	json.NewDecoder(r.Body).Decode(&logData)
	metaJSON, _ := json.Marshal(logData.Metadata)
	_, err := db.Exec(
		`insert into logs(service,level,message,request_id,user_id,metadata)
		 values($1,$2,$3,$4,$5,$6)`,
		logData.Service,
		logData.Level,
		logData.Message,
		logData.RequestID,
		logData.UserID,
		metaJSON,
	)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	level := r.URL.Query().Get("level")
	service := r.URL.Query().Get("service")

	query := "select service,level,message,request_id,user_id,metadata,created_at from logs where 1=1"
	args := []interface{}{}
	i := 1

	if level != "" {
		query += " and level=$" + strconv.Itoa(i)
		args = append(args, level)
		i++
	}
	if service != "" {
		query += " and service=$" + strconv.Itoa(i)
		args = append(args, service)
		i++
	}

	query += " order by created_at desc limit 100"

	rows, _ := db.Query(query, args...)

	var logs []map[string]interface{}

	for rows.Next() {
		var service, level, message, requestID, userID string
		var metadata []byte
		var created time.Time

		rows.Scan(&service, &level, &message, &requestID, &userID, &metadata, &created)

		var meta interface{}
		json.Unmarshal(metadata, &meta)

		logs = append(logs, map[string]interface{}{
			"service":    service,
			"level":      level,
			"message":    message,
			"request_id": requestID,
			"user_id":    userID,
			"metadata":   meta,
			"created_at": created,
		})
	}

	json.NewEncoder(w).Encode(logs)
}
