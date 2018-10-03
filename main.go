package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/lib/pq"

	nsq "github.com/bitly/go-nsq"
	redigo "github.com/garyburd/redigo/redis"
	_ "github.com/lib/pq"
)

var dbHome *sql.DB
var redisPool *redigo.Pool
var err error

type (
	User struct {
		User_id     int
		User_name   string
		Msisdn      string
		User_email  string
		Birth_date  pq.NullTime
		Create_time pq.NullTime
		Update_time pq.NullTime
	}

	renderData struct {
		PageTitle  string
		User       []User
		Keyword    string
		AccessTime string
	}
)

func main() {

	// Consumer Init
	wg := &sync.WaitGroup{}
	wg.Add(1)

	config := nsq.NewConfig()
	q, _ := nsq.NewConsumer("bigproject_user", "mychannel", config)
	q.AddHandler(nsq.HandlerFunc(func(message *nsq.Message) error {
		log.Printf("Got a message: %s", message)
		AddRedisCountLoadPage()
		// wg.Done()
		return nil
	}))
	err := q.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		log.Panic("Could not connect")
	}
	// wg.Wait()

	// init db
	dbHome, err = sql.Open("postgres", "postgres://tkpdtraining:1tZCjrIcYeR1uQblQz0gBlIFU@devel-postgre.tkpd/tokopedia-dev-db?sslmode=disable")
	if err != nil {
		log.Print(err)
	}

	// int redis
	redisPool = redigo.NewPool(func() (redigo.Conn, error) {
		return redigo.Dial("tcp", "devel-redis.tkpd:6379")
	}, 1)

	// route
	http.HandleFunc("/user", handleShowData)
	http.HandleFunc("/filter", handleShowDataFilterByName)
	fmt.Println("Server start on :8181 ...")
	http.ListenAndServe(":8181", nil)

}

func doPublish() {
	// nsq publisher
	config := nsq.NewConfig()
	w, _ := nsq.NewProducer("127.0.0.1:4150", config)

	err := w.Publish("bigproject_user", []byte("reloaded"))
	if err != nil {
		log.Panic("Could not connect")
	}
}

func getMultipleUser(keyword string) []User {
	// log.Println(keyword)
	like_keyword := strings.Join([]string{"%", keyword, "%"}, "")

	stmt := "select user_id, full_name, msisdn, user_email, birth_date, create_time, update_time FROM public.ws_user WHERE LOWER(full_name) LIKE $1 limit 10"

	listUser := []User{}

	// rows, err := dbHome.Query(stmt)
	rows, err := dbHome.Query(stmt, like_keyword)
	if err != nil {
		log.Println(err)
		return listUser
	}

	for rows.Next() {
		u := User{}
		err := rows.Scan(&u.User_id, &u.User_name, &u.Msisdn, &u.User_email, &u.Birth_date, &u.Create_time, &u.Update_time)
		if err != nil {
			log.Println(err)
		}
		listUser = append(listUser, u)
	}

	rows.Close()
	return listUser
}

func handleShowData(w http.ResponseWriter, r *http.Request) {

	// publish message to add counter
	doPublish()

	// getMultipleUser get data user from db
	user_data := getMultipleUser("")

	// printIntoView render into html page
	printIntoView(w, user_data, "")

}

func handleShowDataFilterByName(w http.ResponseWriter, r *http.Request) {

	// publish message to add counter
	doPublish()

	// get Param from url
	queryValues := r.URL.Query()
	keyword := queryValues.Get("keyword")

	// send param as query clause
	user_data := getMultipleUser(keyword)

	// printIntoView render into html page
	printIntoView(w, user_data, keyword)

}

func printIntoView(w http.ResponseWriter, u []User, keyword string) {

	// used to set accessTime value
	count := "0"
	if countAccess, err := GetRedis("training_db:hendrap"); err != nil {
		log.Println(err)
	} else {
		count = countAccess
	}

	// used to render from template
	htmltemplate, err := template.ParseFiles("template/user_row.html")
	if err != nil { // if there is an error
		log.Print("template parsing error: ", err) // log it
	}
	tmpl := template.Must(htmltemplate, err)
	fmt.Println(count)

	data := renderData{
		PageTitle:  "My List User",
		User:       u,
		Keyword:    keyword,
		AccessTime: count,
	}
	tmpl.Execute(w, data)

}
