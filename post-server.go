package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost:9092"})

	if err != nil {
		log.Fatal(err)
	}
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/posts/", func(writer http.ResponseWriter, request *http.Request) {
		fetchPost(writer, request, producer)
	})
	rtr.HandleFunc("/login", handleLogin)
	http.ListenAndServe(":8080", nil)

	defer producer.Close()
}

func fetchPost(w http.ResponseWriter, r *http.Request, producer *kafka.Producer) {
	path := strings.Split(r.URL.Path, "/")
	postId, err := strconv.Atoi(path[len(path)-1])

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error %v", err)
		return
	}

	subject, err := validateRequestAndGetSubject(r)

	if err != nil || subject == "" {
		w.WriteHeader(403)
		fmt.Fprint(w, "Invalid request")
		return
	}

	postView := make(map[string]interface{})
	postView["user"] = subject
	postView["postId"] = postId

	go pushToAnalyticsQueue(postView, producer)

	//Go to the post DB, fetch the post by id and return here
	fmt.Fprintf(w, "POST %v", postId)
}

func pushToAnalyticsQueue(postView map[string]interface{}, producer *kafka.Producer) {
	topic := "post_views"

	body, err := json.Marshal(postView)

	if err == nil {
		producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          body,
		}, nil)
		fmt.Println(postView)
	}

}

type User struct {
	Username string
	Password string `json:"-"`
}

type Claims struct {
	Subject string `json:"sub"`
	jwt.StandardClaims
}

func validateRequestAndGetSubject(r *http.Request) (string, error) {
	type Claims struct {
		Subject string `json:"sub"`
		jwt.StandardClaims
	}
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return "", errors.New("Invalid request")
	}

	tokenString := strings.Fields(authorizationHeader)[1]
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("MY_SECRET_HIDDEN_KEY"), nil
	})

	if claims, ok := token.Claims.(*Claims); ok && token.Valid && err == nil {
		return claims.Subject, nil
	} else {
		return "", err
	}

}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	user := User{} //initialize empty user

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "Error %v", err)
		return
	}

	signingKey := []byte("MY_SECRET_HIDDEN_KEY")
	claims := Claims{
		user.Username,
		jwt.StandardClaims{
			Issuer: "AUTH_SERVER",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(signingKey)

	fmt.Fprint(w, signedToken)
}
