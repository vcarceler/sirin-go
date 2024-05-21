/* This program is free software: you can redistribute it and/or modify it under the 
terms of the GNU General Public License as published by the Free Software 
Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY 
WARRANTY; without even the implied warranty of MERCHANTABILITY or 
FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for 
more details.

You should have received a copy of the GNU General Public License along with this 
program. If not, see <https://www.gnu.org/licenses/>.  */

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
	pending bool
}

var registered map[string]*request
var address string
var port int
var secret string
var timeout string
var duration time.Duration

func getHosts(w http.ResponseWriter, r *http.Request) {
	sec := r.PathValue("secret")
	pb := r.PathValue("playbook")

	if sec != secret {
		log.Printf("/gethosts/%s/%s playbook=%s remoteaddress=%s error='bad secret'", sec, pb, pb, r.RemoteAddr)
		return
	}

	addr, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("/gethosts/ playbook=%s remoteaddress=%s error with net.SplitHostPort()", pb, r.RemoteAddr)
		return
	}

	message := fmt.Sprintf("/gethosts/ playbook=%s addr=%s port=%s", pb, addr, port)

	count := 0
	out := ""
	for host, req := range registered {
		if req.playbook == pb && req.pending == true {
			count++
			out = out + host + ","
			req.pending = false
		}
	}

	message = fmt.Sprintf("%s count=%d hosts=%s", message, count, out)

	log.Printf(message)
	fmt.Fprintf(w, fmt.Sprintf("%s", out))
}

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
		if req.playbook == playbook && req.pending == true {
			count++
		}
	}

	message = fmt.Sprintf("%s count=%d", message, count)

	log.Printf(message)
	fmt.Fprintf(w, fmt.Sprintf("%d", count))
}

func listPendingRequests(w http.ResponseWriter, r *http.Request) {
	_, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("/listpendingrequests/ remoteaddress=%s error with net.SplitHostPort()", r.RemoteAddr)
		return
	}

	log.Printf("/listpendingrequests/ remoteaddress=%s count=%d",r.RemoteAddr, len(registered))

	salida := ""
	for host, req := range registered {
		if req.pending == true {
			salida = salida + fmt.Sprintf("%s %s %s\n", req.timestamp.Format("2006-01-02 15:04:05.00"), host, req.playbook)
		}
	}

	fmt.Fprintf(w, salida)
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
		registered[addr] = &request{playbook, now, true}
	} else {
		// Está registrado, compruebo desde cuando
		message = message + " newrequest=false"
		elapsed := now.Sub(req.timestamp)
		message = fmt.Sprintf("%s elapsed=%s timeout=%s", message, elapsed, duration)
		
		// Comprobamos si ha pasado el timeout
		if elapsed > duration {
			// Actualizo la solicitud
			message = message + " updated"
			registered[addr] = &request{playbook, now, true}
		} else {
			// Descarto la solicitud
			message = message + " discarded"
		}
	}

	log.Printf(message)
	fmt.Fprintf(w, message)
}

func main() {
	registered = make(map[string]*request)

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

	log.Printf("sirin -address %s -port %d -secret XXXX -timeout %s", address, port, timeout)
	
	http.HandleFunc("/gethosts/{secret}/{playbook}", getHosts)
	http.HandleFunc("/getnumberofrequests/", getNumberOfRequests)
	http.HandleFunc("/listpendingrequests/", listPendingRequests)
	http.HandleFunc("/register/", register)

	log.Fatal(http.ListenAndServe(":8080", nil))
}