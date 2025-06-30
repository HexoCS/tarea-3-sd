# Tarea 3


## Integrantes del Grupo

| Nombre y Apellido | Rol |
| :---------------- | :--------------------------------------------- |
| `Bruno Bonati` | `202173057-5` |
| `Joaquin Veliz` | `202004065-6` |
| `Joaquin Torres` | `202073112-8` |

---

### 1. Ubicación de los Componentes

| Componente | Máquina Virtual |
| :--------- | :-------------- |
| Nodo 1     | `[MV1]`     |
| Nodo 2     | `[MV2]`     |
| Nodo 3     | `[MV3]`     |

### 2. Script de Simulación

Para gestionar los nodos, se utiliza el script `scripts/simulate.sh`.

* **Dar permisos de ejecución (solo una vez):**
  Navega a la raíz del proyecto y ejecuta:
  ```bash
  chmod +x scripts/simulate.sh
  ```

* **Iniciar un nodo (ej. Nodo 1):**
  ```bash
  ./scripts/simulate.sh start 1
  ```
  Repite este comando para los nodos 2 y 3 en sus respectivas VMs.

* **Detener un nodo (ej. Nodo 3):**
  ```bash
  ./scripts/simulate.sh kill 3
  ```

* **Ver el log de un nodo en tiempo real (ej. Nodo 2):**
  Cada nodo guarda su log en `scripts/pids/node_X.log`. Para monitorearlo:
  ```bash
  tail -f scripts/pids/node_2.log
  ```

---

## Cómo Probar las Funcionalidades

A continuación se detallan los pasos para demostrar cada una de las funcionalidades clave del sistema.

### A. Simular un Nuevo Evento (Replicación)

1.  Inicia los 3 nodos. El Nodo 3 será elegido como líder.
2.  Para enviar un nuevo evento, ejecuta una petición `POST` al endpoint `/event` del **nodo primario (IP 10.10.28.52, puerto 8083)**:
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"value": "Primer evento del sistema"}' [http://10.10.28.52:8083/event](http://10.10.28.52:8083/event)
    ```
3.  **Verificación**: Observa los logs de los 3 nodos. Verás el evento replicado. Adicionalmente, los archivos `cmd/server/nodo.json` en cada instancia del proyecto serán idénticos.


### B. Simular Fallo y Elección de Líder

1.  Con los 3 nodos en funcionamiento, detén el nodo primario (Nodo 3):
    ```bash
    ./scripts/simulate.sh kill 3
    ```
2.  **Verificación**: Observa los logs de los Nodos 1 y 2. Detectarán la falla e iniciarán una elección. El Nodo 2 (el de mayor ID disponible) ganará y se anunciará como el nuevo primario.

### C. Simular Reintegración y Sincronización

1.  Continuando con el escenario anterior, el Nodo 2 es el líder. Envía un nuevo evento que el Nodo 3 se perderá, esta vez apuntando a la **IP del Nodo 2 (10.10.28.51, puerto 8082)**:
    ```bash
    curl -X POST -H "Content-Type: application/json" -d '{"value": "Evento durante la ausencia del nodo 3"}' [http://10.10.28.51:8082/event](http://10.10.28.51:8082/event) 
    ```
2.  Reinicia el Nodo 3 que estaba caído:
    ```bash
    ./scripts/simulate.sh start 3
    ```
3.  **Verificación**: Observa los logs del Nodo 3.
    * Iniciará una elección y **reclamará su rol como primario** por tener el ID más alto.
    * Solicitará el estado completo al clúster para ponerse al día.
    * Finalmente, su archivo `cmd/server/nodo.json` será idéntico al de los otros nodos, incluyendo el evento que se perdió.


---

## Consideraciones Especiales

* La persistencia de cada nodo se encuentra en el archivo `cmd/server/nodo.json`. Borrar este archivo simula el arranque de un nodo desde un estado completamente nuevo.
* El sistema está configurado para 3 nodos con puertos 8081, 8082 y 8083. Estos valores están definidos en `cmd/server/main.go`.
