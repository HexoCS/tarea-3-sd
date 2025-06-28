package synchronization

import (
	"encoding/json"
	"log"
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

// Synchronizer se encarga de que un nodo obtenga el estado más reciente del primario.
type Synchronizer struct {
	node *node.Node
}

// NewSynchronizer crea una nueva instancia de Synchronizer.
func NewSynchronizer(n *node.Node) *Synchronizer {
	return &Synchronizer{node: n}
}

// FetchStateFromPrimary se ejecuta al arrancar un nodo (después de la elección inicial).
// Contacta al primario para obtener el estado más reciente.
func (s *Synchronizer) FetchStateFromPrimary() {
	// Damos un tiempo prudencial para que la elección de líder se estabilice.
	time.Sleep(5 * time.Second)

	if s.node.IsPrimary {
		// El primario no necesita sincronizarse, ya que tiene la verdad.
		return
	}

	primaryID := s.node.GetPrimaryID()
	if primaryID == 0 {
		log.Printf("Nodo %d: No se pudo sincronizar, no se conoce un primario.", s.node.ID)
		return
	}

	primaryAddress := s.node.Peers[primaryID]
	url := "http://" + primaryAddress + "/state"
	log.Printf("Nodo %d: Solicitando estado actual al primario (Nodo %d) en %s", s.node.ID, primaryID, url)

	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Nodo %d: Error al solicitar estado al primario: %v", s.node.ID, err)
		return
	}
	defer resp.Body.Close()

	var stateFromServer node.State
	if err := json.NewDecoder(resp.Body).Decode(&stateFromServer); err != nil {
		log.Printf("Nodo %d: Error al decodificar el estado del primario: %v", s.node.ID, err)
		return
	}

	// Actualizamos el estado local con la versión del primario.
	s.node.SetState(stateFromServer)
	log.Printf("Nodo %d: Sincronización completada. Estado actualizado desde el primario.", s.node.ID)
}