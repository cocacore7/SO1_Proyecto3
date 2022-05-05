package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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

	router.HandleFunc("/Games/game/{game}/gamename/{gamename}/players/{players}", Juego).Methods("GET")
	router.HandleFunc("/", Test).Methods("GET")

	handler := cors.Default().Handler(router)
	log.Fatal(http.ListenAndServe(":3000", handler))
}

func Juego(w http.ResponseWriter, resp *http.Request) {
	idGame := mux.Vars(resp)["game"]
	players := mux.Vars(resp)["players"]
	gamename := mux.Vars(resp)["gamename"]
	fmt.Print(idGame + players + gamename)

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

	var DatosGame Datos

	err = json.Unmarshal([]byte(idGame), &DatosGame.Game)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(players), &DatosGame.Players)
	if err != nil {
		return
	}

	r, err := c.Jugar(ctx, &pb.JuegoRequest{
		Game:    DatosGame.Game,
		Players: DatosGame.Players,
	})
	if err != nil {
		log.Fatalf("Error al jugar: %v", err)
	}

	_ = json.NewEncoder(w).Encode(r.GetResultado())
}

func Test(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode("Cliente Go Funciona")
}
