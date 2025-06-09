package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	// Aún no importamos el paquete "node", lo haremos en el siguiente paso.
	// "mi-tarea-sd/internal/node"
)

func main() {
	// Definimos los flags para configurar el nodo desde la línea de comandos.
	// Esto nos permitirá lanzar varios nodos fácilmente.
	nodeID := flag.Int("id", 0, "ID del nodo (1, 2, 3, etc.)")
	flag.Parse()

	if *nodeID == 0 {
		log.Fatal("El argumento -id es obligatorio. Ej: go run main.go -id 1")
	}

	// Lista de todos los nodos en el sistema. En una aplicación real,
	// esto podría venir de un archivo de configuración o un servicio de descubrimiento.
	peers := map[int]string{
		1: "localhost:8081",
		2: "localhost:8082",
		3: "localhost:8083",
	}

	// (Próximamente) Crearemos la instancia del nodo aquí.
	// currentNode := node.NewNode(*nodeID, peers)
	// log.Printf("Nodo %d iniciado. Escuchando en %s", currentNode.ID, peers[*nodeID])

	// --- (Esto es solo un placeholder para ver que funciona) ---
	log.Printf("Iniciando Nodo %d...", *nodeID)
	log.Printf("Dirección: %s", peers[*nodeID])
	log.Println("Para detener el nodo, presione CTRL+C.")
	// -----------------------------------------------------------

	// Esperar una señal de interrupción (CTRL+C) para terminar de forma limpia.
	// Esto es útil para la simulación de fallos.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nApagando el nodo...")
	// Aquí iría la lógica de apagado limpio.
}