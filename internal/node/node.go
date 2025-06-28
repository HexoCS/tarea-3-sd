package node

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// Event representa una entrada individual en el log de eventos replicado.
// Coincide con la estructura pedida en el enunciado.
type Event struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

// State define la estructura del estado replicado.
// Ahora incluye un log de eventos además del número de secuencia.
type State struct {
	SequenceNumber int     `json:"sequence_number"`
	EventLog       []Event `json:"event_log"`
}

// Node representa un servidor en el sistema distribuido.
type Node struct {
	ID                 int
	IsPrimary          bool
	PrimaryID          int
	Peers              map[int]string
	State              State // <-- El estado ahora tiene la nueva estructura
	mutex              sync.RWMutex
	logFile            string
	IsActive           bool
	electionInProgress bool
}

// NewNode crea y retorna una nueva instancia de Node.
func NewNode(id int, peers map[int]string) *Node {
	n := &Node{
		ID:    id,
		Peers: peers,
		State: State{ // <-- Inicializamos el estado con un log vacío
			SequenceNumber: 0,
			EventLog:       make([]Event, 0),
		},
		logFile:            "nodo.json",
		IsActive:           true,
		electionInProgress: false,
	}

	// Lógica de elección inicial (sin cambios)
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

// AddEvent es la nueva versión de IncrementSequence.
// Añade un nuevo evento al log, incrementa el número de secuencia y guarda el estado.
// Esta función será llamada por el primario.
func (n *Node) AddEvent(eventValue string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	// 1. Incrementar el número de secuencia.
	n.State.SequenceNumber++

	// 2. Crear el nuevo evento.
	newEvent := Event{
		ID:    n.State.SequenceNumber, // El ID del evento es el nuevo número de secuencia.
		Value: eventValue,
	}

	// 3. Añadir el evento al log.
	n.State.EventLog = append(n.State.EventLog, newEvent)

	log.Printf("Nodo %d: Evento añadido (Seq: %d, Valor: '%s'). Estado actualizado.", n.ID, n.State.SequenceNumber, eventValue)

	// 4. Persistir el nuevo estado en el archivo.
	n.saveStateInternal()
}

// --- El resto de las funciones no necesitan cambios significativos ---

// GetPrimaryID retorna el ID del primario actual de forma segura.
func (n *Node) GetPrimaryID() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return n.PrimaryID
}

// SetPrimaryID establece el ID del primario de forma segura.
func (n *Node) SetPrimaryID(primaryID int) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.PrimaryID = primaryID
	log.Printf("Nodo %d: El nuevo primario reconocido es el Nodo %d.", n.ID, primaryID)
}

// SetElectionInProgress controla el estado de la elección de forma segura.
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

// loadState carga el estado del nodo desde su archivo de log JSON.
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

// saveState guarda el estado actual del nodo en el archivo de log JSON.
func (n *Node) saveState() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.saveStateInternal()
}

// saveStateInternal es una versión no bloqueante para ser llamada desde otras funciones que ya tienen el lock.
func (n *Node) saveStateInternal() {
	data, err := json.MarshalIndent(n.State, "", "  ")
	if err != nil {
		log.Fatalf("Nodo %d: Error al codificar el estado a JSON: %v", n.ID, err)
	}

	if err := ioutil.WriteFile(n.logFile, data, 0644); err != nil {
		log.Fatalf("Nodo %d: Error al escribir en el archivo de log: %v", n.ID, err)
	}
}

// SetPrimary establece si el nodo es primario o no.
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

	// Persistir el estado recién actualizado.
	n.saveStateInternal()
}