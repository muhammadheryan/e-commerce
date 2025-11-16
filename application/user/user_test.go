package user_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	appuser "github.com/muhammadheryan/e-commerce/application/user"
	"github.com/muhammadheryan/e-commerce/cmd/config"
	"github.com/muhammadheryan/e-commerce/constant"
	redismocks "github.com/muhammadheryan/e-commerce/mocks/repository/redis"
	usermocks "github.com/muhammadheryan/e-commerce/mocks/repository/user"
	"github.com/muhammadheryan/e-commerce/model"
	cerr "github.com/muhammadheryan/e-commerce/utils/errors"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserApp_Register(t *testing.T) {
	type fields struct {
		config    *config.Config
		userRepo  *usermocks.UserRepository
		redisRepo *redismocks.RedisRepository
	}
	type args struct {
		ctx context.Context
		req *model.RegisterRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		want     *model.RegisterResponse
		wantErr  bool
		errCode  constant.ErrorType
	}{
		{
			name: "success: register new user",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.RegisterRequest{
					Name:     "Test User",
					Email:    "test@example.com",
					Phone:    "081234567890",
					Password: "password123",
				},
			},
			mockCall: func(f fields) {
				// Check email doesn't exist
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(nil, nil).
					Once()

				// Check phone doesn't exist
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Phone: "081234567890"}).
					Return(nil, nil).
					Once()

				// Create user
				f.userRepo.
					On("Create", mock.Anything, mock.MatchedBy(func(ent *model.UserEntity) bool {
						return ent.Name == "Test User" &&
							ent.Email == "test@example.com" &&
							ent.Phone == "081234567890" &&
							ent.PasswordHash != ""
					})).
					Return(&model.UserEntity{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						Phone:        "081234567890",
						PasswordHash: "hashed_password",
						CreatedAt:    time.Now(),
					}, nil).
					Once()
			},
			want: &model.RegisterResponse{
				Name:  "Test User",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "error: email already exists",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.RegisterRequest{
					Name:     "Test User",
					Email:    "existing@example.com",
					Phone:    "081234567890",
					Password: "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "existing@example.com"}).
					Return(&model.UserEntity{
						ID:    1,
						Email: "existing@example.com",
					}, nil).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrCredentialExists,
		},
		{
			name: "error: phone already exists",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.RegisterRequest{
					Name:     "Test User",
					Email:    "test@example.com",
					Phone:    "081111111111",
					Password: "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(nil, nil).
					Once()

				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Phone: "081111111111"}).
					Return(&model.UserEntity{
						ID:    1,
						Phone: "081111111111",
					}, nil).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrCredentialExists,
		},
		{
			name: "error: repository Get email returns error",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.RegisterRequest{
					Name:     "Test User",
					Email:    "test@example.com",
					Phone:    "081234567890",
					Password: "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(nil, errors.New("db error")).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: repository Create returns error",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.RegisterRequest{
					Name:     "Test User",
					Email:    "test@example.com",
					Phone:    "081234567890",
					Password: "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(nil, nil).
					Once()

				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Phone: "081234567890"}).
					Return(nil, nil).
					Once()

				f.userRepo.
					On("Create", mock.Anything, mock.AnythingOfType("*model.UserEntity")).
					Return(nil, errors.New("create failed")).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := appuser.NewUserApp(tt.fields.config, tt.fields.userRepo, tt.fields.redisRepo)

			got, err := app.Register(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[tt.errCode] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[tt.errCode])
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Register() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUserApp_Login(t *testing.T) {
	type fields struct {
		config    *config.Config
		userRepo  *usermocks.UserRepository
		redisRepo *redismocks.RedisRepository
	}
	type args struct {
		ctx context.Context
		req *model.LoginRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields)
		want     *model.LoginResponse
		wantErr  bool
		errCode  constant.ErrorType
	}{
		{
			name: "success: login with email",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "test@example.com",
					Password:   "password123",
				},
			},
			mockCall: func(f fields) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(&model.UserEntity{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						Phone:        "081234567890",
						PasswordHash: string(hashedPassword),
						CreatedAt:    time.Now(),
					}, nil).
					Once()

				f.redisRepo.
					On("SetSession", mock.Anything, mock.AnythingOfType("string"), uint64(1), time.Hour).
					Return(nil).
					Once()
			},
			want: &model.LoginResponse{
				Name:  "Test User",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "success: login with phone",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "081234567890",
					Password:   "password123",
				},
			},
			mockCall: func(f fields) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Phone: "081234567890"}).
					Return(&model.UserEntity{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						Phone:        "081234567890",
						PasswordHash: string(hashedPassword),
						CreatedAt:    time.Now(),
					}, nil).
					Once()

				f.redisRepo.
					On("SetSession", mock.Anything, mock.AnythingOfType("string"), uint64(1), time.Hour).
					Return(nil).
					Once()
			},
			want: &model.LoginResponse{
				Name:  "Test User",
				Email: "test@example.com",
			},
			wantErr: false,
		},
		{
			name: "error: user not found",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "notfound@example.com",
					Password:   "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "notfound@example.com"}).
					Return(nil, nil).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrNotFound,
		},
		{
			name: "error: invalid password",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "test@example.com",
					Password:   "wrongpassword",
				},
			},
			mockCall: func(f fields) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(&model.UserEntity{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInvalidPassword,
		},
		{
			name: "error: repository Get returns error",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "test@example.com",
					Password:   "password123",
				},
			},
			mockCall: func(f fields) {
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(nil, errors.New("db error")).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
		{
			name: "error: SetSession returns error",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
				req: &model.LoginRequest{
					Identifier: "test@example.com",
					Password:   "password123",
				},
			},
			mockCall: func(f fields) {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				f.userRepo.
					On("Get", mock.Anything, &model.UserFilter{Email: "test@example.com"}).
					Return(&model.UserEntity{
						ID:           1,
						Name:         "Test User",
						Email:        "test@example.com",
						PasswordHash: string(hashedPassword),
					}, nil).
					Once()

				f.redisRepo.
					On("SetSession", mock.Anything, mock.AnythingOfType("string"), uint64(1), time.Hour).
					Return(errors.New("redis error")).
					Once()
			},
			want:    nil,
			wantErr: true,
			errCode: constant.ErrInternal,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields)
			}
			app := appuser.NewUserApp(tt.fields.config, tt.fields.userRepo, tt.fields.redisRepo)

			got, err := app.Login(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				var ce cerr.CustomError
				if !errors.As(err, &ce) {
					t.Fatalf("error type = %T, want CustomError", err)
				}
				if ce.ErrorCode() != constant.ErrorTypeCode[tt.errCode] {
					t.Fatalf("error code = %s, want %s", ce.ErrorCode(), constant.ErrorTypeCode[tt.errCode])
				}
				return
			}

			if got.Name != tt.want.Name || got.Email != tt.want.Email {
				t.Fatalf("Login() = %+v, want %+v", got, tt.want)
			}
			if got.Token == "" {
				t.Fatal("Login() token should not be empty")
			}
		})
	}
}

func TestUserApp_ValidateToken(t *testing.T) {
	type fields struct {
		config    *config.Config
		userRepo  *usermocks.UserRepository
		redisRepo *redismocks.RedisRepository
	}
	type args struct {
		ctx         context.Context
		tokenString string
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockCall func(f fields, tokenString string)
		want     uint64
		wantErr  bool
	}{
		{
			name: "success: valid token",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
			},
			mockCall: func(f fields, tokenString string) {
				f.redisRepo.
					On("GetSession", mock.Anything, mock.AnythingOfType("string")).
					Return(uint64(1), nil).
					Once()
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "error: invalid token format",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx:         context.Background(),
				tokenString: "invalid.token.string",
			},
			mockCall: nil,
			want:     0,
			wantErr:  true,
		},
		{
			name: "error: session not found in redis",
			fields: fields{
				config: &config.Config{
					Auth: config.AuthConfig{
						JWTSecret:      "test-secret-key-for-jwt-signing",
						JWTExpiration:  time.Hour,
						SessionExpTime: time.Hour,
					},
				},
				userRepo:  usermocks.NewUserRepository(t),
				redisRepo: redismocks.NewRedisRepository(t),
			},
			args: args{
				ctx: context.Background(),
			},
			mockCall: func(f fields, tokenString string) {
				f.redisRepo.
					On("GetSession", mock.Anything, mock.AnythingOfType("string")).
					Return(uint64(0), errors.New("session not found")).
					Once()
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Generate a valid token for success case
			if tt.name == "success: valid token" || tt.name == "error: session not found in redis" {
				app := appuser.NewUserApp(tt.fields.config, tt.fields.userRepo, tt.fields.redisRepo)
				// Create a valid token by logging in first
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				tt.fields.userRepo.On("Get", mock.Anything, mock.Anything).Return(&model.UserEntity{
					ID:           1,
					PasswordHash: string(hashedPassword),
				}, nil).Once()
				tt.fields.redisRepo.On("SetSession", mock.Anything, mock.Anything, uint64(1), time.Hour).Return(nil).Once()

				loginResp, _ := app.Login(context.Background(), &model.LoginRequest{
					Identifier: "test@example.com",
					Password:   "password123",
				})
				if loginResp != nil {
					tt.args.tokenString = loginResp.Token
				}
			}

			if tt.mockCall != nil {
				ttFields := tt.fields
				tt.mockCall(ttFields, tt.args.tokenString)
			}

			app := appuser.NewUserApp(tt.fields.config, tt.fields.userRepo, tt.fields.redisRepo)

			got, err := app.ValidateToken(tt.args.ctx, tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got != tt.want {
				t.Fatalf("ValidateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
