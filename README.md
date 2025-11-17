# E-Commerce Service

E-Commerce API built with **Golang**, **MySQL**, and **RabbitMQ**, featuring:
- Simple registration and auth
- Order creation with stock reservation
- Auto cancellation with delayed message RabbitMQ
- Transactional stock locking using `SELECT ... FOR UPDATE`
- Clean architecture (transport → application → repository → database)

---

## 🚀 Getting Started

This project run **full mode** via Docker:  
- Go service
- MySQL
- RabbitMQ
- Migration runner
- Swagger

---

## 🐳 Running with Docker

Requirement:
- Docker
- Docker Compose v2

Run:

```bash
docker-compose up -d --build
```

Open swagger via browser
```bash
http://localhost:8080/swagger/index.html
```
Dont forget to embed token in protected API via swagger athorize option 
```bash
/public
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWI....

/internal
Bearer <INTERNAL_API_KEY from .ENV>
```

---

## 📋 Current Features

- ✅ User Registration & Authentication (JWT)
- ✅ Product Listing & Detail
- ✅ Order Creation with Stock Reservation
- ✅ Order Payment
- ✅ Order Cancellation (Manual & Auto via RabbitMQ)
- ✅ Redis Session Management
- ✅ Swagger API Documentation

---

## 🚧 Next Steps / Future Enhancements
### 🏪 CRUD Management
- [ ] **Enhance CRUD Operations**
  - Complate CRUD feature

### 🔍 Search & Filtering
- [ ] **Advanced Search**
  - Category-based filtering
  - Full-text search for products
  - Search by product name, description
  - Price range filtering
  - Multi-criteria search

### 🛡️ Security & Performance
- [ ] **Security Enhancements**
  - Request validation improvements
  - CORS configuration
  - Password strength validation
  - Account lockout after failed login attempts

- [ ] **Performance Optimization**
  - Redis caching for frequently search
  - Database query optimization

### 📝 Testing & Quality
- [ ] **Testing**
  - Unit tests for all layers
  - Integration tests
  - Load testing
  - Test coverage reports

### 🚀 DevOps & Infrastructure
- [ ] **CI/CD Pipeline**
  - GitHub Actions / GitLab CI
  - Automated testing on PR
  - Automated deployment
  - Docker image optimization

- [ ] **Monitoring**
  - Structured logging improvements
  - Error tracking (Sentry, jaeger, etc.)
  - Application metrics (Prometheus)
  - Health check endpoints
  - Performance monitoring
---

