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

	var mOfOpen sync.Mutex

	fmt.Println("inicio")

	now := time.Now()

	origen, err := ioutil.ReadDir("origen")

	if err != nil {
		log.Fatal(err)
	}

	canalArchivo := make(chan *os.File, len(origen))

	for _, infoArchivo := range origen {

		wOfOpen.Add(1)

		go abrir("origen", infoArchivo.Name(), os.O_RDONLY, canalArchivo, &wOfOpen)

	}

	wOfOpen.Wait()

	for _, infoArchivo := range origen {

		wOfMethod.Add(1)

		mOfOpen.Lock()

		archivoOrigen := <-canalArchivo

		mOfOpen.Unlock()

		direccion := filepath.Join("origen", infoArchivo.Name())

		go leer(archivoOrigen, direccion, &wOfMethod)

	}

	wOfMethod.Wait()

	fmt.Println(time.Since(now))
	fmt.Println("fin")
}

func abrir(carpeta string, nameFile string, flag int, canalArchivo chan<- *os.File, w *sync.WaitGroup) {

	defer w.Done()

	direccion := filepath.Join(carpeta, nameFile)

	fmt.Println("abro archivo: ", direccion)
	archivo, err := os.OpenFile(direccion, flag, 0)

	if err != nil {
		log.Println(err)
		return
	}
	canalArchivo <- archivo

}

func leer(archivo *os.File, direccion string, w *sync.WaitGroup) {

	defer w.Done()

	defer archivo.Close()

	defer fmt.Printf("Cierro archivo: %s\n", direccion)

	buf := make([]byte, 1)

	for {
		_, err := archivo.Read(buf)

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return
		}

	}

	return

}
