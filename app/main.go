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

func Random() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 10)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
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
	token := Random()
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
	CharacterID int    `json:"characterID,string"`
	Name        string `json:"name"`
}
type GachaDrawResponse struct {
	Results []GachaResult
}

func GachaDraw(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("x-token")
	var weight = []int{
		5,  // S
		15, // A
		30, // B
		50, // C
	}
	var rarity = []string{"S", "A", "B", "C"}
	var rank string

	var reqGacha GachaDrawRequest
	json.Unmarshal(StreamToByte(r.Body), &reqGacha)

	/* キャラクターの取得 */
	// DBから指定ガチャのキャラの排出率を取得
	gachaId := 3 // 今回はガチャの種類を固定値とする
	var gacharesult GachaResult
	var rankCharacter []GachaResult // ガチャのランクで取得できたキャラの一時保管
	var result []GachaResult        // ガチャ結果
	db := DbInfo()
	defer db.Close()

	for time := 0; time < reqGacha.Times; time++ {

		boundaries := make([]int, len(weight)+1)
		for i := 1; i < len(weight)+1; i++ {
			boundaries[i] = boundaries[i-1] + weight[i-1]
		}
		boundaries = boundaries[1:len(boundaries)] // 頭の0値は不要なので破棄

		draw := rand.Intn(boundaries[len(boundaries)-1])

		for i, boundary := range boundaries {
			if draw <= boundary {
				rank = rarity[i]
				break
			}
		}
		rows, _ := db.Query("SELECT e.character_id, c.name FROM emission e JOIN `character` c ON c.id = e.character_id WHERE e.gacha_id = ? AND c.rarity = ?",
			gachaId,
			rank,
		)
		cnt := 0
		for rows.Next() {
			rows.Scan(&gacharesult.CharacterID, &gacharesult.Name)
			rankCharacter = append(rankCharacter, gacharesult)
			cnt++
		}
		result = append(result, rankCharacter[rand.Intn(cnt)])
	}

	var resGacha GachaDrawResponse
	resGacha.Results = result

	/* tokenからuseridの取得 */
	var id int
	db.QueryRow("SELECT id FROM user WHERE `x-token` = ?", token).Scan(&id)

	/* キャラクターの保存 */
	for _, character := range result {
		stmtInsert, _ := db.Prepare("INSERT INTO possession (`character_id`, `user_id`) VALUES(?, ?)")
		stmtInsert.Exec(character.CharacterID, id)
		stmtInsert.Close()
	}

	b, _ := json.Marshal(resGacha)
	w.Write(b)
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

	var userCharacter UserCharacter
	var resList CharacterListResponse

	rows, _ := db.Query(
		"SELECT p.id AS userCharacterID, c.id AS characterID, c.name FROM `possession` p JOIN `user` u ON u.id = p.user_id JOIN `character` c ON c.id = p.character_id WHERE u.`x-token` = ?",
		token,
	)

	for rows.Next() {
		rows.Scan(&userCharacter.UserCharacterID, &userCharacter.CharacterID, &userCharacter.Name)
		resList.Characters = append(resList.Characters, userCharacter)
	}

	b, _ := json.Marshal(resList)
	w.Write(b)
}
