package main

import (
	"dice/html"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	staticFiles := http.FS(html.StaticFiles)
	hub := newHub()
	go hub.run()

	r := chi.NewRouter()
	r.Handle("/public/*", http.FileServer(staticFiles))

	// main views
	r.Get("/", home)
	r.Post("/create", createHandler)
	r.Get("/game/{slug}", gameHandler)
	r.Post("/game/{slug}", joinHandler)

	// returns partials
	r.Get("/playerlobby/{slug}", getPlayerList)
	r.Get("/showtables", showActiveTables)
	r.Get("/join-table-form/{slug}", showJoinForm)

	// web socket
	r.Get("/ws/{slug}/{index}", func(w http.ResponseWriter, r *http.Request) {
		table := chi.URLParam(r, "slug")
		playerIndex := chi.URLParam(r, "index")
		if _, ok := activeTables[table]; !ok {
			return
		}

		if playerIndex == "" {
			return
		}
		index, err := strToInt(playerIndex)
		if err != nil {
			return
		}

		serveWs(activeTables[table], index, hub, w, r)
	})

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

func joinHandler(w http.ResponseWriter, r *http.Request) {
	tableSlug := chi.URLParam(r, "slug")
	playerName := r.PostFormValue("player-name")
	if playerName == "" {
		return
	}

	player := newPlayer(playerName)
	player.Index = len(activeTables[tableSlug].Players)

	activeTables[tableSlug].Players = append(activeTables[tableSlug].Players, player)

	html.Game(w, activeTables[tableSlug], player.Index)
}

func showActiveTables(w http.ResponseWriter, r *http.Request) {
	ts := make(map[string]string)
	for _, t := range activeTables {
		ts[t.InternalName] = t.Name
	}
	html.ShowActiveTables(w, ts)
}

func showJoinForm(w http.ResponseWriter, r *http.Request) {
	tableSlug := chi.URLParam(r, "slug")
	if _, ok := activeTables[tableSlug]; !ok {
		return
	}

	html.ShowJoinForm(w, activeTables[tableSlug])
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	tableSlug := chi.URLParam(r, "slug")
	table := activeTables[tableSlug]
	html.Game(w, table, 0)
}
