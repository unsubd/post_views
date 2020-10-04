package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	// connect to the cluster
	cluster := gocql.NewCluster("192.168.56.5")
	cluster.Keyspace = "analytics"
	cluster.Consistency = gocql.One
	session, _ := cluster.CreateSession()

	defer  session.Close()

	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	//Fetch all users who have read a particular post
	rtr.HandleFunc("/analytics/posts/{id}/users", func(writer http.ResponseWriter, request *http.Request) {
		fetchUsersForPost(writer, request, session)
	})

	//Fetch all posts read by a particular user
	rtr.HandleFunc("/analytics/users/{id}/posts", func(writer http.ResponseWriter, request *http.Request) {
		filterPostsByUser(writer, request, session)
	})

	http.ListenAndServe(":9090", nil)
}

func fetchUsersForPost(writer http.ResponseWriter, request *http.Request, session *gocql.Session) {
	postId := mux.Vars(request)["id"]

	query := fmt.Sprintf("select users from post_users where postId='%v'", postId)
	iter := session.Query(query).Iter()
	var users []string
	for iter.Scan(&users) {
		break
	}
	writer.Header().Add("content-type", "application/json")

	json.NewEncoder(writer).Encode(users)

}

func filterPostsByUser(writer http.ResponseWriter, request *http.Request, session *gocql.Session) {
	userId := mux.Vars(request)["id"]

	query := fmt.Sprintf("select postId from post_users where users contains '%v'", userId)
	iter := session.Query(query).Iter()
	var postId string
	var posts []string

	for iter.Scan(&postId) {
		posts = append(posts, postId)
	}

	writer.Header().Add("content-type", "application/json")
	json.NewEncoder(writer).Encode(posts)
}
