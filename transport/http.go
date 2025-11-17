package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	orderapp "github.com/muhammadheryan/e-commerce/application/order"
	prodapp "github.com/muhammadheryan/e-commerce/application/product"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	warehouseapp "github.com/muhammadheryan/e-commerce/application/warehouse"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	utilsContext "github.com/muhammadheryan/e-commerce/utils/context"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	validatorx "github.com/muhammadheryan/e-commerce/utils/validator"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RestHandler struct {
	UserApp      userapp.UserApp
	ProductApp   prodapp.ProductApp
	OrderApp     orderapp.OrderApp
	WarehouseApp warehouseapp.WarehouseApp
}

func NewTransport(UserApp userapp.UserApp, ProductApp prodapp.ProductApp, OrderApp orderapp.OrderApp, WarehouseApp warehouseapp.WarehouseApp, internalAPIKey string) http.Handler {
	router := mux.NewRouter()

	rh := &RestHandler{
		UserApp:      UserApp,
		ProductApp:   ProductApp,
		OrderApp:     OrderApp,
		WarehouseApp: WarehouseApp,
	}

	// Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Public routes
	router.HandleFunc("/public/v1/register", rh.Register).Methods(http.MethodPost)
	router.HandleFunc("/public/v1/login", rh.Login).Methods(http.MethodPost)

	// Product routes
	router.HandleFunc("/public/v1/product", rh.GetProducts).Methods(http.MethodGet)
	router.HandleFunc("/public/v1//product/{id}", rh.GetProduct).Methods(http.MethodGet)

	// Order
	router.HandleFunc("/public/v1/order", rh.CreateOrder).Methods(http.MethodPost)
	router.HandleFunc("/public/v1/order/{id}/pay", rh.PayOrder).Methods(http.MethodPost)
	router.HandleFunc("/public/v1/order/{id}/cancel", rh.CancelOrder).Methods(http.MethodPost)

	// middleware
	router.Use(LoggingMiddleware())
	router.Use(AuthMiddleware(UserApp))

	// Internal route for MQ cancel (no auth, just API key)
	internal := mux.NewRouter()
	internal.HandleFunc("/internal/v1/order/{id}/cancel", rh.InternalCancelOrder).Methods(http.MethodPost)

	// Warehouse internal routes
	internal.HandleFunc("/internal/v1/warehouses/{id}/activate", rh.ActivateWarehouse).Methods(http.MethodPatch)
	internal.HandleFunc("/internal/v1/warehouses/{id}/deactivate", rh.DeactivateWarehouse).Methods(http.MethodPatch)
	internal.HandleFunc("/internal/v1/warehouses/transfer", rh.TransferStock).Methods(http.MethodPost)

	internal.Use(InternalMiddleware(internalAPIKey))
	router.PathPrefix("/internal/").Handler(internal)

	return router
}

// Register handler
// @Summary Register user
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Request"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} errors.CustomError
// @Router /public/v1/register [post]
func (s *RestHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if err := validatorx.ValidateStruct(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if s.UserApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}

	res, err := s.UserApp.Register(ctx, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, res)
}

// Login handler
// @Summary Login user
// @Description Login with email or phone and receive JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.LoginRequest true "Login Request"
// @Success 200 {object} model.LoginResponse
// @Failure 400 {object} errors.CustomError
// @Router /public/v1/login [post]
func (s *RestHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if err := validatorx.ValidateStruct(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if s.UserApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}

	res, err := s.UserApp.Login(ctx, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, res)
}

// @Summary List products
// @Description Get paginated list of products with shop and available stock
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(10)
// @Success 200 {object} model.ProductListResponse
// @Failure 400 {object} errors.CustomError
// @Security BearerAuth
// @Router /public/v1/product [get]
func (s *RestHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qs := r.URL.Query()
	page := 1
	perPage := 10
	if v := qs.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := qs.Get("per_page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			perPage = p
		}
	}

	res, err := s.ProductApp.ListProducts(ctx, page, perPage)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, res)
}

// @Summary Get product detail
// @Description Get product detail by id
// @Tags Product
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} model.ProductDetail
// @Failure 400 {object} errors.CustomError
// @Security BearerAuth
// @Router /public/v1/product/{id} [get]
func (s *RestHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	res, err := s.ProductApp.GetProduct(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, res)
}

// @Summary Create order
// @Description Create a new order and reserve stock
// @Tags Order
// @Accept json
// @Produce json
// @Param request body model.OrderRequest true "Order Request"
// @Success 200 {object} model.OrderResponse
// @Failure 400 {object} errors.CustomError
// @Security BearerAuth
// @Router /public/v1/order [post]
func (s *RestHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if err := validatorx.ValidateStruct(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	userID, ok := utilsContext.GetUserID(ctx)
	if !ok || userID == 0 {
		writeError(w, errors.SetCustomError(constant.ErrUnauthorize))
		return
	}

	res, err := s.OrderApp.CreateOrder(ctx, userID, &req)
	if err != nil {
		writeError(w, err)
		return
	}

	writeSuccess(w, res)
}

// @Summary Pay order
// @Description Mark order as paid and adjust stock
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.CustomError
// @Security BearerAuth
// @Router /public/v1/order/{id}/pay [post]
func (s *RestHandler) PayOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.OrderApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	userID, ok := utilsContext.GetUserID(ctx)
	if !ok || userID == 0 {
		writeError(w, errors.SetCustomError(constant.ErrUnauthorize))
		return
	}

	if err := s.OrderApp.PayOrder(ctx, id); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "paid"})
}

// @Summary Cancel order
// @Description Cancel order and release reservations
// @Tags Order
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.CustomError
// @Security BearerAuth
// @Router /public/v1/order/{id}/cancel [post]
func (s *RestHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.OrderApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	userID, ok := utilsContext.GetUserID(ctx)
	if !ok || userID == 0 {
		writeError(w, errors.SetCustomError(constant.ErrUnauthorize))
		return
	}

	if err := s.OrderApp.CancelOrder(ctx, id); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "cancelled"})
}

// InternalCancelOrder handles MQ-triggered cancel with API key only
func (s *RestHandler) InternalCancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	if err := s.OrderApp.CancelOrder(ctx, id); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "cancelled"})
}

// @Summary Activate warehouse
// @Description Activate a warehouse
// @Tags Warehouse
// @Accept json
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.CustomError
// @Security InternalAPIKey
// @Router /internal/v1/warehouses/{id}/activate [patch]
func (s *RestHandler) ActivateWarehouse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	if s.WarehouseApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}
	if err := s.WarehouseApp.ActivateWarehouse(ctx, id); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "activated"})
}

// @Summary Deactivate warehouse
// @Description Deactivate a warehouse. Cannot deactivate if there's reserved stock
// @Tags Warehouse
// @Accept json
// @Produce json
// @Param id path int true "Warehouse ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.CustomError
// @Security InternalAPIKey
// @Router /internal/v1/warehouses/{id}/deactivate [patch]
func (s *RestHandler) DeactivateWarehouse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	if s.WarehouseApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}
	if err := s.WarehouseApp.DeactivateWarehouse(ctx, id); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "deactivated"})
}

// @Summary Transfer stock between warehouses
// @Description Transfer stock from one warehouse to another. Only available stock (stock - reserved) can be transferred
// @Tags Warehouse
// @Accept json
// @Produce json
// @Param request body model.TransferStockHTTPRequest true "Transfer Stock Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.CustomError
// @Security InternalAPIKey
// @Router /internal/v1/warehouses/transfer [post]
func (s *RestHandler) TransferStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req model.TransferStockHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	if err := validatorx.ValidateStruct(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}
	if s.WarehouseApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}
	transferReq := &model.TransferStockRequest{
		ProductID:       req.ProductID,
		FromWarehouseID: req.FromWarehouseID,
		ToWarehouseID:   req.ToWarehouseID,
		Quantity:        req.Quantity,
	}
	if err := s.WarehouseApp.TransferStock(ctx, transferReq); err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, map[string]string{"status": "transferred"})
}
