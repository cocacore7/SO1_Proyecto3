package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	pb "github.com/cocacore7/grpc/proto"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Datos struct {
	Game    int32 `json:"game_id"`
	Players int32 `json:"players"`
}

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/Games", Juego).Methods("POST")
	router.HandleFunc("/", Test).Methods("GET")

	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":4000", handler))
}

func Juego(w http.ResponseWriter, resp *http.Request) {
	flag.Parse()
	conn, err := grpc.Dial(os.Getenv("GRCP_SERVER"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("no conecto: %v", err)
	}
	defer func(conn *grpc.ClientConn) {
		err = conn.Close()
		if err != nil {
			log.Fatalf("%v", err)
		}
	}(conn)
	c := pb.NewJuegoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var Datos Datos
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &Datos)
	if err != nil {
		return
	}

	r, err := c.Jugar(ctx, &pb.JuegoRequest{
		Game:    Datos.Game,
		Players: Datos.Players,
	})
	if err != nil {
		log.Fatalf("Error al jugar: %v", err)
	}

	_ = json.NewEncoder(w).Encode(r.GetResultado())
}

func Test(w http.ResponseWriter, resp *http.Request) {
	_ = json.NewEncoder(w).Encode("Cliente Go Funciona")
}
