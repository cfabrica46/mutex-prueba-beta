package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main() {

	log.SetFlags(log.Llongfile)

	var wOfMethod, wOfOpen sync.WaitGroup
	var m sync.Mutex

	fmt.Println("inicio")

	now := time.Now()

	origen, err := ioutil.ReadDir("origen")

	if err != nil {
		log.Fatal(err)
	}

	for i, infoArchivo := range origen {

		contenido := make(chan byte)

		fin := make(chan bool)

		direccion := filepath.Join("origen", infoArchivo.Name())

		fmt.Println("abro archivo: ", direccion)

		archivoOrigen := abrir("origen", infoArchivo.Name(), os.O_RDONLY)

		wOfMethod.Add(2)

		go leer(archivoOrigen, direccion, contenido, fin, &wOfMethod, &m)

		go escribir(contenido, fin, i, infoArchivo.Name(), &wOfMethod, &m)

	}

	wOfMethod.Wait()

	fmt.Println(time.Since(now))
	fmt.Println("fin")
}

func abrir(carpeta string, nameFile string, flag int) (archivo *os.File) {

	direccion := filepath.Join(carpeta, nameFile)

	archivo, err := os.OpenFile(direccion, flag, 0)

	if err != nil {
		log.Println(err)
		return
	}

	return
}

func leer(archivo *os.File, direccion string, contenido chan<- byte, fin chan<- bool, w *sync.WaitGroup, m *sync.Mutex) {

	defer w.Done()

	origen, err := os.Open(direccion)

	if err != nil {
		log.Println(err)
		return
	}

	defer origen.Close()

	buf := make([]byte, 1)

	for {

		_, err := origen.Read(buf)

		if err != nil {
			if err == io.EOF {

				fin <- true
				break
			}
			log.Println(err)
			contenido <- buf[0]
			return
		}

		contenido <- buf[0]
	}

}

func escribir(contenido <-chan byte, fin <-chan bool, i int, name string, w *sync.WaitGroup, m *sync.Mutex) {

	defer w.Done()

	var ok bool

	direccion := filepath.Join("destino", name)

	destino, err := os.OpenFile(direccion, os.O_RDWR|os.O_CREATE, 0)

	if err != nil {
		log.Println(err)
		return
	}

	defer destino.Close()

	for ok == false {
		select {
		case <-fin:

			ok = true
			break
		default:
			cont := <-contenido

			b := []byte{cont}
			_, err = destino.Write(b)

			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	return
}
