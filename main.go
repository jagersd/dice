package main

import (
	"dice/html"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	staticFiles := http.FS(html.StaticFiles)
	r := chi.NewRouter()
	r.Handle("/public/*", http.FileServer(staticFiles))

	// main views
	r.Get("/", home)
	r.Post("/create", createHandler)
	r.Get("/game/{slug}", gameHandler)

	// returns partials
	r.Get("/playerlobby/{slug}", getPlayerList)
	r.Get("/showtables", showActiveTables)

	http.ListenAndServe(":8080", r)
}

func home(w http.ResponseWriter, r *http.Request) {
	html.Home(w)
}

func getPlayerList(w http.ResponseWriter, r *http.Request) {
	tableSlug := chi.URLParam(r, "slug")
	players := activeTables[tableSlug].Players
	var playerNames []string
	for _, p := range players {
		playerNames = append(playerNames, p.Name)
	}
	html.LobbyPlayerList(w, playerNames)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	playerName := r.PostFormValue("player-name")
	tableName := r.PostFormValue("table-name")
	if playerName == "" || tableName == "" {
		return
	}
	slug, err := newTable(tableName, playerName)
	if err != nil {
		fmt.Fprint(w, err)
	}
	html.Lobby(w, activeTables[slug])
}

func showActiveTables(w http.ResponseWriter, r *http.Request) {
	ts := make(map[string]string)
	for _, t := range activeTables {
		ts[t.InternalName] = t.Name
	}
	html.ShowActiveTables(w, ts)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "todo")
}
