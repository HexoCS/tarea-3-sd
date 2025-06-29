package node

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type Event struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

type State struct {
	SequenceNumber int     `json:"sequence_number"`
	EventLog       []Event `json:"event_log"`
}

type Node struct {
	ID                 int
	IsPrimary          bool
	PrimaryID          int
	Peers              map[int]string
	State              State
	mutex              sync.RWMutex
	logFile            string
	IsActive           bool
	electionInProgress bool
}

func NewNode(id int, peers map[int]string) *Node {
	n := &Node{
		ID:    id,
		Peers: peers,
		State: State{
			SequenceNumber: 0,
			EventLog:       make([]Event, 0),
		},
		logFile:            "nodo.json",
		IsActive:           true,
		electionInProgress: false,
	}

	highestID := 0
	for peerID := range n.Peers {
		if peerID > highestID {
			highestID = peerID
		}
	}
	n.SetPrimaryID(highestID)
	if n.ID == highestID {
		n.SetPrimary(true)
	}

	n.loadState()
	return n
}

func (n *Node) AddEvent(eventValue string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.State.SequenceNumber++

	newEvent := Event{
		ID:    n.State.SequenceNumber,
		Value: eventValue,
	}

	n.State.EventLog = append(n.State.EventLog, newEvent)

	log.Printf("Nodo %d: Evento añadido (Seq: %d, Valor: '%s'). Estado actualizado.", n.ID, n.State.SequenceNumber, eventValue)

	n.saveStateInternal()
}

func (n *Node) GetPrimaryID() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.PrimaryID
}

func (n *Node) SetPrimaryID(primaryID int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.PrimaryID = primaryID
	log.Printf("Nodo %d: El nuevo primario reconocido es el Nodo %d.", n.ID, primaryID)
}

func (n *Node) SetElectionInProgress(status bool) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if n.electionInProgress && status {
		return false
	}
	n.electionInProgress = status
	if status {
		log.Printf("Nodo %d: Estado de elección cambiado a: EN CURSO.", n.ID)
	} else {
		log.Printf("Nodo %d: Estado de elección cambiado a: FINALIZADA.", n.ID)
	}
	return true
}

func (n *Node) loadState() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	data, err := ioutil.ReadFile(n.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Nodo %d: No se encontró el archivo de log '%s'. Se creará uno nuevo.", n.ID, n.logFile)
			n.saveStateInternal()
			return
		}
		log.Fatalf("Nodo %d: Error al leer el archivo de log: %v", n.ID, err)
	}

	if err := json.Unmarshal(data, &n.State); err != nil {
		log.Fatalf("Nodo %d: Error al decodificar el estado del log: %v", n.ID, err)
	}
	log.Printf("Nodo %d: Estado cargado desde '%s'. Número de secuencia actual: %d", n.ID, n.logFile, n.State.SequenceNumber)
}

func (n *Node) saveState() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.saveStateInternal()
}

func (n *Node) saveStateInternal() {
	data, err := json.MarshalIndent(n.State, "", "  ")
	if err != nil {
		log.Fatalf("Nodo %d: Error al codificar el estado a JSON: %v", n.ID, err)
	}

	if err := ioutil.WriteFile(n.logFile, data, 0644); err != nil {
		log.Fatalf("Nodo %d: Error al escribir en el archivo de log: %v", n.ID, err)
	}
}

func (n *Node) SetPrimary(isPrimary bool) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.IsPrimary = isPrimary
	if isPrimary {
		n.PrimaryID = n.ID
	}
	log.Printf("Nodo %d: Ahora es %s.", n.ID, map[bool]string{true: "PRIMARIO", false: "SECUNDARIO"}[isPrimary])
}

func (n *Node) SetState(newState State) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.State = newState
	log.Printf("Nodo %d: Estado sobrescrito por el primario. Nuevo Seq: %d.", n.ID, n.State.SequenceNumber)

	n.saveStateInternal()
}
