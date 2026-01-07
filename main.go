package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Tarea struct {
	Nombre string `json:"nombre"`
	Estado bool   `json:"estado"`
	Fecha  string `json:"fecha"`
}

type Datos struct {
	Tareas      []Tarea
	Total       int
	Completadas int
	Usuario     string
}

type datosLogin struct {
	Error string
}

var usuarioTareas = make(map[string][]Tarea)
var tmpl = template.Must(template.ParseFiles("index.html"))

var mu sync.Mutex

func main() {
	cargarTareas()

	http.HandleFunc("/crear", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("usuario")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		usuarioActual := cookie.Value
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
		mu.Lock()
		usuarioTareas[usuarioActual] = append([]Tarea{nueva}, usuarioTareas[usuarioActual]...)
		mu.Unlock()

		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("usuario"); err == nil {
			http.Redirect(w, r, "/tareas", http.StatusSeeOther)
			return
		}

		tmplLogin := template.Must(template.ParseFiles("login.html"))
		tmplLogin.Execute(w, nil)
	})

	http.HandleFunc("/entrar", func(w http.ResponseWriter, r *http.Request) {
		nombre := r.FormValue("usuario")

		if nombre == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		cookie := http.Cookie{
			Name:    "usuario",
			Value:   nombre,
			Path:    "/",
			Expires: time.Now().Add(48 * time.Hour),
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/borrar", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("usuario")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		usuarioActual := cookie.Value
		idTexto := r.URL.Query().Get("id")

		id, err := strconv.Atoi(idTexto)
		if err != nil || id < 0 || id >= len(usuarioTareas[usuarioActual]) {
			http.Error(w, "Error: ID inválido", http.StatusBadRequest)
			return
		}

		mu.Lock()
		usuarioTareas[usuarioActual] = append(usuarioTareas[usuarioActual][:id], usuarioTareas[usuarioActual][id+1:]...)
		mu.Unlock()
		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/completar", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("usuario")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		usuarioActual := cookie.Value
		idTexto := r.URL.Query().Get("id")

		id, err := strconv.Atoi(idTexto)

		if err != nil || id < 0 || id >= len(usuarioTareas[usuarioActual]) {
			http.Error(w, "Error a modificarlo", http.StatusInternalServerError)
			return
		}

		mu.Lock()
		usuarioTareas[usuarioActual][id].Estado = !usuarioTareas[usuarioActual][id].Estado
		mu.Unlock()

		guardarenArchivos()

		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/tareas", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("usuario")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		usuarioActual := cookie.Value
		tareasUsuario := usuarioTareas[usuarioActual]
		completadas := 0
		for _, tarea := range tareasUsuario {
			if tarea.Estado {
				completadas++
			}
		}

		data := Datos{
			Tareas:      tareasUsuario,
			Total:       len(tareasUsuario),
			Completadas: completadas,
			Usuario:     usuarioActual,
		}

		tmpl.Execute(w, data)
	})

	http.HandleFunc("/limpiar", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("usuario")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		usuarioActual := cookie.Value
		var temp []Tarea

		for _, t := range usuarioTareas[usuarioActual] {
			if !t.Estado {
				temp = append(temp, t)
			}
		}

		usuarioTareas[usuarioActual] = temp
		guardarenArchivos()
		http.Redirect(w, r, "/tareas", http.StatusSeeOther)
	})

	http.HandleFunc("/salir", func(w http.ResponseWriter, r *http.Request) {
		cookie := http.Cookie{
			Name:   "usuario",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		}
		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusSeeOther)
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

	json.Unmarshal(datos, &usuarioTareas)

}

func guardarenArchivos() {
	dataJson, _ := json.MarshalIndent(usuarioTareas, "", " ")

	os.WriteFile("tareas.json", dataJson, 0644)
}
