package main

import (
  "context"
  "encoding/json"
  "fmt"
  "net"
  "net/http"

  "github.com/bxcodec/faker/v3"
  "github.com/gorilla/mux"
  "github.com/rs/zerolog/log"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"

  "github.com/imrenagi/client-side-encryption/payment"
)

func main() {

  dsn := fmt.Sprintf("host=%s port=%s user=%s DB.name=%s password=%s sslmode=disable",
    "localhost",
    "5437",
    "payment",
    "payment",
    "payment")
  log.Debug().Msg(dsn)
  gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn}), &gorm.Config{})
  if err != nil {
    log.Fatal().Err(err).Msg("unable to create db connection")
  }

  err = gormDB.AutoMigrate(&payment.User{}, &payment.CreditCard{})

  router := mux.NewRouter()
  srv := &Server{
    Router: router,
    db:     gormDB,
  }

  srv.routesV1()

  srv.Run(context.Background(), 8090)
}

// Server ...
type Server struct {
  Router *mux.Router
  stopCh chan struct{}

  db *gorm.DB
}

// Run ...
func (g *Server) Run(ctx context.Context, port int) {

  httpS := http.Server{
    Addr:    fmt.Sprintf(":%d", port),
    Handler: g.Router,
  }

  // Start listener
  conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
  if err != nil {
    log.Fatal().Err(err).Msgf("failed listen")
  }

  log.Info().Msgf("payment service serving on port %d ", port)

  go func() { g.checkServeErr("httpS", httpS.Serve(conn)) }()

  g.stopCh = make(chan struct{})
  <-g.stopCh
  if err := conn.Close(); err != nil {
    panic(err)
  }
}

// checkServeErr checks the error from a .Serve() call to decide if it was a graceful shutdown
func (g *Server) checkServeErr(name string, err error) {
  if err != nil {
    if g.stopCh == nil {
      // a nil stopCh indicates a graceful shutdown
      log.Info().Msgf("graceful shutdown %s: %v", name, err)
    } else {
      log.Fatal().Msgf("%s: %v", name, err)
    }
  } else {
    log.Info().Msgf("graceful shutdown %s", name)
  }
}

func (g *Server) routesV1() {

  r := payment.UserRepository{DB: g.db}

  g.Router.HandleFunc("/", hcHandler())

  // serve api
  api := g.Router.PathPrefix("/users/").Subrouter()
  api.HandleFunc("/", listUsers(r)).Methods("GET")
  api.HandleFunc("/", createUser(r)).Methods("POST")
  api.HandleFunc("/rotate", rotate(r)).Methods("POST")
}

func hcHandler() http.HandlerFunc {
  return func(rw http.ResponseWriter, r *http.Request) {
    rw.Write([]byte("oke"))
  }
}


func rotate(repo payment.UserRepository) http.HandlerFunc {
  return func(rw http.ResponseWriter, r *http.Request) {

    err := repo.Rotate(r.Context())
    if err != nil {
      rw.WriteHeader(http.StatusInternalServerError)
      rw.Write([]byte(err.Error()))
    }

    rw.Header().Set("Content-Type", "application/json")
    rw.Write([]byte(`{}`))
  }
}


func listUsers(repo payment.UserRepository) http.HandlerFunc {
  return func(rw http.ResponseWriter, r *http.Request) {

    users, err := repo.FindAll(r.Context())
    if err != nil {
      rw.WriteHeader(http.StatusInternalServerError)
      rw.Write([]byte(err.Error()))
    }

    bytes, err := json.Marshal(users)
    if err != nil {
      rw.WriteHeader(http.StatusInternalServerError)
      rw.Write([]byte(err.Error()))
      return
    }

    rw.Header().Set("Content-Type", "application/json")
    rw.Write(bytes)
  }
}

func createUser(repo payment.UserRepository) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {

    u := &payment.User{}
    faker.FakeData(u)

    cc := &payment.CreditCard{}
    faker.FakeData(cc)

    user := payment.NewUserBuilder(u.Name).
      AddCard(cc.Number, cc.ExpireAt, cc.CVA).Build()

    err := repo.Save(r.Context(), user)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(err.Error()))
    }

    savedUser, err := repo.FindByID(r.Context(), user.ID.String())
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(err.Error()))
    }

    bytes, err := json.Marshal(savedUser)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(err.Error()))
      return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(bytes)
  }
}
