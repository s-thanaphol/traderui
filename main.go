//	@title			Traderui API
//	@version		1.0
//	@termsOfService	http://somewhere.com/

//	@schemes	https http

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"

	docs "github.com/quickfixgo/traderui/docs"
    swaggerfiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gorilla/mux"
	"github.com/quickfixgo/traderui/basic"
	"github.com/quickfixgo/traderui/oms"
	"github.com/quickfixgo/traderui/secmaster"

	"github.com/quickfixgo/quickfix"
)

type fixFactory interface {
	NewOrderSingle(ord oms.Order) (msg quickfix.Messagable, err error)
	OrderCancelRequest(ord oms.Order, clOrdID string) (msg quickfix.Messagable, err error)
	SecurityDefinitionRequest(req secmaster.SecurityDefinitionRequest) (msg quickfix.Messagable, err error)
}

type tradeClient struct {
	SessionIDs map[string]quickfix.SessionID
	fixFactory
	*oms.OrderManager
}

func newTradeClient(factory fixFactory, idGen oms.ClOrdIDGenerator) *tradeClient {
	tc := &tradeClient{
		SessionIDs:   make(map[string]quickfix.SessionID),
		fixFactory:   factory,
		OrderManager: oms.NewOrderManager(idGen),
	}

	return tc
}

func (c tradeClient) SessionsAsJSON() (string, error) {
	sessionIDs := make([]string, 0, len(c.SessionIDs))

	for s := range c.SessionIDs {
		sessionIDs = append(sessionIDs, s)
	}

	b, err := json.Marshal(sessionIDs)
	return string(b), err
}

func (c tradeClient) OrdersAsJSON() (string, error) {
	c.RLock()
	defer c.RUnlock()

	b, err := json.Marshal(c.GetAll())
	return string(b), err
}

func (c tradeClient) ExecutionsAsJSON() (string, error) {
	c.RLock()
	defer c.RUnlock()

	b, err := json.Marshal(c.GetAllExecutions())
	return string(b), err
}

func (c tradeClient) traderView(w http.ResponseWriter, r *http.Request) {
	var templates = template.Must(template.New("traderui").ParseFiles("tmpl/index.html"))
	if err := templates.ExecuteTemplate(w, "index.html", c); err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c tradeClient) fetchRequestedOrder(r *gin.Context) (*oms.Order, error) {
	tmp := r.Param("id")
	id, err := strconv.Atoi(tmp)
	if err != nil {
		panic(err)
	}

	return c.Get(id)
}

func (c tradeClient) fetchRequestedExecution(r *http.Request) (*oms.Execution, error) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(err)
	}

	return c.GetExecution(id)
}

//	@Summary		getOrder
//	@Description	get 1 order
//	@Produce		json
//  @param id path int true "id of order"
//	@Success		200	{object} oms.OrderForSwag	
//	@Router			/orders/:id [get]
func (c tradeClient) getOrder(r *gin.Context) {
	c.RLock()
	defer c.RUnlock()

	order, err := c.fetchRequestedOrder(r)
	if err != nil {
		r.JSON(http.StatusBadRequest, http.StatusNotFound)
		return
	}

	c.writeOrderJSON(r, order)
}

func (c tradeClient) writeOrderJSON(r *gin.Context, order *oms.Order) {
	fmt.Println(order)
	outgoingJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		r.JSON(http.StatusBadRequest, http.StatusInternalServerError)
		return
	}

	var m oms.Order
	json.Unmarshal([]byte(outgoingJSON), &m)
	r.JSON(http.StatusOK, m)
}

func (c tradeClient) getExecution(w http.ResponseWriter, r *http.Request) {
	c.RLock()
	defer c.RUnlock()

	exec, err := c.fetchRequestedExecution(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	outgoingJSON, err := json.Marshal(exec)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(outgoingJSON))
}

func (c tradeClient) deleteOrder(r *gin.Context) {
	c.Lock()
	defer c.Unlock()

	order, err := c.fetchRequestedOrder(r)
	if err != nil {
		r.JSON(http.StatusBadRequest, http.StatusNotFound)
		return
	}

	clOrdID := c.AssignNextClOrdID(order)
	msg, err := c.OrderCancelRequest(*order, clOrdID)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		r.JSON(http.StatusBadRequest, http.StatusInternalServerError)
		return
	}

	err = quickfix.SendToTarget(msg, order.SessionID)
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		r.JSON(http.StatusBadRequest, http.StatusInternalServerError)
		return
	}

	c.writeOrderJSON(r, order)
}

//	@Summary		getOrders
//	@Description	get all order
//	@Produce		json
//	@Success		200	{array} oms.OrderForSwag	
//	@Router			/orders [get]
func (c tradeClient) getOrders(r *gin.Context) {
	outgoingJSON, err := c.OrdersAsJSON()
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		r.JSON(http.StatusBadRequest, gin.H{"error": "err getOrders funtion"})
		return
	}

	var m []oms.Order
	json.Unmarshal([]byte(outgoingJSON), &m)
	r.JSON(http.StatusOK, m)
}

func (c tradeClient) getExecutions(w http.ResponseWriter, r *http.Request) {
	outgoingJSON, err := c.ExecutionsAsJSON()
	if err != nil {
		log.Printf("[ERROR] err = %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, outgoingJSON)
}

func (c tradeClient) newSecurityDefintionRequest(w http.ResponseWriter, r *http.Request) {
	var secDefRequest secmaster.SecurityDefinitionRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&secDefRequest)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("secDefRequest = %+v\n", secDefRequest)

	if sessionID, ok := c.SessionIDs[secDefRequest.Session]; ok {
		secDefRequest.SessionID = sessionID
	} else {
		log.Println("[ERROR] Invalid SessionID")
		http.Error(w, "Invalid SessionID", http.StatusBadRequest)
		return
	}

	msg, err := c.fixFactory.SecurityDefinitionRequest(secDefRequest)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = quickfix.SendToTarget(msg, secDefRequest.SessionID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// @Summary newOrder
// @Description new 1 Order api
// @Accept json
// @Produce json
// @Param Order body oms.OrderForSwag true "Order data for sending to executor "
// @Success 200 {string} sting "OK"
// @Router /orders [post]
func (c tradeClient) newOrder(r *gin.Context) {
	fmt.Println("start new order")
	var order oms.Order
	decoder := json.NewDecoder(r.Request.Body)
	err := decoder.Decode(&order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		r.JSON(http.StatusBadRequest, gin.H{"error": "decode order err"})
		return
	}

	fmt.Println(2)
	fmt.Printf("%+v\n", order)
	if sessionID, ok := c.SessionIDs[order.Session]; ok {
		order.SessionID = sessionID
	} else {
		log.Println("[ERROR] Invalid SessionID")
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SessionID"})
		return
	}

	if err = order.Init(); err != nil {
		log.Printf("[ERROR] %v\n", err)
		r.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	c.Lock()
	_ = c.OrderManager.Save(&order)
	c.Unlock()

	msg, err := c.NewOrderSingle(order)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		r.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}

	err = quickfix.SendToTarget(msg, order.SessionID)

	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": ""})
		return
	}
	fmt.Println(1)
	r.JSON(http.StatusOK, "send 1 order successful")
	fmt.Println(2)
}

func main() {
	flag.Parse()

	cfgFileName := path.Join("config", "tradeclient.cfg")
	if flag.NArg() > 0 {
		cfgFileName = flag.Arg(0)
	}

	cfg, err := os.Open(cfgFileName)
	if err != nil {
		fmt.Printf("Error opening %v, %v\n", cfgFileName, err)
		return
	}

	appSettings, err := quickfix.ParseSettings(cfg)
	if err != nil {
		fmt.Println("Error reading cfg,", err)
		return
	}

	logFactory := NewFancyLog()

	var fixApp quickfix.Application
	app := newTradeClient(basic.FIXFactory{}, new(basic.ClOrdIDGenerator))
	fixApp = &basic.FIXApplication{
		SessionIDs:   app.SessionIDs,
		OrderManager: app.OrderManager,
	}

	initiator, err := quickfix.NewInitiator(fixApp, quickfix.NewMemoryStoreFactory(), appSettings, logFactory)
	if err != nil {
		log.Fatalf("Unable to create Initiator: %s\n", err)
	}

	if err = initiator.Start(); err != nil {
		log.Fatal(err)
	}
	defer initiator.Stop()

	router := gin.Default() //mux.NewRouter().StrictSlash(true)
	docs.SwaggerInfo.BasePath = ""
	router.GET("orders", app.getOrders)
	router.POST("/orders", app.newOrder)
	router.GET("orders/:id", app.getOrder)
	router.DELETE("/orders/:id", app.deleteOrder)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	//router.HandleFunc("/orders", app.newOrder).Methods("POST")
	//router.HandleFunc("/orders", app.getOrders).Methods("GET")

	//router.HandleFunc("/orders/{id:[0-9]+}", app.getOrder).Methods("GET")
	//router.HandleFunc("/orders/{id:[0-9]+}", app.deleteOrder).Methods("DELETE")

	// router.HandleFunc("/executions", app.getExecutions).Methods("GET")
	// router.HandleFunc("/executions/{id:[0-9]+}", app.getExecution).Methods("GET")

	// router.HandleFunc("/securitydefinitionrequest", app.newSecurityDefintionRequest).Methods("POST")

	// router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	// router.HandleFunc("/", app.traderView)

	//log.Fatal(http.ListenAndServe(":8080", router))
	router.Run(":8080")
}
