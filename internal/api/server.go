package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"mi-tarea-sd/internal/coordination"
	"mi-tarea-sd/internal/node"
	"net/http"
	"time"
)

type Server struct {
	node   *node.Node
	router *http.ServeMux
	bully  *coordination.Bully
}

func NewServer(n *node.Node) *Server {
	s := &Server{
		node:   n,
		router: http.NewServeMux(),
		bully:  coordination.NewBully(n),
	}
	s.registerHandlers()
	return s
}

func (s *Server) registerHandlers() {
	s.router.HandleFunc("/heartbeat", s.handleHeartbeat)
	s.router.HandleFunc("/election", s.handleElection)
	s.router.HandleFunc("/coordinator", s.handleCoordinator)
	s.router.HandleFunc("/event", s.handleEvent)
	s.router.HandleFunc("/state-update", s.handleStateUpdate)
	s.router.HandleFunc("/state", s.handleGetState)
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	if !s.node.IsPrimary {

		http.Error(w, "No soy el nodo primario.", http.StatusServiceUnavailable)
		return
	}

	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Cuerpo de la petición inválido.", http.StatusBadRequest)
		return
	}
	eventValue, ok := payload["value"]
	if !ok {
		http.Error(w, "El campo 'value' es requerido.", http.StatusBadRequest)
		return
	}

	log.Printf("Nodo Primario %d: Recibido nuevo evento con valor '%s'", s.node.ID, eventValue)

	s.node.AddEvent(eventValue)

	go s.broadcastStateUpdate()

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Evento recibido y replicado.")
}

func (s *Server) handleStateUpdate(w http.ResponseWriter, r *http.Request) {
	if s.node.IsPrimary {

		http.Error(w, "Soy el primario, no puedo recibir actualizaciones de estado.", http.StatusBadRequest)
		return
	}

	var newState node.State
	if err := json.NewDecoder(r.Body).Decode(&newState); err != nil {
		http.Error(w, "Cuerpo de la petición de estado inválido.", http.StatusBadRequest)
		return
	}

	s.node.SetState(newState)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) broadcastStateUpdate() {
	currentState := s.node.State
	jsonData, err := json.Marshal(currentState)
	if err != nil {
		log.Printf("Nodo Primario %d: Error al codificar el estado para broadcast: %v", s.node.ID, err)
		return
	}

	for peerID, peerAddress := range s.node.Peers {
		if peerID == s.node.ID {
			continue
		}

		go func(id int, addr string) {
			url := "http://" + addr + "/state-update"
			log.Printf("Nodo Primario %d: Replicando estado a Nodo %d en %s", s.node.ID, id, url)

			client := http.Client{Timeout: 2 * time.Second}
			resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Printf("Nodo Primario %d: Falló la replicación a Nodo %d: %v", s.node.ID, id, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("Nodo Primario %d: Nodo %d respondió con error a la replicación: %s", s.node.ID, id, resp.Status)
			}
		}(peerID, peerAddress)
	}
}

func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	log.Printf("Nodo %d: Heartbeat recibido", s.node.ID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "ACK")
}

func (s *Server) handleElection(w http.ResponseWriter, r *http.Request) {
	log.Printf("Nodo %d: Mensaje de ELECCIÓN recibido.", s.node.ID)
	w.WriteHeader(http.StatusOK)
	go s.bully.StartElection()
}

func (s *Server) handleCoordinator(w http.ResponseWriter, r *http.Request) {
	var payload map[string]int
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	primaryID := payload["primary_id"]
	log.Printf("Nodo %d: Mensaje de COORDINADOR recibido. Nuevo primario es Nodo %d.", s.node.ID, primaryID)
	s.node.SetPrimaryID(primaryID)
	s.node.SetPrimary(s.node.ID == primaryID)
	s.node.SetElectionInProgress(false)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Start() {
	address := s.node.Peers[s.node.ID]
	log.Printf("Nodo %d: Servidor API escuchando en http://%s", s.node.ID, address)
	if err := http.ListenAndServe(address, s.router); err != nil {
		log.Fatalf("Nodo %d: El servidor falló al iniciar: %v", s.node.ID, err)
	}
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	if !s.node.IsPrimary {
		http.Error(w, "No soy el nodo primario.", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.node.State); err != nil {
		log.Printf("Nodo Primario %d: Error al codificar estado para enviar: %v", s.node.ID, err)
		http.Error(w, "Error interno del servidor.", http.StatusInternalServerError)
	}
}
