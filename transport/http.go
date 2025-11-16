package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	prodapp "github.com/muhammadheryan/e-commerce/application/product"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	validatorx "github.com/muhammadheryan/e-commerce/utils/validator"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RestHandler struct {
	UserApp    userapp.UserApp
	ProductApp prodapp.ProductApp
}

func NewTransport(UserApp userapp.UserApp, ProductApp prodapp.ProductApp) http.Handler {
	mux := mux.NewRouter()

	rh := &RestHandler{
		UserApp:    UserApp,
		ProductApp: ProductApp,
	}

	// Swagger UI
	mux.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Public routes
	mux.HandleFunc("/register", rh.Register).Methods(http.MethodPost)
	mux.HandleFunc("/login", rh.Login).Methods(http.MethodPost)

	// Product routes
	mux.HandleFunc("/products", rh.GetProducts).Methods(http.MethodGet)
	mux.HandleFunc("/products/{id}", rh.GetProduct).Methods(http.MethodGet)

	// middleware
	mux.Use(LoggingMiddleware())
	mux.Use(AuthMiddleware(UserApp))

	return mux
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
// @Router /products [get]
func (s *RestHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.ProductApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}

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
// @Router /products/{id} [get]
func (s *RestHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if s.ProductApp == nil {
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

	res, err := s.ProductApp.GetProduct(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeSuccess(w, res)
}
