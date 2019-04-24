package restclientsupport

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	msg_pb "github.com/harmony-one/harmony/api/proto/message"
	"github.com/harmony-one/harmony/internal/utils"
)

/*
 * This service is specific to support the lottery app.
 */

// Constants for rest client support service.
const (
	Port = "30000"
)

// Service is the struct for rest client support service.
type Service struct {
	router                          *mux.Router
	server                          *http.Server
	CreateTransactionForEnterMethod func(int64, string) error
	GetResult                       func(string) ([]string, []*big.Int)
	CreateTransactionForPickWinner  func() error
	messageChan                     chan *msg_pb.Message
	CallFaucetContract              func(common.Address) common.Hash
}

// New returns new client support service.
func New(
	CreateTransactionForEnterMethod func(int64, string) error,
	GetResult func(string) ([]string, []*big.Int),
	CreateTransactionForPickWinner func() error,
	CallFaucetContract func(common.Address) common.Hash) *Service {
	return &Service{
		CreateTransactionForEnterMethod: CreateTransactionForEnterMethod,
		GetResult:                       GetResult,
		CreateTransactionForPickWinner:  CreateTransactionForPickWinner,
		CallFaucetContract:              CallFaucetContract,
	}
}

// StartService starts rest client support service.
func (s *Service) StartService() {
	utils.GetLogInstance().Info("Starting rest client support.")
	s.server = s.Run()
}

// StopService shutdowns rest client support service.
func (s *Service) StopService() {
	utils.GetLogInstance().Info("Shutting down rest client support service.")
	if err := s.server.Shutdown(context.Background()); err != nil {
		utils.GetLogInstance().Error("Error when shutting down rest client support server", "error", err)
	} else {
		utils.GetLogInstance().Error("Shutting down rest client support server successufully")
	}
}

// Run is to run serving rest client support.
func (s *Service) Run() *http.Server {
	// Init address.
	addr := net.JoinHostPort("", Port)

	s.router = mux.NewRouter()
	// Set up router for blocks.
	s.router.Path("/enter").Queries("key", "{[0-9A-Fa-fx]*?}", "amount", "{[0-9]*?}").HandlerFunc(s.Enter).Methods("GET")
	s.router.Path("/enter").HandlerFunc(s.Enter)

	// Set up router for result.
	s.router.Path("/result").Queries("key", "{[0-9A-Fa-fx]*?}").HandlerFunc(s.Result).Methods("GET")
	s.router.Path("/result").HandlerFunc(s.Result)

	// Set up router for fundme.
	s.router.Path("/fundme").Queries("key", "{[0-9A-Fa-fx]*?}").HandlerFunc(s.FundMe).Methods("GET")
	s.router.Path("/fundme").HandlerFunc(s.FundMe)

	// Set up router for fundme.
	s.router.Path("/balance").Queries("key", "{[0-9A-Fa-fx]*?}").HandlerFunc(s.GetBalance).Methods("GET")
	s.router.Path("/balance").HandlerFunc(s.GetBalance)

	// Set up router for winner.
	s.router.Path("/winner").HandlerFunc(s.Winner)

	// Do serving now.
	utils.GetLogInstance().Info("Listening on ", "port: ", Port)
	server := &http.Server{Addr: addr, Handler: s.router}
	go server.ListenAndServe()
	return server
}

// Response is the data struct used the respoind to lottery app.
type Response struct {
	Players  []string `json:"players"`
	Balances []string `json:"balances"`
	Success  bool     `json:"success"`
}

// GetBalance implements the GetFreeToken interface to request free token.
func (s *Service) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := r.FormValue("key")
	res := &Response{Success: false}
	res.Players = append(res.Players, key)
	// res.Balances = append(res.Players, GETMYBALANCE) // implements this
	json.NewEncoder(w).Encode(res)

}

// FundMe implements the GetFreeToken interface to request free token.
func (s *Service) FundMe(w http.ResponseWriter, r *http.Request) {
	// (ctx context.Context, request *proto.GetFreeTokenRequest) (*proto.GetFreeTokenResponse, error) {
	// var address common.Address
	// address.SetBytes(request.Address)
	// //	log.Println("Returning GetFreeTokenResponse for address: ", address.Hex())
	// return &proto.GetFreeTokenResponse{TxId: s.callFaucetContract(address).Bytes()}, nil
	w.Header().Set("Content-Type", "application/json")
	// key := r.FormValue("key")

	// TODO(RJ): implement the logic similar to get FreeToken with key.

	// GetFreeToken implements the GetFreeToken interface to request free token.
	// func (s *Server) GetFreeToken(ctx context.Context, request *proto.GetFreeTokenRequest) (*proto.GetFreeTokenResponse, error) {
	// 	var address common.Address
	// 	address.SetBytes(request.Address)
	// 	//	log.Println("Returning GetFreeTokenResponse for address: ", address.Hex())
	// 	return &proto.GetFreeTokenResponse{TxId: s.callFaucetContract(address).Bytes()}, nil
	// }

	res := &Response{Success: false}
	json.NewEncoder(w).Encode(res)
}

// Enter triggers enter method of smart contract.
func (s *Service) Enter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := r.FormValue("key")
	amount := r.FormValue("amount")
	fmt.Println("enter: key", key, "amount", amount)
	amountInt, err := strconv.ParseInt(amount, 10, 64)

	res := &Response{Success: false}
	if err != nil {
		json.NewEncoder(w).Encode(res)
		return
	}
	if s.CreateTransactionForEnterMethod == nil {
		json.NewEncoder(w).Encode(res)
		return
	}
	if err := s.CreateTransactionForEnterMethod(amountInt, key); err != nil {
		json.NewEncoder(w).Encode(res)
		return
	}

	res.Success = true
	json.NewEncoder(w).Encode(res)
}

// Result generates result of result end point.
func (s *Service) Result(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	key := r.FormValue("key")
	res := &Response{
		Success: false,
	}

	fmt.Println("result: key", key)
	if s.GetResult == nil {
		json.NewEncoder(w).Encode(res)
		return
	}
	players, balances := s.GetResult(key)
	balancesString := []string{}
	for _, balance := range balances {
		balancesString = append(balancesString, balance.String())
	}
	res.Players = players
	res.Balances = balancesString
	res.Success = true
	json.NewEncoder(w).Encode(res)
}

// Winner triggers winner method of lottery smart contract.
func (s *Service) Winner(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Println("winner")
	res := &Response{
		Success: false,
	}

	if s.CreateTransactionForPickWinner == nil {
		json.NewEncoder(w).Encode(res)
		return
	}

	if err := s.CreateTransactionForPickWinner(); err != nil {
		utils.GetLogInstance().Error("error", err)
		json.NewEncoder(w).Encode(res)
		return
	}
	res.Success = true
	json.NewEncoder(w).Encode(res)
}

// NotifyService notify service
func (s *Service) NotifyService(params map[string]interface{}) {
	return
}

// SetMessageChan sets up message channel to service.
func (s *Service) SetMessageChan(messageChan chan *msg_pb.Message) {
	s.messageChan = messageChan
}