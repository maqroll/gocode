package main

import (
  	"strconv"
	"os"
	"io"
	"bufio"
	"strings"
  	"net"
	"fmt"
	"flag"
	"time"
	"container/ring"
	"runtime"
    "log"
    "math/rand"
)

//NOTA: para analizar los tiempos de respuesta lo mejor es registrar los tiempos en un fichero de log
// y procesarlo a posteriori.
var servidor = flag.String("servidor", "dcdcorejees1", "servidor metropolis")
var puerto = flag.Int("puerto", 15345, "puerto del conector of2k")
var ficheroRq = flag.String("rq", "tira.txt", "nombre del fichero que contiene las peticiones")
var procesoLlegada = flag.String("kind","D", "proceso de llegadas según la notación de Kendall (M|D)")
var zero uint64 = 0
var nroMaxPeticiones *uint64 = &zero /*flag.Uint64("peticiones", 0, "numero de peticiones")*/
var lambda = flag.Float64("lambda", 10, "tasa de generacion en transacciones/s")
var C = flag.Int("C", 2, "numero de servidores en la notacion de Kendall")
var tenThousand = 10000
var k = &tenThousand
//var k = flag.Int("k",10000, "capacidad de la cola de espera(K-C en la notacion de Kendall)")

var datos *ring.Ring = nil /* Apunta al último valor insertado. Inicialmente datos.Len() = 0 */
var runForever bool = false
var contadorPeticiones uint64 = 0

type RequestData struct {
	StartTime int64
	EndTime   int64
}

var peticiones = make(chan []byte, 1000)
var resultados = make(chan *RequestData, 1000)
var fin = make(chan int)

func generadorTPS(nroMaxPeticiones uint64, lambda float64, q chan<- []byte, contadorPeticiones *uint64, forever bool, tiras *ring.Ring, procesoLlegada string) {
    if procesoLlegada == "M" {
        for ; forever || *contadorPeticiones < nroMaxPeticiones ; *contadorPeticiones++ {
            time.Sleep(time.Duration(int64(1000000000.0 * rand.ExpFloat64()/lambda)))
            q <- (tiras.Value.([]byte))
            tiras = tiras.Next()
        }
    } else { /* procesoLlegada == "D" */
        // Utilizo un Ticker como referencia.
        // Lo encapsulo en generadorTPS porque el Ticker utiliza un canal sincrono.
        // aunque el ticker nunca se bloquea
        ticker := time.NewTicker(time.Duration(int64(1000000000.0 / lambda)))

        for ; forever || *contadorPeticiones < nroMaxPeticiones ; *contadorPeticiones++ {
            <-ticker.C
            q <- (tiras.Value.([]byte))
            tiras = tiras.Next()
        }

        ticker.Stop()
    }
	close(q)
}

func visualizador(resultados <-chan *RequestData) {
	aspas := [...]byte{'-', '\\', '|', '/', '-'}
	count_aspas := 0
	var contadorPeticiones uint64 = 0
	var lastN = make(map[int]*ring.Ring, 3)
	lastN[5] = nil
	lastN[50] = nil
	lastN[500] = nil // Se mantiene por generalidad
	var tpsLastN = make(map[int]float32, 3)

	fmt.Printf("\r  %6s %6s %6s %6s %5s\n", "served", "-5", "-50", "-500", "queue")

	for r := range resultados {
		contadorPeticiones++
		if datos.Len() == 0 {
			datos = new(ring.Ring)
		} else if datos.Len() == 500 {
			datos = datos.Next()
		} else {
			datos.Link(new(ring.Ring))
			datos = datos.Next()
		}

		datos.Value = *r

		for i, anillo := range lastN {
			if anillo != nil {
				lastN[i] = anillo.Next()
			}

			if datos.Len() == i {
				lastN[i] = datos.Next()
			}

			if lastN[i] != nil {
				tpsLastN[i] = (float32(i) * 1.0e9) / float32(r.EndTime-lastN[i].Value.(RequestData).StartTime)
			} else {
				tpsLastN[i] = 0.0
			}
		}

		fmt.Printf("\r%s %6d %6.2f %6.2f %6.2f %5d                   ", string(aspas[count_aspas]), contadorPeticiones, tpsLastN[5], tpsLastN[50], tpsLastN[500], len(peticiones))
		count_aspas = (count_aspas + 1) % len(aspas)
	}
}

func notifyEnd(fin chan<- int, resultados chan<- *RequestData) {
	//close(resultados);
	fin <- 1
}

func generadorPeticiones(q <-chan []byte, resultados chan<- *RequestData, fin chan<- int) {
	defer notifyEnd(fin, resultados)
    var tira []byte
	for _ = range q {
		datosPeticion := RequestData{time.Now().UnixNano(), 0 /* inicialmente a cero*/ }
		con, errConexion := net.Dial("tcp", *servidor+":"+strconv.Itoa(*puerto))
		if errConexion != nil {
			log.Fatalf("%s\n",errConexion.Error())
		}
		sEscritura := tira
		var err error = nil
		var count int = 0

		for len(sEscritura) > 0 && err == nil {
			count, err = con.Write(sEscritura)
			sEscritura = sEscritura[count:len(sEscritura)]
		}
		if err != nil {
			log.Fatalf("%s\n",err.Error())
		}
		var respB = make([]byte, 1024)
		leidos := 0
		count = 1

		for count > 0 && err == nil {
			count, err = con.Read(respB[leidos:len(respB)])
			leidos = leidos + count
			if leidos == len(respB) {
				newSlice := make([]byte, len(respB)+1024)
				copy(respB, newSlice)
				respB = newSlice
			}
		}
		if err != nil && err != io.EOF {
			log.Fatalf("%s\n",err.Error())
		}

		datosPeticion.EndTime = time.Now().UnixNano()

		resultados <- &datosPeticion
        con.Close();
	}
}

func main() {
	flag.Parse()
	peticiones = make(chan []byte, *k)
	resultados = make(chan *RequestData, *k)

	if *nroMaxPeticiones == 0 {
		runForever = true
	}

	runtime.GOMAXPROCS(2)

	fTira, errorFTira := os.Open(*ficheroRq /*, os.O_RDONLY, 0*/)
	if errorFTira != nil {
		log.Fatalf("%s\n",errorFTira.Error())
	}
	defer fTira.Close()

	tiras := leeTiras(fTira)

	for c := 1; c <= *C; c = c + 1 {
		go generadorPeticiones(peticiones, resultados, fin)
	}
	go visualizador(resultados)
	go generadorTPS(*nroMaxPeticiones, *lambda, peticiones, &contadorPeticiones, runForever, tiras, *procesoLlegada)

	for c := 1; c <= *C; c = c + 1 {
		<-fin
	}
	// el programa termina con main
}

func leeTiras(fTira io.Reader) (tiras *ring.Ring) {
	tiras = nil
	var err error = nil
	lector := bufio.NewReader(fTira)

	var linea string 

	for linea, err = lector.ReadString('\n'); err == nil; linea, err = lector.ReadString('\n') {
        linea = strings.TrimRight(linea,"\r\n")

        if len(linea) > 0 {
			//fmt.Printf("<%s>\n",linea)

        	if (tiras.Len() == 0) {
            	tiras = new(ring.Ring)
        	} else {
	            tiras.Link(new(ring.Ring))
            	tiras = tiras.Next()    
        	}	        
    
        	arr := []byte(linea)

	        tiras.Value = arr[0:len(arr)]
    	}
	}

	if err == io.EOF {
		err = nil
	} else {
        log.Fatalf("%s\n",err.Error())
    }

    if tiras == nil {
        log.Fatalf("%s\n","No hay tiras")
    }
    
	return tiras
}
