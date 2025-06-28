package monitoring

import (
	"log"
	"mi-tarea-sd/internal/coordination" // Importamos el paquete de coordinación
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

// Monitor se encarga de las tareas de monitoreo, como enviar heartbeats.
type Monitor struct {
	node *node.Node
}

// NewMonitor crea una nueva instancia de Monitor.
func NewMonitor(n *node.Node) *Monitor {
	return &Monitor{node: n}
}

// StartHeartbeatProcess inicia el proceso de envío de heartbeats si el nodo no es el primario.
func (m *Monitor) StartHeartbeatProcess() {
	// Damos un pequeño tiempo para que todos los nodos se inicien antes de empezar el monitoreo.
	time.Sleep(5 * time.Second)

	// Creamos un cliente HTTP con un timeout.
	// Si el primario no responde en 2 segundos, se considera una falla.
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	for {
		// El proceso de heartbeat se ejecuta solo si el nodo es secundario y está activo.
		if !m.node.IsPrimary && m.node.IsActive {
			m.sendHeartbeat(client)
		}
		// Esperamos un tiempo antes del próximo heartbeat.
		time.Sleep(3 * time.Second)
	}
}

// sendHeartbeat envía un mensaje GET al endpoint /heartbeat del primario.
func (m *Monitor) sendHeartbeat(client *http.Client) {
	primaryID := m.node.GetPrimaryID()
	if primaryID == 0 {
		log.Printf("Nodo %d: No hay un primario conocido. Saltando heartbeat e iniciando elección.", m.node.ID)
		// Si no hay primario, se inicia una elección para establecer uno.
		bully := coordination.NewBully(m.node)
		go bully.StartElection()
		return
	}

	// Si por alguna razón este nodo cree que él mismo es el primario, no hace nada.
	if m.node.ID == primaryID {
		return
	}

	primaryAddress, ok := m.node.Peers[primaryID]
	if !ok {
		log.Printf("Nodo %d: No se encontró la dirección del primario con ID %d", m.node.ID, primaryID)
		return
	}

	url := "http://" + primaryAddress + "/heartbeat"
	log.Printf("Nodo %d: Enviando heartbeat a Primario (Nodo %d) en %s", m.node.ID, primaryID, url)

	resp, err := client.Get(url)
	if err != nil {
		// Si hay un error (timeout, conexión rechazada), asumimos que el primario ha caído.
		log.Printf("Nodo %d: ERROR - El primario (Nodo %d) no responde. %v", m.node.ID, primaryID, err)

		// --- INICIO DE ELECCIÓN ---
		// Se crea una instancia de Bully y se comienza la elección en una goroutine
		// para no bloquear el proceso de monitoreo.
		bully := coordination.NewBully(m.node)
		go bully.StartElection()

		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Nodo %d: Heartbeat exitoso. El primario (Nodo %d) está activo.", m.node.ID, primaryID)
	} else {
		log.Printf("Nodo %d: El primario (Nodo %d) respondió con estado: %s", m.node.ID, primaryID, resp.Status)
	}
}