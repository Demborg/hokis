package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Turn struct {
	Image       string
	Description string
}

var turns []Turn

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload-turn", uploadTurnHandler)
	http.HandleFunc("/recent-turns", recentTurnsHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, nil); err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadTurnHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB max file size
	if err != nil {
		log.Println("Error parsing multipart form:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("turnImage")
	if err != nil {
		log.Println("Error retrieving form file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	filePath := filepath.Join("static", "uploads", handler.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		log.Println("Error creating file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = out.ReadFrom(file)
	if err != nil {
		log.Println("Error reading from file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	description := r.FormValue("description")
	newTurn := Turn{Image: "/" + filePath, Description: description}
	turns = append(turns, newTurn)

	tmpl, err := template.New("turn").Parse(`
        <div class="turn">
            <img src="{{.Image}}" alt="Turn">
            <p>{{.Description}}</p>
            <button>Vote</button>
        </div>
    `)
	if err != nil {
		log.Println("Error parsing turn template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, newTurn); err != nil {
		log.Println("Error executing turn template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func recentTurnsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("turn").Parse(`
        <div class="turn">
            <img src="{{.Image}}" alt="Turn">
            <p>{{.Description}}</p>
            <button>Vote</button>
        </div>
    `)
	if err != nil {
		log.Println("Error parsing turn template:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	for _, turn := range turns {
		if err := tmpl.Execute(w, turn); err != nil {
			log.Println("Error executing turn template:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
