package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"log"
)

func main() {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "postViewAnalyticsConsumerGroup",
		"auto.offset.reset": "latest",
	})

	// connect to the cluster
	cluster := gocql.NewCluster("192.168.56.5")
	cluster.Keyspace = "analytics"
	cluster.Consistency = gocql.One
	session, _ := cluster.CreateSession()
	defer session.Close()

	if err != nil {
		panic(err)
	}

	consumer.SubscribeTopics([]string{"post_views"}, nil)

	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			var message map[string]interface{}
			json.Unmarshal(msg.Value, &message)
			go updateCounts(message, session)
		} else {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	consumer.Close()
}

func updateCounts(postView map[string]interface{}, session *gocql.Session) {
	user := postView["user"]
	postID := postView["postId"]
	query := fmt.Sprintf("update post_users set users = users + {'%v'} where postId = '%v'", user, postID)
	if err := session.Query(query).Exec(); err != nil {
		log.Println(err)
	}

	query = fmt.Sprintf("update post_views set counts = counts + 1 where postId = '%v'", postID)
	if err := session.Query(query).Exec(); err != nil {
		log.Println(err)
	}

}
