package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Cola struct {
	Game     int32  `json:"game_id"`
	Players  int32  `json:"players"`
	Winner   string `json:"winner"`
	GameName string `json:"game_n"`
}

type Log struct {
	GameId   int32  `json:"game_id"`
	Players  int32  `json:"players"`
	Winner   string `json:"winner"`
	GameName string `json:"game_n"`
	Queue    string `json:"queue"`
	Fecha    string `json:"Fecha"`
}

var collection *mongo.Collection
var ctx = context.TODO()

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@" + os.Getenv("ADD_RABBIT") + "/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"game", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			//Crear log para Mongo
			var Cola Cola
			var Log Log
			body, err := ioutil.ReadAll(bytes.NewReader([]byte(d.Body)))
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(body, &Cola)
			if err != nil {
				return
			}
			t := time.Now()
			Log.GameId = Cola.Game
			Log.Players = Cola.Players
			Log.GameName = Cola.GameName
			Log.Winner = Cola.Winner
			Log.Queue = "RabbitMQ"
			Log.Fecha = t.Format("2006-01-02 15:04:05")

			//Almacenar Mongo//Conectar con mongodb
			clientOptions := options.Client().ApplyURI("mongodb://admin:pass123@" + os.Getenv("ADD_MONGO") + ":27017")
			client, err := mongo.Connect(ctx, clientOptions)
			if err != nil {
				log.Fatal(err)
			}
			//Crear colleccion y base de datos si no existen y registrar en coleccion
			collection = client.Database("SO1_Proyecto3").Collection("Game_Logs")
			respuesta, err := collection.InsertOne(context.TODO(), Log)
			if err != nil {
				fmt.Print("Logs No Registrado")
				fmt.Print(respuesta)
				panic(err)
			}

			//Convertir datos Log en JSON
			tiempoReal, err := json.Marshal(Log)
			if err != nil {
				fmt.Println(err)
				return
			}

			//Conectar con clientes
			clientredis := redis.NewClient(&redis.Options{
				Addr:     os.Getenv("ADD_REDIS"),
				Password: "",
				DB:       0,
			})

			clienttidis := redis.NewClient(&redis.Options{
				Addr:     os.Getenv("ADD_TIDIS"),
				Password: "",
				DB:       0,
			})

			//Almacenar registros redis y tidis

			errorReids := clientredis.Set("tiempoReal", tiempoReal, 0).Err()
			if err != nil {
				panic(errorReids)
			}

			errorTidis := clienttidis.Set("tiempoReal", tiempoReal, 0).Err()
			if err != nil {
				panic(errorTidis)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
