package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"star-mining/internal/cache"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	router      *mux.Router
	handler     *APIHandler
	roomManager *RoomManager
	cache       *cache.RedisCache
	wsServer    *WebSocketServer
	addr        string
	staticDir   string
}

func NewHTTPServer(addr string, staticDir string, redisCache *cache.RedisCache) *HTTPServer {
	roomManager := NewRoomManager()
	wsServer := NewWebSocketServer(roomManager)
	handler := NewAPIHandler(roomManager, redisCache, wsServer)

	server := &HTTPServer{
		router:      mux.NewRouter(),
		handler:     handler,
		roomManager: roomManager,
		cache:       redisCache,
		wsServer:    wsServer,
		addr:        addr,
		staticDir:   staticDir,
	}

	server.setupRoutes()

	return server
}

func (s *HTTPServer) setupRoutes() {
	s.router.Use(s.corsMiddleware)

	apiRouter := s.router.PathPrefix("/api").Subrouter()

	apiRouter.HandleFunc("/login", s.handler.Login).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms", s.handler.CreateRoom).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms", s.handler.ListRooms).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}", s.handler.GetRoom).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/join", s.handler.JoinRoom).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/start", s.handler.StartGame).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/next-turn", s.handler.NextTurn).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/ready", s.handler.PlayerReady).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/unready", s.handler.PlayerUnready).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/action", s.handler.PlayerAction).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}", s.handler.GetPlayerState).Methods("GET", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/build-station", s.handler.BuildStation).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/build-refinery", s.handler.BuildRefinery).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/build-shipyard", s.handler.BuildShipyard).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/build-ship", s.handler.BuildShip).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/create-fleet", s.handler.CreateFleet).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/move-fleet", s.handler.MoveFleet).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/research-tech", s.handler.ResearchTech).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/place-bid", s.handler.PlaceBid).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/block-lane", s.handler.BlockLane).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/hire-pirates", s.handler.HirePirates).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/buy-stock", s.handler.BuyStock).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/sell-stock", s.handler.SellStock).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/propose-takeover", s.handler.ProposeTakeover).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/load-cargo", s.handler.LoadCargo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/unload-cargo", s.handler.UnloadCargo).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/upgrade-station", s.handler.UpgradeStation).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/upgrade-refinery", s.handler.UpgradeRefinery).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/buy-order", s.handler.PlaceBuyOrder).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/sell-order", s.handler.PlaceSellOrder).Methods("POST", "OPTIONS")
	apiRouter.HandleFunc("/rooms/{roomId}/players/{playerId}/cancel-order", s.handler.CancelOrder).Methods("POST", "OPTIONS")

	s.router.HandleFunc("/ws/{roomId}/{playerId}", s.handler.WebSocketHandler)

	if s.staticDir != "" {
		s.setupStaticFiles()
	}
}

func (s *HTTPServer) setupStaticFiles() {
	if _, err := os.Stat(s.staticDir); os.IsNotExist(err) {
		log.Printf("Static directory %s does not exist, skipping static file serving", s.staticDir)
		return
	}

	fs := http.FileServer(http.Dir(s.staticDir))

	s.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
			return
		}

		fullPath := filepath.Join(s.staticDir, r.URL.Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
			return
		}

		fs.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) Start() error {
	log.Printf("Server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

func (s *HTTPServer) GetRoomManager() *RoomManager {
	return s.roomManager
}

func (s *HTTPServer) GetHandler() *APIHandler {
	return s.handler
}
