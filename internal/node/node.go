package node

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// State define la estructura del estado replicado que se guarda en el log.
type State struct {
	SequenceNumber int `json:"sequence_number"`
	// Se podrían agregar más campos aquí, como un log de eventos.
}

// Node representa un servidor en el sistema distribuido.
type Node struct {
	ID         int
	IsPrimary  bool
	Peers      map[int]string // Mapea ID de nodo a su dirección (ej: "localhost:8081")
	State      State
	mutex      sync.Mutex // Protege el acceso concurrente al estado del nodo
	logFile    string
	IsActive   bool // Indica si el nodo está participando activamente
}

// NewNode crea y retorna una nueva instancia de Node.
func NewNode(id int, peers map[int]string) *Node {
	n := &Node{
		ID:        id,
		IsPrimary: false, // Inicialmente, ningún nodo se asume como primario
		Peers:     peers,
		State:     State{SequenceNumber: 0},
		logFile:   "nodo.json", // Nombre del archivo de log según la especificación 
		IsActive:  true,
	}
	n.loadState() // Carga el estado desde el archivo al iniciar
	return n
}

// loadState carga el estado del nodo desde su archivo de log JSON.
func (n *Node) loadState() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	data, err := ioutil.ReadFile(n.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Nodo %d: No se encontró el archivo de log '%s'. Se creará uno nuevo.", n.ID, n.logFile)
			n.saveState() // Crea el archivo si no existe
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
// Esta función es clave para la persistencia. 
func (n *Node) saveState() {
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
	log.Printf("Nodo %d: Ahora es %s.", n.ID, map[bool]string{true: "PRIMARIO", false: "SECUNDARIO"}[isPrimary])
}

// IncrementSequence actualiza el número de secuencia y persiste el estado.
// Esta función será llamada por el primario al coordinar un nuevo evento. 
func (n *Node) IncrementSequence() {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.State.SequenceNumber++
	log.Printf("Nodo %d: Número de secuencia incrementado a %d.", n.ID, n.State.SequenceNumber)
	n.saveState()
}