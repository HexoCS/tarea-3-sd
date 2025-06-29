package monitoring

import (
	"log"
	"mi-tarea-sd/internal/coordination"
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

type Monitor struct {
	node *node.Node
}

func NewMonitor(n *node.Node) *Monitor {
	return &Monitor{node: n}
}

func (m *Monitor) StartHeartbeatProcess() {
	time.Sleep(5 * time.Second)
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	for {

		if !m.node.IsPrimary && m.node.IsActive {
			m.sendHeartbeat(client)
		}

		time.Sleep(3 * time.Second)
	}
}

func (m *Monitor) sendHeartbeat(client *http.Client) {
	primaryID := m.node.GetPrimaryID()
	if primaryID == 0 {
		log.Printf("Nodo %d: No hay un primario conocido. Saltando heartbeat e iniciando elección.", m.node.ID)

		bully := coordination.NewBully(m.node)
		go bully.StartElection()
		return
	}

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

		log.Printf("Nodo %d: ERROR - El primario (Nodo %d) no responde. %v", m.node.ID, primaryID, err)

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
