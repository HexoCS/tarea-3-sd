package coordination

import (
	"bytes"
	"encoding/json"
	"log"
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

// Bully se encarga de la lógica de elección.
type Bully struct {
	node *node.Node
}

// NewBully crea una instancia del coordinador del algoritmo del matón.
func NewBully(n *node.Node) *Bully {
	return &Bully{node: n}
}

// StartElection es llamada cuando un nodo detecta la caída del primario.
func (b *Bully) StartElection() {
	if !b.node.SetElectionInProgress(true) {
		log.Printf("Nodo %d: Ya hay una elección en curso, no se iniciará una nueva.", b.node.ID)
		return
	}
	defer b.node.SetElectionInProgress(false)

	log.Printf("Nodo %d: Iniciando elección. Buscando nodos con ID mayor.", b.node.ID)

	higherNodes := false
	// Usamos un canal para saber si al menos un nodo superior respondió.
	higherNodeResponded := make(chan bool, 1)

	for peerID := range b.node.Peers {
		if peerID > b.node.ID {
			higherNodes = true
			go func(id int) {
				// Si el nodo superior responde, enviamos 'true' al canal.
				if b.sendElectionMessage(id) {
					higherNodeResponded <- true
				}
			}(peerID)
		}
	}

	// Si no existen nodos con ID superior, este nodo gana la elección inmediatamente.
	if !higherNodes {
		log.Printf("Nodo %d: No hay nodos con ID mayor. Me declaro primario.", b.node.ID)
		b.announceVictory()
		return
	}

	// Usamos select para esperar una respuesta o un timeout.
	select {
	case <-higherNodeResponded:
		// Un nodo superior respondió. Este nodo no hace nada más.
		log.Printf("Nodo %d: Un nodo con ID superior respondió. Deteniendo mi candidatura.", b.node.ID)
		return
	case <-time.After(3 * time.Second):
		// Nadie respondió en el tiempo límite. Este nodo gana la elección.
		log.Printf("Nodo %d: Timeout. Ningún nodo con ID superior respondió. Me declaro primario.", b.node.ID)
		b.announceVictory()
	}
}

// sendElectionMessage envía un mensaje de "ELECCIÓN" y devuelve true si recibe un OK.
func (b *Bully) sendElectionMessage(peerID int) bool {
	url := "http://" + b.node.Peers[peerID] + "/election"
	log.Printf("Nodo %d: Enviando mensaje de ELECCIÓN a Nodo %d en %s", b.node.ID, peerID, url)

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Nodo %d: Nodo %d no respondió al mensaje de elección (probablemente caído).", b.node.ID, peerID)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("Nodo %d: Recibí OK de Nodo %d. Dejo que él continúe la elección.", b.node.ID, peerID)
		return true // El nodo superior está activo y tomará el control.
	}
	return false
}

// announceVictory el nodo se declara a sí mismo como el nuevo primario y lo anuncia.
func (b *Bully) announceVictory() {
	b.node.SetPrimary(true)
	b.node.SetPrimaryID(b.node.ID)

	log.Printf("Nodo %d: ¡VICTORIA! Soy el nuevo primario. Anunciando a los demás.", b.node.ID)

	for peerID := range b.node.Peers {
		if peerID != b.node.ID {
			go b.sendCoordinatorMessage(peerID)
		}
	}
}

// sendCoordinatorMessage envía un mensaje de "COORDINADOR" para anunciar al nuevo líder.
func (b *Bully) sendCoordinatorMessage(peerID int) {
	url := "http://" + b.node.Peers[peerID] + "/coordinator"
	message := map[string]int{"primary_id": b.node.ID}
	jsonData, _ := json.Marshal(message)

	client := http.Client{Timeout: 2 * time.Second}
	_, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Nodo %d: No se pudo anunciar como coordinador a Nodo %d.", b.node.ID, peerID)
	}
}

func (b *Bully) AnnouncePresenceAndChallenge() {
	// Dar un pequeño margen para que el servidor API se levante completamente.
	time.Sleep(2 * time.Second)
	log.Printf("Nodo %d: Anunciando presencia y desafiando el liderazgo actual.", b.node.ID)
	
	// La forma más simple de asegurar que el nodo correcto sea el líder
	// es simplemente iniciar una elección. El "matón" siempre ganará.
	b.StartElection()
}
	