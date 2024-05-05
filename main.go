package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/context"
	"github.com/joho/godotenv"
)

var tpl = template.Must(template.ParseGlob("templates/*.html"))

type Motion struct {
	Uuid string `json:"uuid"`
	Info string `json:"info"`
	Time string `json:"time"`
}

func dbConn() (db *sql.DB) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbDriver := os.Getenv("DB_DRIVER")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	fmt.Println(dbDriver, dbUser, dbPass, dbName)
	db, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp(127.0.0.1:3306)/"+dbName+"?parseTime=true")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("DB Connected!!")
	return db
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	row, err := db.Query("SELECT info, time FROM motion")
	if err != nil {
		log.Println(err)
	}
	defer row.Close()
	motion := []Motion{}

	for row.Next() {
		p := Motion{}
		err := row.Scan(&p.Info, &p.Time)
		if err != nil {
			fmt.Println(err)
			continue
		}
		motion = append(motion, p)
	}
	tpl.ExecuteTemplate(w, "index.html", motion)
}

func insertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		db := dbConn()
		uuid := uuid.NewString()
		info := "Motion detected!"
		today := time.Now().Format("2006-01-02 15:04:05")
		_, err := db.Exec("INSERT INTO motion (uuid, info, time) VALUES(?,?,?)", uuid, info, today)
		if err != nil {
			fmt.Println("Error when inserting: ", err.Error())
			panic(err.Error())
		}
		http.Redirect(w, r, "/index", http.StatusMovedPermanently)
	} else if r.Method == "GET" {
		tpl.ExecuteTemplate(w, "insert.html", nil)
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/insert", insertHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started on: http://localhost:8080")
	err := http.ListenAndServe(":8080", context.ClearHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}
}
