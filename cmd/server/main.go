package main

import (
	"flag"
	"fmt"
	"log"
	"mi-tarea-sd/internal/api"
	"mi-tarea-sd/internal/coordination"
	"mi-tarea-sd/internal/monitoring"
	"mi-tarea-sd/internal/node"
	"mi-tarea-sd/internal/synchronization" // Importamos el nuevo paquete
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 1. Configuración de Flags para la Línea de Comandos
	// Permite especificar el ID del nodo al ejecutar, ej: go run . -id 1
	nodeID := flag.Int("id", 0, "ID del nodo (1, 2, 3, etc.)")
	flag.Parse()

	if *nodeID == 0 {
		log.Fatal("El argumento -id es obligatorio. Ej: go run main.go -id 1")
	}

	// 2. Definición de los Nodos del Sistema (Peers)
	// En una aplicación real, esto podría leerse de un archivo de configuración.
	peers := map[int]string{
		1: "10.10.28.50:8081", // IP de la VM para el Nodo 1
		2: "10.10.28.51:8082", // IP de la VM para el Nodo 2
		3: "10.10.28.52:8083", // IP de la VM para el Nodo 3
	}	

	// 3. Creación de la Instancia del Nodo Principal
	// Este objeto contendrá el estado y la lógica central de nuestro nodo.
	currentNode := node.NewNode(*nodeID, peers)
	log.Printf("Nodo %d iniciado. Dirección: %s", currentNode.ID, peers[*nodeID])

	// 4. Inicio del Servidor API
	// Se ejecuta en una goroutine para no bloquear el hilo principal.
	// Escuchará las peticiones de otros nodos (heartbeat, elección, replicación, etc.).
	apiServer := api.NewServer(currentNode)
	go apiServer.Start()

	// 5. Inicio del Proceso de Monitoreo (Heartbeat)
	// Si este nodo es un secundario, se encargará de enviar heartbeats al primario.
	monitor := monitoring.NewMonitor(currentNode)
	go monitor.StartHeartbeatProcess()

	// 6. Anuncio de Presencia y Elección Inicial
	// Al arrancar, cada nodo inicia una elección para asegurar que el nodo con el ID
	// más alto siempre se convierta en el líder, manejando la reintegración.
	bully := coordination.NewBully(currentNode)
	go bully.AnnouncePresenceAndChallenge()

	// 7. Sincronización de Estado al Reintegrarse
	// Después de la elección, el nodo intenta sincronizar su estado con el del primario
	// para obtener los eventos que podría haber perdido mientras estaba desconectado.
	synchronizer := synchronization.NewSynchronizer(currentNode)
	go synchronizer.FetchStateFromPrimary()

	log.Println("Sistema distribuido en funcionamiento. Para detener el nodo, presione CTRL+C.")

	// 8. Espera de Señal de Apagado
	// El programa se bloquea aquí, permitiendo que todos los servicios en goroutines sigan
	// funcionando hasta que se presione CTRL+C para un apagado controlado.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("\nRecibida señal de apagado. Terminando el nodo...")
	// Aquí podría ir lógica de limpieza si fuera necesario.
}