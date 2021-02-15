package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/user/create", UserCreate).Methods("POST")
	r.HandleFunc("/user/get", UserGet).Methods("GET")
	r.HandleFunc("/user/update", UserUpdate).Methods("PUT")

	log.Println("サーバー起動 : 8080 port で受信")

	// log.Fatal は、異常を検知すると処理の実行を止めてくれる
	log.Fatal(http.ListenAndServe(":8080", r))

}

// io.Readerをbyteのスライスに変換
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

type UserCreateRequest struct {
	Name string `json:"name"`
}
type UserCreateResponse struct {
	Token string `json:"token"`
}

func UserCreate(w http.ResponseWriter, r *http.Request) {
	var reqUser UserCreateRequest

	json.Unmarshal(StreamToByte(r.Body), &reqUser)
	token := "aaabbbcc"
	resUser := UserCreateResponse{
		Token: token,
	}

	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	stmtInsert, _ := db.Prepare("INSERT INTO user(`name`, `x-token`) VALUES(?, ?)")
	defer stmtInsert.Close()
	stmtInsert.Exec(reqUser.Name, resUser.Token)

	b, _ := json.Marshal(resUser)
	w.Write(b)
}

type UserGetResponse struct {
	Name string `json:"name"`
}

func UserGet(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	var resUser UserGetResponse

	row := db.QueryRow("SELECT name FROM user WHERE `x-token` = ?", token)

	row.Scan(&resUser.Name)

	b, _ := json.Marshal(resUser)
	w.Write(b)
}

type UserUpdateRequest struct {
	Name string `json:"name"`
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")

	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	defer db.Close()

	var reqUser UserUpdateRequest

	json.Unmarshal(StreamToByte(r.Body), &reqUser)

	stmtInsert, _ := db.Prepare("UPDATE user SET name = ? WHERE `x-token` = ?")
	defer stmtInsert.Close()
	stmtInsert.Exec(reqUser.Name, token)
}
