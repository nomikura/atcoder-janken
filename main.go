package janken

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type AllData struct {
	ID1        string
	ID2        string
	Result     string
	ResultHTML string
	Results    []Result
	Date_start string
	Date_end string
}

type Result struct {
	Title           string
	ID              string
	Place1          int
	Place2          int
	BackGroundColor1 string
	BackGroundColor2 string
}

type History struct {
	IsRated           bool      `json:"IsRated"`
	Place             int       `json:"Place"`
	NewRating         int       `json:"NewRating"`
	Performance       int       `json:"Performance"`
	InnerPerformance  int       `json:"InnerPerformance"`
	ContestScreenName string    `json:"ContestScreenName"`
	ContestName       string    `json:"ContestName"`
	EndTime           time.Time `json:"EndTime"`
}

func init() {
	// router := gin.Default()
	router := gin.New()

	router.GET("/", func(c *gin.Context) {
		// idを取得する
		id1, id2, date_start, date_end := c.Query("id1"), c.Query("id2"), c.Query("date_start"), c.Query("date_end")
		

		// 結果を取得する
		data := GetData(id1, id2, date_start, date_end, c)

		// テンプレートHTMLにデータを入れる
		t, _ := template.ParseFiles("main.html")
		t.Execute(c.Writer, data)

		// c.String(http.StatusOK, "%s \n\n %s", history1, history2)
	})

	http.Handle("/", router)
	// router.Run(":8080")
}

func GetData(id1 string, id2 string, date_start string, date_end string, c *gin.Context) AllData {
	// 2人のhistoryを取得
	var history1, history2 []History
	SetUserHistory(id1, &history1, c)
	SetUserHistory(id2, &history2, c)

	// history2のmapを作る. [contestID]Place
	history2Map := make(map[string]int)
	for _, contest := range history2 {
		history2Map[contest.ContestScreenName] = contest.Place
	}

	white := "FFFFFF"
	red := "EBCCCC"
	green := "D0E9C6"

	layout := "2006-01"
	ds, _ := time.Parse(layout, date_start)
	dt, _ := time.Parse(layout, date_end)

	// 結果を生成
	var result []Result
	var id1Count, id2Count = 0, 0
	for _, contest := range history1 {
		// ID2の順位
		place2, ok := history2Map[contest.ContestScreenName]
		// ID2がコンテストに出ていなければスキップ
		if !ok {
			continue
		}

		if contest.EndTime.Before(ds) {
			continue
		}
		
		if contest.EndTime.After(dt) {
			continue
		}

		backGroundColor1 := white
		backGroundColor2 := white

		if contest.Place < place2 {
			backGroundColor1 = green
			backGroundColor2 = red
			id1Count++
		} else if contest.Place > place2 {
			backGroundColor1 = red
			backGroundColor2 = green
			id2Count++
		} else {
			id1Count++
			id2Count++
		}

		result = append(result, Result{
			Title:           contest.ContestName,
			Place1:          contest.Place,
			Place2:          place2,
			BackGroundColor1: backGroundColor1,
			BackGroundColor2: backGroundColor2,
		})
	}

	resultStr := "【AtCoderじゃんけん】\n"
	resultStr += id1 + " vs " + id2 + "\n"
	resultStr += strconv.Itoa(id1Count) + "対" + strconv.Itoa(id2Count) + "で"
	resultHTML := strconv.Itoa(id1Count) + "対" + strconv.Itoa(id2Count) + "で"
	if id1Count > id2Count {
		resultStr += id1 + "の勝利です！！"
		resultHTML += id1 + "の勝利です！！"
	} else if id1Count < id2Count {
		resultStr += id2 + "の勝利です！！"
		resultHTML += id2 + "の勝利です！！"
	} else {
		resultStr += "勝負は引き分けです！！"
		resultHTML += "勝負は引き分けです！！"
	}
	resultHTML += "(" + date_start + " ~ " + date_end + ")"
	resultStr += "(" + date_start + " ~ " + date_end + ")"
	resultStr += "\n"
	resultStr += "#AtCoderじゃんけん\n"

	data := AllData{
		ID1:        id1,
		ID2:        id2,
		Result:     resultStr,
		ResultHTML: resultHTML,
		Results:    result,
		Date_start: date_start,
		Date_end: date_end,
	}

	return data
}

func SetUserHistory(id string, history *[]History, c *gin.Context) {
	var request *http.Request = c.Request
	context := appengine.NewContext(request)
	client := urlfetch.Client(context)

	// GET
	userURL := "https://beta.atcoder.jp/users/" + id + "/history/json"
	resp, err := client.Get(userURL)
	if err != nil {
		log.Infof(context, "%v", err)
	}

	// バイナリ取得
	byteArray, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(byteArray, &history)
}
