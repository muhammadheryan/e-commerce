package transport

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	userapp "github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/constant"
	"github.com/muhammadheryan/e-commerce/model"
	utilsContext "github.com/muhammadheryan/e-commerce/utils/context"
	"github.com/muhammadheryan/e-commerce/utils/errors"
	validatorx "github.com/muhammadheryan/e-commerce/utils/validator"
	httpSwagger "github.com/swaggo/http-swagger"
)

type RestHandler struct {
	UserApp userapp.UserApp
}

func NewTransport(UserApp userapp.UserApp) http.Handler {
	mux := mux.NewRouter()

	rh := &RestHandler{
		UserApp: UserApp,
	}

	// Swagger UI
	mux.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Public routes
	mux.HandleFunc("/register", rh.Register).Methods(http.MethodPost)
	mux.HandleFunc("/login", rh.Login).Methods(http.MethodPost)

	// protected routes
	mux.HandleFunc("/test-doang", rh.TestDoang).Methods(http.MethodGet)

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

func (s *RestHandler) TestDoang(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, errors.SetCustomError(constant.ErrInvalidRequest))
		return
	}

	if s.UserApp == nil {
		writeError(w, errors.SetCustomError(constant.ErrInternal))
		return
	}

	id, _ := utilsContext.GetUserID(ctx)

	res := struct {
		ID uint64
	}{
		ID: id,
	}

	writeSuccess(w, res)
}
