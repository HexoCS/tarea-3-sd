#!/bin/bash
SERVER_DIR="../cmd/server"

PID_DIR="./pids"
mkdir -p $PID_DIR


start_node() {
    ID=$1
    PID_FILE="$PID_DIR/node_$ID.pid"
    if [ -f "$PID_FILE" ]; then
        echo "El Nodo $ID ya está en ejecución (PID $(cat $PID_FILE))."
        return
    fi

    echo "Iniciando Nodo $ID..."
    
    (cd "$SERVER_DIR" && go run . -id "$ID" > "$PID_DIR/node_$ID.log" 2>&1 &)
    
    echo $! > "$PID_FILE"
    echo "Nodo $ID iniciado con PID $(cat $PID_FILE). Log en $PID_DIR/node_$ID.log"
}


kill_node() {
    ID=$1
    PID_FILE="$PID_DIR/node_$ID.pid"
    if [ ! -f "$PID_FILE" ]; then
        echo "El Nodo $ID no está en ejecución (no se encontró .pid)."
        return
    fi

    PID=$(cat "$PID_FILE")
    echo "Deteniendo Nodo $ID (PID $PID)..."
    kill "$PID"
    
    sleep 1
    rm "$PID_FILE"
    echo "Nodo $ID detenido."
}


case "$1" in
    start)
        start_node "$2"
        ;;
    kill)
        kill_node "$2"
        ;;
    *)
        echo "Uso: $0 {start|kill} [id_del_nodo]"
        exit 1
        ;;