@echo off
SETLOCAL

REM --- Script para simular el inicio y detención de nodos en Windows ---

REM Navega a la carpeta donde está el main.go de los servidores.
cd cmd\server

IF /I "%1"=="start" (
    echo Iniciando Nodo %2...
    REM El comando start abre una nueva ventana. El primer "" es para el título de la ventana.
    start "Nodo %2" go run . -id %2
    echo Nodo %2 iniciado en una nueva ventana.

) ELSE IF /I "%1"=="kill" (
    echo Deteniendo Nodo %2...
    REM taskkill busca y detiene el proceso por el título de la ventana que le dimos.
    taskkill /FI "WINDOWTITLE eq Nodo %2" /F
    echo Nodo %2 detenido.

) ELSE (
    echo Uso:
    echo   simulate.bat start [id]
    echo   simulate.bat kill [id]
    echo.
    echo Ejemplos:
    echo   simulate.bat start 1
    echo   simulate.bat kill 3
)

ENDLOCAL