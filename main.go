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

	canalError := make(chan bool, len(origen))

	for i, infoArchivo := range origen {

		wOfOpen.Add(1)

		go abrir("origen", infoArchivo.Name(), os.O_RDONLY, canalArchivo, &wOfOpen, i, &mOfOpen)

		wOfMethod.Add(1)

		archivoOrigen := <-canalArchivo
		mOfOpen.Unlock()
		wOfOpen.Wait()

		direccion := filepath.Join("origen", infoArchivo.Name())

		go leer(archivoOrigen, direccion, &wOfMethod, &mOfOpen, canalError)

	}

	wOfMethod.Wait()

	fmt.Println(time.Since(now))
	fmt.Println("fin")
}

func abrir(carpeta string, nameFile string, flag int, canalArchivo chan<- *os.File, w *sync.WaitGroup, i int, m *sync.Mutex) {

	//En la función abir pongo la condicion para que halla un archivo el cual no tenga
	//El permiso de leer y me arroje un error.

	if i == 0 {
		defer w.Done()

		direccion := filepath.Join(carpeta, nameFile)

		fmt.Println("abro archivo: ", direccion)
		archivo, err := os.OpenFile(direccion, os.O_WRONLY, 0)

		if err != nil {
			log.Println(err)
			m.Lock()
			canalArchivo <- archivo
			return
		}
		m.Lock()
		canalArchivo <- archivo

	} else {
		defer w.Done()

		direccion := filepath.Join(carpeta, nameFile)

		fmt.Println("abro archivo: ", direccion)
		archivo, err := os.OpenFile(direccion, flag, 0)

		if err != nil {
			log.Println(err)
			m.Lock()
			canalArchivo <- archivo
			return
		}
		m.Lock()
		canalArchivo <- archivo
	}
}

func leer(archivo *os.File, direccion string, w *sync.WaitGroup, m *sync.Mutex, canalError chan bool) {

	defer w.Done()

	if archivo == nil {
		return
	}

	defer archivo.Close()

	defer fmt.Printf("Cierro archivo: %s\n", direccion)

	buf := make([]byte, 1)

	for {
		_, err := archivo.Read(buf)

		if err != nil {
			if err == io.EOF {
				break
			}
			m.Lock()
			canalError <- true

			//El fin de esta practica es hacer algo similar a lo que sucede cuando
			//En Windows ocurre un error al mover/copiar datos(se freezea los demas procesos)

			fmt.Printf("Ocurrio un error de acceso con el archivo %s , Introdusca cualquier valor para continuar\n", direccion)
			fmt.Scanln()
			m.Unlock()
			return
		}

		select {
		case <-canalError:
			m.Lock()

		default:
			break
		}

	}

	return

}
