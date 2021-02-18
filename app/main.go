package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {

	db, _ := sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	fmt.Printf("%T\n", db)
	r := mux.NewRouter()

	r.HandleFunc("/user/create", UserCreate).Methods("POST")
	r.HandleFunc("/user/get", UserGet).Methods("GET")
	r.HandleFunc("/user/update", UserUpdate).Methods("PUT")

	r.HandleFunc("/gacha/draw", GachaDraw).Methods("POST")

	r.HandleFunc("/character/list", CharacterList).Methods("GET")

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

func DbInfo() (db *sql.DB) {
	db, _ = sql.Open("mysql", "root:root@tcp(mysql:3306)/ca_tech_dojo")
	return db
}

// ==================== User ====================

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

	db := DbInfo()
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

	db := DbInfo()
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

	db := DbInfo()
	defer db.Close()

	var reqUser UserUpdateRequest

	json.Unmarshal(StreamToByte(r.Body), &reqUser)

	stmtInsert, _ := db.Prepare("UPDATE user SET name = ? WHERE `x-token` = ?")
	defer stmtInsert.Close()
	stmtInsert.Exec(reqUser.Name, token)
}

// ==================== Gacha ====================

type GachaDrawRequest struct {
	Times int `json:"times"`
}
type GachaResult struct {
	CharacterID string `json:"characterID"`
	Name        string `json:"name"`
}
type GachaDrawResponse struct {
	Results []GachaResult
}

type Weight struct {
	CharacterId int
	Rate int
}
func GachaDraw(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")

	var reqGacha GachaDrawRequest
	json.Unmarshal(StreamToByte(r.Body), &reqGacha)

	/* キャラクターの取得 */
	// DBから指定ガチャのキャラの排出率を取得
	gachaId := 1
	var weights []Weight
	var weight Weight
	db := DbInfo()
	defer db.Close()
	row := db.QueryRow("SELECT character_id, rate FROM emission WHERE gacha_id = ?", gachaId)
	for row.Next() {
		row.Scan(&weight.CharacterId, &weight.Rate)
		weights = append(weights, weight)
	}



	/* キャラクターの保存 */

}
// 重み付き乱数
func CharacterGet(Weights []int) {
	sort.Ints(Weights) // 重みの昇順でチェックしていくので重み一覧をソートする

	// 抽選結果チェックの基準となる境界値を生成
	boundaries := make([]int, len(Weights)+1) // intのスライスを作成(重み一覧より１大きい容量)

	for i := 1; i < len(Weights)+1; i++ {
		boundaries[i] = boundaries[i-1] + Weights[i-1]
	}
	boundaries = boundaries[1:len(boundaries)] // 頭の0値は不要なので破棄

	rand.Seed(time.Now().UnixNano())
	result := make([]int, len(Weights))

	draw := rand.Intn(boundaries[len(boundaries)-1]) // + 1

	for i, boundary := range boundaries {

		if draw <= boundary {
			result[i]++
			break

		}
	}
}

// ==================== Character ====================

type UserCharacter struct {
	UserCharacterID string `json:"userCharacterID"`
	CharacterID     string `json:"characterID"`
	Name            string `json:"name"`
}
type CharacterListResponse struct {
	Characters []UserCharacter `json:"characters"`
}

func CharacterList(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")

	db := DbInfo()
	defer db.Close()

	var resUser CharacterListResponse

	row := db.QueryRow(
		"SELECT  FROM `possession` p JOIN `user` u ON u.id = p.user_id WHERE u.x-token = ?",
		token,
	)

	row.Scan(&resUser)

	b, _ := json.Marshal(resUser)
	w.Write(b)
}
