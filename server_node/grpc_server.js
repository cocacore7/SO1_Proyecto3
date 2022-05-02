var PROTO_PATH = './proto/juego.proto';

var grpc = require('@grpc/grpc-js');
var protoLoader = require('@grpc/proto-loader');
var amqp = require('amqplib/callback_api');
const { functionsIn } = require('lodash');
var packageDefinition = protoLoader.loadSync(
    PROTO_PATH,
    {keepCase: true,
     longs: String,
     enums: String,
     defaults: true,
     oneofs: true
    });
var juego_proto = grpc.loadPackageDefinition(packageDefinition).proto;
var RabbitUser = process.env.RABBIT_USERNAME || 'guest';
var RabbitPass = process.env.RABBIT_PASSWORD || 'guest';
var RabbitHost = process.env.RABBIT_HOST || 'localhost';
var RabbitPort = process.env.RABBIT_PORT || '5672';

function Jugar(call, callback) {
  if (call.request.players > 0){
    let JuegoDatos = Juegos(call.request.game, CrearJugadores(call.request.players));
    callback(null, {resultado: 'Juego-id: ' + call.request.game + ', Juego: ' + JuegoDatos[1] + ', jugadores: ' + call.request.players + ', Ganador: ' + JuegoDatos[0]});
    //Envair ganador por rabitMQ
    amqp.connect('amqp://' + RabbitUser + ":" + RabbitPass + '@' + RabbitHost + ":" + RabbitPort , function(error0, connection) {
      if (error0) {
          throw error0;
      }
      connection.createChannel(function(error1, channel) {
        if (error1) {
          throw error1;
        }

        var queue = 'game';
        var game = {
          game_id: call.request.game,
          players: call.request.players,
          winner:  JuegoDatos[0],
          game_n: JuegoDatos[1]
        };

        channel.assertQueue(queue, {
          durable: false
        });
        channel.sendToQueue(queue, Buffer.from(JSON.stringify(game)));

        console.log(" [x] Enviado: %s", JSON.stringify(game));
      });
      setTimeout(function() {
          connection.close();
      }, 500);
    });
  }else{
    callback(null, {resultado: 'No se puede jugar, 0 jugadores'})
  }
}

function CrearJugadores(players) {
  let jugadores = []
  for(var i = 0; i < players;i++){
    jugadores.push("Jugador" + (i+1).toString());
  }
  return jugadores
}

function Juegos(game, players) {
  switch (game){
    case 1:
      var pos = Math.floor(Math.random()*players.length);
      return [players[pos], "Random"];
    case 2:
      return [players[0], "First"];
    case 3:
      return [players.pop(), "Last"];
    case 4:
      var pos = Math.round (players.length / 2);
      return [players[pos-1], "Mitad"];
    case 5:
      let jugadoresPrimos = []
      for(var i = 1; i <= players.length;i++){
        if (primo(i)){
          jugadoresPrimos.push(players[i-1]);
        }
      }
      var pos = Math.floor(Math.random()*jugadoresPrimos.length);
      return [jugadoresPrimos[pos], "RandomPrimo"];
  }
}

function primo(numero) {
  for (var i = 2; i < numero; i++) {
    if (numero % i === 0) {
      return false;
    }
  }
  return numero !== 1;
}

function main() {
  var server = new grpc.Server();
  server.addService(juego_proto.Juego.service, {Jugar: Jugar});
  server.bindAsync('0.0.0.0:50051', grpc.ServerCredentials.createInsecure(), () => {
    server.start();
    console.log('Servidor gRPc en puerto 50051')
  });
}

main();
