package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	muxR "github.com/gorilla/mux"
	orderapp "github.com/muhammadheryan/e-commerce/application/order"
	prodapp "github.com/muhammadheryan/e-commerce/application/product"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	utilsContext "github.com/muhammadheryan/e-commerce/utils/context"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	validatorx "github.com/muhammadheryan/e-commerce/utils/validator"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RestHandler struct {
	UserApp    userapp.UserApp
	ProductApp prodapp.ProductApp
	OrderApp   orderapp.OrderApp
}

func NewTransport(UserApp userapp.UserApp, ProductApp prodapp.ProductApp, OrderApp orderapp.OrderApp, internalAPIKey string) http.Handler {
	mux := muxR.NewRouter()

	rh := &RestHandler{
		UserApp:    UserApp,
		ProductApp: ProductApp,
		OrderApp:   OrderApp,
	}

	// Swagger UI
	mux.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Public routes
	mux.HandleFunc("/register", rh.Register).Methods(http.MethodPost)
	mux.HandleFunc("/login", rh.Login).Methods(http.MethodPost)

	// Product routes
	mux.HandleFunc("/product", rh.GetProducts).Methods(http.MethodGet)
	mux.HandleFunc("/product/{id}", rh.GetProduct).Methods(http.MethodGet)

	// Order
	mux.HandleFunc("/api/order", rh.CreateOrder).Methods(http.MethodPost)
	mux.HandleFunc("/api/order/{id}/pay", rh.PayOrder).Methods(http.MethodPost)
	mux.HandleFunc("/api/order/{id}/cancel", rh.CancelOrder).Methods(http.MethodPost)

	// middleware
	mux.Use(LoggingMiddleware())
	mux.Use(AuthMiddleware(UserApp))

	// Internal route for MQ cancel (no auth, just API key)
	internal := muxR.NewRouter()
	internal.HandleFunc("/internal/order/{id}/cancel", rh.InternalCancelOrder).Methods(http.MethodPost)
	internal.Use(InternalMiddleware(internalAPIKey))
	mux.PathPrefix("/internal/").Handler(internal)

	return mux
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

// Register handler
// @Summary Register user
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body model.RegisterRequest true "Register Request"
// @Success 200 {object} model.RegisterResponse
// @Failure 400 {object} errors.CustomError
// @Router /register [post]
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
// @Router /login [post]
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
// @Router /product [get]
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
// @Router /product/{id} [get]
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
// @Router /api/order [post]
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
// @Router /api/order/{id}/pay [post]
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
// @Router /api/order/{id}/cancel [post]
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
