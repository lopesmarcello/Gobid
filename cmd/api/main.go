package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/lopesmarcello/gobid/internal/api"
	"github.com/lopesmarcello/gobid/internal/services"
)

func init() {
	gob.Register(uuid.UUID{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	fmt.Println("Setting env:")
	fmt.Println("User:", os.Getenv("GOBID_DATABASE_USER"))
	fmt.Println("Database:", os.Getenv("GOBID_DATABASE_NAME"))

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable search_path=public",
		os.Getenv("GOBID_DATABASE_USER"),
		os.Getenv("GOBID_DATABASE_PASSWORD"),
		os.Getenv("GOBID_DATABASE_HOST"),
		os.Getenv("GOBID_DATABASE_PORT"),
		os.Getenv("GOBID_DATABASE_NAME"),
	),
	)
	if err != nil {
		panic(err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	session := scs.New()
	session.Store = pgxstore.New(pool)
	session.Lifetime = 24 * time.Hour
	session.Cookie.HttpOnly = true
	session.Cookie.SameSite = http.SameSiteLaxMode

	api := api.API{
		Router:         chi.NewMux(),
		UserService:    services.NewUserService(pool),
		Session:        session,
		ProductService: services.NewProductServive(pool),
		BidsService:    services.NewBidsService(pool),
		WsUpgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // DEV only
			},
		},
		AuctionLobby: services.AuctionLobby{
			Rooms: make(map[uuid.UUID]*services.AuctionRoom),
		},
	}
	api.BindRoutes()

	port := os.Getenv("GOBID_APP_PORT")

	if port == "" {
		port = "3080"
	}

	fmt.Printf("Starting server on port :%s\n", port)
	if err := http.ListenAndServe(":"+port, api.Router); err != nil {
		panic(err)
	}
}
