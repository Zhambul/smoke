package db

import (
	"os"
	"database/sql"
	"fmt"
	"time"
	"log"
	_ "github.com/lib/pq"
	"io/ioutil"
)

var db *sql.DB

func Init(runDdl bool) {
	host := os.Getenv("POSTGRES_HOST")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")

	t := time.NewTicker(1 * time.Second)
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-t.C:
			dataSource := fmt.Sprintf("host=%v user=%v password=%v dbname=%v sslmode=disable", host, user, password, dbName)
			log.Printf("datasource - %v\n", dataSource)
			_db, err := sql.Open("postgres", dataSource)
			if err != nil {
				log.Print(err)
				continue
			}
			if err := _db.Ping(); err != nil {
				log.Println("db connection is NOT successful, trying again")
				log.Print(err)
				continue
			}
			log.Println("db connection is successful")
			db = _db
			if runDdl {
				ddl()
			}
			return
		case <-timer.C:
			log.Fatal("could not connect to DB")
			break
		}
	}
}

func ddl() {
	db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public")
	bytes, err := ioutil.ReadFile("ddl.sql")
	if err != nil {
		log.Fatal(err)
	}
	sql := string(bytes)
	result, err1 := db.Exec(sql)
	log.Printf("ddl result %+v", result)
	if err1 != nil {
		log.Fatal(err1)
	}
}
