package coordination

import (
	"bytes"
	"encoding/json"
	"log"
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

type Bully struct {
	node *node.Node
}

func NewBully(n *node.Node) *Bully {
	return &Bully{node: n}
}

func (b *Bully) StartElection() {
	if !b.node.SetElectionInProgress(true) {
		log.Printf("Nodo %d: Ya hay una elección en curso, no se iniciará una nueva.", b.node.ID)
		return
	}
	defer b.node.SetElectionInProgress(false)

	log.Printf("Nodo %d: Iniciando elección. Buscando nodos con ID mayor.", b.node.ID)

	higherNodes := false

	higherNodeResponded := make(chan bool, 1)

	for peerID := range b.node.Peers {
		if peerID > b.node.ID {
			higherNodes = true
			go func(id int) {

				if b.sendElectionMessage(id) {
					higherNodeResponded <- true
				}
			}(peerID)
		}
	}

	if !higherNodes {
		log.Printf("Nodo %d: No hay nodos con ID mayor. Me declaro primario.", b.node.ID)
		b.announceVictory()
		return
	}

	select {
	case <-higherNodeResponded:

		log.Printf("Nodo %d: Un nodo con ID superior respondió. Deteniendo mi candidatura.", b.node.ID)
		return
	case <-time.After(3 * time.Second):

		log.Printf("Nodo %d: Timeout. Ningún nodo con ID superior respondió. Me declaro primario.", b.node.ID)
		b.announceVictory()
	}
}

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
		return true
	}
	return false
}

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

	time.Sleep(2 * time.Second)
	log.Printf("Nodo %d: Anunciando presencia y desafiando el liderazgo actual.", b.node.ID)

	b.StartElection()
}
