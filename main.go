package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Tarea struct {
	Nombre string `json:"nombre"`
	Estado bool   `json:"estado"`
	Fecha  string `json:"fecha"`
}

var listaTareas []Tarea

func main() {
	cargarTareas()

	http.HandleFunc("/crear", func(w http.ResponseWriter, r *http.Request) {
		nombreRecibido := r.URL.Query().Get("nombre")

		if nombreRecibido == "" {
			fmt.Fprint(w, "Error: Debes especificar un nombre", http.StatusBadRequest)
			return
		}

		ahora := time.Now().Format("02/01 15:04")
		nueva := Tarea{
			Nombre: nombreRecibido,
			Estado: false,
			Fecha:  ahora,
		}
		listaTareas = append(listaTareas, nueva)

		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/borrar", func(w http.ResponseWriter, r *http.Request) {
		idTexto := r.URL.Query().Get("id")

		id, err := strconv.Atoi(idTexto)
		if err != nil || id < 0 || id >= len(listaTareas) {
			fmt.Fprint(w, "Error: ID inválido", http.StatusBadRequest)
			return
		}

		listaTareas = append(listaTareas[:id], listaTareas[id+1:]...)
		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/completar", func(w http.ResponseWriter, r *http.Request) {
		idTexto := r.URL.Query().Get("id")

		id, err := strconv.Atoi(idTexto)

		if err != nil || id < 0 || id >= len(listaTareas) {
			http.Error(w, "Error a modificarlo", http.StatusInternalServerError)
			return
		}

		listaTareas[id].Estado = !listaTareas[id].Estado

		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/tareas", func(w http.ResponseWriter, r *http.Request) {
		plantilla, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, "Error al cargar la plantilla", 500)
			return
		}

		plantilla.Execute(w, listaTareas)
	})

	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}

	err := http.ListenAndServe(":"+puerto, nil)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
	}
}

func cargarTareas() {
	datos, err := os.ReadFile("tareas.json")
	if err != nil {
		return
	}

	json.Unmarshal(datos, &listaTareas)

}

func guardarenArchivos() {
	datosJson, _ := json.MarshalIndent(listaTareas, "", " ")

	os.WriteFile("tareas.json", datosJson, 0644)
}
