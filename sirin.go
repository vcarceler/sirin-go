package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type request struct {
	playbook string
	timestamp time.Time
}

var registered map[string]request
var address string
var port int
var secret string
var timeout string
var duration time.Duration

func getNumberOfRequests(w http.ResponseWriter, r *http.Request) {
	playbook := strings.TrimPrefix(r.URL.Path, "/getnumberofrequests/")
	addr, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("/getnumberofrequests/ playbook=%s remoteaddress=%s error with net.SplitHostPort()", playbook, r.RemoteAddr)
		return
	}

	message := fmt.Sprintf("/getnumberofrequests/ playbook=%s addr=%s port=%s", playbook, addr, port)

	count := 0
	for _, req := range registered {
		if req.playbook == playbook {
			count++
		}
	}

	message = fmt.Sprintf("%s count=%d", message, count)

	log.Printf(message)
	fmt.Fprintf(w, fmt.Sprintf("%d", count))
}

func listpendingrequests(w http.ResponseWriter, r *http.Request) {
	log.Printf("/listpendingrequests/ Total pending requests: %d", len(registered))

	salida := ""
	for host, req := range registered {
		salida = salida + fmt.Sprintf("%s %s %s\n", req.timestamp.Format("2006-01-02 15:04:05.00"), host, req.playbook)
	}

	fmt.Fprintf(w, salida)
}

func load(w http.ResponseWriter, r *http.Request) {

	// Cargo datos de prueba
	registered["10.0.0.1"] = request{"playbook1", time.Now()}
	registered["10.0.0.2"] = request{"playbook2", time.Now()}
	registered["10.0.0.3"] = request{"playbook3", time.Now()}

	fmt.Fprintf(w, "He cargado datos!")
}

func register(w http.ResponseWriter, r *http.Request) {
	playbook := strings.TrimPrefix(r.URL.Path, "/register/")
	addr, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("/register/ playbook=%s remoteaddress=%s error with net.SplitHostPort()", playbook, r.RemoteAddr)
		return
	}

	message := fmt.Sprintf("/register/ playbook=%s addr=%s port=%s", playbook, addr, port)

	now := time.Now()
	// Buscamos si el equipo está registrado
	req, ok := registered[addr]
	if ok == false {
		// No está registrado, lo añado
		message = message + " newrequest=true"
		registered[addr] = request{playbook, now}
	} else {
		// Estrá registrado, compruebo desde cuando
		message = message + " newrequest=false"
		elapsed := now.Sub(req.timestamp)
		message = fmt.Sprintf("%s elapsed=%s timeout=%s", message, elapsed, duration)
		
		// Comprobamos si ha pasado el timeout
		if elapsed > duration {
			// Actualizo la solicitud
			message = message + " updated"
			registered[addr] = request{playbook, now}
		} else {
			// Descarto la solicitud
			message = message + " discarded"
		}
	}

	log.Printf(message)
	fmt.Fprintf(w, message)
}

func main() {
	registered = make(map[string]request)

	flag.StringVar(&address, "address", "0.0.0.0", "Dirección para recibir peticiones")
	flag.IntVar(&port, "port", 8080, "Puerto")
	flag.StringVar(&secret, "secret", "SIRIN", "Token secreto")
	flag.StringVar(&timeout, "timeout", "23h", "Tiempo antes de registrar una nueva petición")

	flag.Parse()

	var err error
	duration, err = time.ParseDuration(timeout)
	if err != nil {
		log.Printf("Error en time.ParseDuration() de %s", timeout)
		log.Printf("timeout incorrecto")
		os.Exit(1)
	}

	log.Printf("sirin -address %s -port %d -secret %s -timeout %s", address, port, secret, timeout)
	
	http.HandleFunc("/getnumberofrequests/", getNumberOfRequests)
	http.HandleFunc("/listpendingrequests/", listpendingrequests)
	http.HandleFunc("/load/", load)
	http.HandleFunc("/register/", register)

	log.Fatal(http.ListenAndServe(":8080", nil))
}