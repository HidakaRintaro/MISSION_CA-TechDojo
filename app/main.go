package main

import (
	"database/sql"
	"encoding/json"
	_ "fmt"
	"log"
	"net/http"

	_ "github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	Name  string `json:"name,omitempty"`
	Token string `json:"x-token,omitempty"`
}

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/user/create", UserCreateRequest).Methods("POST")
	r.HandleFunc("/user/get", UserGetResponse).Methods("GET")
	r.HandleFunc("/user/update", UserUpdateRequest).Methods("PUT")

	log.Println("サーバー起動 : 8080 port で受信")

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":8080", r))

}

func UserCreateRequest(w http.ResponseWriter, r *http.Request) {
	headerToken := "aaabbbcc"

	var user User
	user.Token = headerToken
	
	json.NewDecoder(r.Body).Decode(&user)

	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	stmtInsert, _ := db.Prepare("INSERT INTO user(`name`, `x-token`) VALUES(?, ?)")
	defer stmtInsert.Close()

	stmtInsert.Exec(user.Name, headerToken)

	json.NewEncoder(w).Encode(user)
}

//func UserCreateResponse(w http.ResponseWriter, r *http.Request)  {
//
//}

func UserGetResponse(w http.ResponseWriter, r *http.Request) {
	headerToken := r.Header.Get("x-token")
	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	var user User

	row := db.QueryRow("SELECT name FROM user WHERE `x-token` = ?", headerToken)

	row.Scan(&user.Name)

	json.NewEncoder(w).Encode(user)
}

func UserUpdateRequest(w http.ResponseWriter, r *http.Request) {
	headerToken := r.Header.Get("x-token")

	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	var user User
	json.NewDecoder(r.Body).Decode(&user)

	stmtInsert, _ := db.Prepare("UPDATE user SET name = ? WHERE `x-token` = ?")
	defer stmtInsert.Close()

	stmtInsert.Exec(user.Name, headerToken)
}
