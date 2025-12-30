# 🚀 Hướng Dẫn Nhanh - Go Fiber Backend

## 📖 Tổng quan dự án

Dự án này sử dụng **Go Fiber** framework với mô hình **Clean Architecture** và **Dependency Injection**, bao gồm các thành phần chính:

- **Controllers**: Xử lý HTTP requests
- **Services**: Business logic 
- **Repositories**: Data access layer
- **Entities**: Data models
- **DTOs**: Data Transfer Objects
- **Cache**: Redis caching
- **Database**: GORM với PostgreSQL/MySQL

## 🏗️ Kiến trúc tổng quan

```
├── backend/
│   ├── main.go                    # Entry point
│   ├── app/                       # Application layer
│   │   ├── controllers/           # HTTP controllers
│   │   ├── services/              # Business logic implementations
│   │   ├── repository/            # Repository implementations
│   │   ├── interfaces/            # Interface definitions
│   │   │   ├── services/          # Service interfaces
│   │   │   └── repositories/      # Repository interfaces
│   │   ├── entities/              # Data models
│   │   └── dto/                   # Data transfer objects
│   ├── platform/                  # Infrastructure layer
│   │   ├── database/              # Database connections
│   │   └── cache/                 # Cache layer
│   └── pkg/                       # Shared packages
│       ├── di/                    # Dependency injection
│       ├── routes/                # Route definitions
│       ├── middleware/            # HTTP middleware
│       ├── mappers/               # Object mappers
│       └── utils/                 # Utilities
```

## 🔄 Luồng hoạt động của Request

```
HTTP Request → Router → Middleware → Controller → Service → Repository → Database
                                       ↓
Response ← JSON ← Api Response ← DTO ← Mapper ← Entity ← Query Result
```

## 🚀 Khởi tạo ứng dụng

### 1. **Entry Point** - [`main.go`](backend/main.go)

```go
func main() {
    // 1. Tạo Fiber app config
    config := configs.FiberConfig()
    app := fiber.New(config)
    
    // 2. Khởi tạo Dependency Injection Container
    container, err := di.NewContainer()
    
    // 3. Setup middleware
    middleware.FiberMiddleware(app)
    
    // 4. Setup routes với controllers từ container
    routes.PublicRoutes(app, container.AuthController)
    routes.PrivateRoutes(app, container.AuthController, 
                         container.TokenController, container.TaskController)
    
    // 5. Start server
    utils.StartServer(app)
}
```

### 2. **Dependency Injection** - [`pkg/di/container.go`](backend/pkg/di/container.go)

Container khởi tạo tất cả dependencies theo thứ tự:

```go
func NewContainer() (*Container, error) {
    // 1. Database connection
    gormDB, err := database.OpenGORMDBConnection()
    
    // 2. Cache service
    cacheService, err := cache.NewCacheService()
    
    // 3. Repositories (implement interfaces)
    var userRepo apprepos.UserRepository = repository.NewUserRepository(gormDB)
    var taskRepo apprepos.TaskRepository = repository.NewTaskRepository(gormDB)
    var taskHistoryRepo apprepos.TaskHistoryRepository = repository.NewTaskHistoryRepository(gormDB)
    
    // 4. Services (inject repositories)
    authService := serviceimpl.NewAuthService(userRepo, cacheService)
    tokenService := serviceimpl.NewTokenService(userRepo, cacheService)
    taskService := serviceimpl.NewTaskService(taskRepo, taskHistoryRepo)
    
    // 5. Controllers (inject services)
    authCtrl := controllers.NewAuthController(authService)
    tokenCtrl := controllers.NewTokenController(tokenService)
    taskCtrl := controllers.NewTaskController(taskService)
    
    // 6. Middleware
    jwtConfig := middleware.JWTConfig{
        SecretKey: os.Getenv("JWT_SECRET_KEY"),
    }
    jwtMiddleware := middleware.NewJWTProtected(jwtConfig)
    
    return &Container{...}, nil
}
```

## 🔗 Tổ chức Interfaces

Trong cấu trúc mới, tất cả interfaces được tổ chức riêng trong thư mục [`app/interfaces/`](backend/app/interfaces/):

### **Service Interfaces** - [`app/interfaces/services/`](backend/app/interfaces/services/)
- [`AuthService`](backend/app/interfaces/services/auth_service.go) - Authentication business logic
- [`TaskService`](backend/app/interfaces/services/task_service.go) - Task management business logic
- [`TokenService`](backend/app/interfaces/services/token_service.go) - Token management

### **Repository Interfaces** - [`app/interfaces/repositories/`](backend/app/interfaces/repositories/)
- [`UserRepository`](backend/app/interfaces/repositories/user_repository.go) - User data access
- [`TaskRepository`](backend/app/interfaces/repositories/task_repository.go) - Task data access
- [`TaskHistoryRepository`](backend/app/interfaces/repositories/task_history_repository.go) - Task history tracking

**Lợi ích của cách tổ chức này:**
- **Clean Separation**: Interfaces được tách biệt hoàn toàn với implementations
- **Easy Testing**: Dễ dàng mock interfaces cho unit testing
- **Dependency Inversion**: Tuân thủ nguyên tắc SOLID
- **Better Organization**: Code được tổ chức rõ ràng và dễ bảo trì

## 🎯 Chi tiết từng Layer

### **1. Controllers** - HTTP Request Handlers

**Ví dụ:** [`app/controllers/task_controller.go`](backend/app/controllers/task_controller.go)

```go
type TaskController struct {
    taskService services.TaskService  // Inject service interface
}

func NewTaskController(taskService services.TaskService) *TaskController {
    return &TaskController{taskService: taskService}
}

// HTTP handler
func (ctl *TaskController) GetTasks(c *fiber.Ctx) error {
    // 1. Gọi service
    resp, err := ctl.taskService.GetTasks(c.Context())
    
    // 2. Handle error
    if err != nil {
        return c.Status(500).JSON(core.Error(500, "internal error", err.Error(), nil))
    }
    
    // 3. Return response
    return c.Status(resp.Code).JSON(resp)
}
```

**Trách nhiệm:**
- Parse HTTP request (query params, body, headers)
- Validate input cơ bản
- Gọi service methods
- Format và return JSON response

### **2. Services** - Business Logic Layer

**Interface:** [`app/interfaces/services/task_service.go`](backend/app/interfaces/services/task_service.go)

```go
type TaskService interface {
    GetTasks(ctx context.Context) (*core.ApiResponse, error)
    GetTask(ctx context.Context, id string) (*core.ApiResponse, error)
    Create(ctx context.Context, c any, req *dto.CreateTaskReq) (*core.ApiResponse, error)
    Update(ctx context.Context, c any, task *models.Task) (*core.ApiResponse, error)
    Delete(ctx context.Context, c any, id string) (*core.ApiResponse, error)
}
```

**Implementation:** [`app/services/task_service.go`](backend/app/services/task_service.go)

```go
type TaskServiceImpl struct {
    taskRepo        repositories.TaskRepository
    taskHistoryRepo repositories.TaskHistoryRepository
}

func NewTaskService(taskRepo repositories.TaskRepository, taskHistoryRepo repositories.TaskHistoryRepository) services.TaskService {
    return &TaskServiceImpl{
        taskRepo:        taskRepo,
        taskHistoryRepo: taskHistoryRepo,
    }
}

func (s *TaskServiceImpl) GetTasks(ctx context.Context) (*core.ApiResponse, error) {
    // 1. Gọi repository để lấy data (với context)
    tasks, err := s.taskRepo.GetTasks(ctx)
    
    // 2. Handle error
    if err != nil {
        return core.Error(fiber.StatusNotFound, "tasks not found", err.Error(), fiber.Map{
            "count": 0,
        }), nil
    }
    
    // 3. Map entity to response DTO
    mapper := &genmapper.TaskMapperImpl{}
    res := mapper.EntitiesToResList(tasks)
    
    // 4. Return success response
    return core.Success(fiber.StatusOK, "ok", res, nil), nil
}
```

**Trách nhiệm:**
- Business logic và validation
- Authentication và authorization
- Gọi multiple repositories nếu cần
- Transform data (mapping entities ↔ DTOs)
- Return standardized API responses

### **3. Repositories** - Data Access Layer

**Interface:** [`app/interfaces/repositories/task_repository.go`](backend/app/interfaces/repositories/task_repository.go)

```go
type TaskRepository interface {
    GetTasks(ctx context.Context) ([]models.Task, error)
    GetTask(ctx context.Context, id uuid.UUID) (models.Task, error)
    GetTasksByStatus(ctx context.Context, status string) ([]models.Task, error)
    CreateTask(ctx context.Context, t *models.Task) error
    UpdateTask(ctx context.Context, id uuid.UUID, t *models.Task) error
    DeleteTask(ctx context.Context, id uuid.UUID) error
}
```

**Implementation:** [`app/repository/task_repository.go`](backend/app/repository/task_repository.go)

```go
type TaskRepositoryImpl struct {
    db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) repositories.TaskRepository {
    return &TaskRepositoryImpl{db: db}
}

func (r *TaskRepositoryImpl) GetTasks(ctx context.Context) ([]models.Task, error) {
    var tasks []models.Task
    err := r.db.WithContext(ctx).  // Sử dụng context
        Preload("Creator").          // Eager loading relationships
        Preload("Assignee").
        Order("created_at DESC").
        Find(&tasks).Error
    return tasks, err
}
```

**Trách nhiệm:**
- Database queries và operations với context support
- Object-Relational Mapping (ORM)
- Handle database errors và timeouts
- Optimize queries (preload, indexes, etc.)

**Context Usage:**
```go
// Tất cả repository methods sử dụng context để:
func (r *TaskRepositoryImpl) CreateTask(ctx context.Context, t *models.Task) error {
    now := time.Now()
    if t.ID == uuid.Nil {
        t.ID = uuid.New()
    }
    if t.CreatedAt.IsZero() {
        t.CreatedAt = now
    }
    t.UpdatedAt = now
    return r.db.WithContext(ctx).Create(t).Error
}

// - Handle request timeouts
// - Cancellation propagation
// - Tracing và logging
// - Database connection management
```

### **4. Entities** - Data Models

**Ví dụ:** [`app/entities/task_model.go`](backend/app/entities/task_model.go)

```go
type Task struct {
    ID          uuid.UUID `json:"id" gorm:"type:char(36);primarykey"`
    Title       string    `json:"title" gorm:"size:255;not null"`
    Description string    `json:"description" gorm:"type:text"`
    Status      string    `json:"status" gorm:"size:50;default:'NEW'"`
    CreatedBy   uuid.UUID `json:"created_by" gorm:"type:char(36);not null"`
    AssignedTo  *uuid.UUID `json:"assigned_to,omitempty" gorm:"type:char(36)"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // Relationships
    Creator  User  `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
    Assignee *User `json:"assignee,omitempty" gorm:"foreignKey:AssignedTo;references:ID"`
}
```

### **5. Mappers** - Object Transformation

**Generated Mappers:** [`pkg/mappers/generated/generated.go`](backend/pkg/mappers/generated/generated.go)

```go
type TaskMapperImpl struct {}

// Convert request DTO to entity
func (m *TaskMapperImpl) CreateReqToEntity(req dto.CreateTaskReq) models.Task {
    return models.Task{
        Title:       req.Title,
        Description: req.Description,
        AssignedTo:  req.AssignedTo,
    }
}

// Convert entity to response DTO
func (m *TaskMapperImpl) EntityToRes(entity models.Task) dto.TaskRes {
    return dto.TaskRes{
        ID:          entity.ID,
        Title:       entity.Title,
        Description: entity.Description,
        Status:      entity.Status,
        CreatedAt:   entity.CreatedAt,
        UpdatedAt:   entity.UpdatedAt,
        // Map relationships
        Creator:     UserEntityToRes(entity.Creator),
        Assignee:    entity.Assignee != nil ? &UserEntityToRes(*entity.Assignee) : nil,
    }
}

// Convert slice of entities to slice of response DTOs
func (m *TaskMapperImpl) EntitiesToResList(entities []models.Task) []dto.TaskRes {
    result := make([]dto.TaskRes, len(entities))
    for i, entity := range entities {
        result[i] = m.EntityToRes(entity)
    }
    return result
}
```

**Trách nhiệm:**
- Transform data giữa các layer khác nhau
- Decouple internal models từ external interfaces
- Handle complex object mapping với relationships
- Ensure data consistency và validation

### **6. DTOs** - Data Transfer Objects

**Request DTO:** [`app/dto/task_req.go`](backend/app/dto/task_req.go)
**Response DTO:** [`app/dto/task_res.go`](backend/app/dto/task_res.go)

```go
type CreateTaskReq struct {
    Title       string     `json:"title" validate:"required,min=3,max=255"`
    Description string     `json:"description" validate:"max=1000"`
    AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
}

type TaskRes struct {
    ID          uuid.UUID `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Creator     UserRes   `json:"creator"`
    Assignee    *UserRes  `json:"assignee,omitempty"`
}
```

## 🔗 Routes & Endpoints

### **Public Routes** - [`pkg/routes/public_routes.go`](backend/pkg/routes/public_routes.go)

```go
func PublicRoutes(a *fiber.App, authController *controllers.AuthController) {
    route := a.Group("/api/v1")
    
    route.Post("/user/sign/up", authController.UserSignUp)
    route.Post("/user/sign/in", authController.UserSignIn)
}
```

### **Private Routes** - [`pkg/routes/private_routes.go`](backend/pkg/routes/private_routes.go)

```go
func PrivateRoutes(a *fiber.App, auth *controllers.AuthController, 
                   token *controllers.TokenController, task *controllers.TaskController) {
    route := a.Group("/api/v1")
    
    // Auth routes
    route.Post("/user/sign/out", middleware.JWTProtected(), auth.UserSignOut)
    route.Post("/token/renew", middleware.JWTProtected(), token.RenewTokens)
    
    // Task management
    route.Post("/task", middleware.JWTProtected(), task.CreateTask)
    route.Put("/task/:id", middleware.JWTProtected(), task.UpdateTask)
    route.Delete("/task/:id", middleware.JWTProtected(), task.DeleteTask)
    route.Get("/tasks", middleware.JWTProtected(), task.GetTasks)
    route.Get("/task/:id", middleware.JWTProtected(), task.GetTask)
}
```

## 🔒 Authentication & Security

### **JWT Middleware** - [`pkg/middleware/jwt_middleware.go`](backend/pkg/middleware/jwt_middleware.go)

```go
func JWTProtected() func(*fiber.Ctx) error {
    return jwtware.New(jwtware.Config{
        SigningKey:   jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET_KEY"))},
        ContextKey:   "jwt",
        ErrorHandler: jwtError,
    })
}
```

## 💾 Database Layer

### **Connection Setup** - [`platform/database/open_db_connection.go`](backend/platform/database/open_db_connection.go)

```go
func OpenGORMDBConnection() (*gorm.DB, error) {
    dbType := os.Getenv("DB_TYPE")
    
    switch dbType {
    case "pgx":
        return GORMPostgreSQLConnection()
    case "mysql":
        return GORMMysqlConnection()
    default:
        return GORMPostgreSQLConnection() // default to PostgreSQL
    }
}
```

### **PostgreSQL Connection** - [`platform/database/gorm_postgres.go`](backend/platform/database/gorm_postgres.go)

```go
func GORMPostgreSQLConnection() (*gorm.DB, error) {
    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
        os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSL_MODE"), os.Getenv("DB_TIMEZONE"),
    )
    
    return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
```

## 🚀 Cache & Redis

### **Cache Service** - [`platform/cache/cache_service.go`](backend/platform/cache/cache_service.go)

```go
type CacheService struct {
    client *redis.Client
}

func NewCacheService() (*CacheService, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })
    
    return &CacheService{client: client}, nil
}
```

## 📝 API Response Format

### **Success Response**

```json
{
    "code": 200,
    "message": "ok",
    "data": {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "title": "Sample Task",
        "status": "NEW"
    },
    "meta": null
}
```

### **Error Response**

```json
{
    "code": 400,
    "message": "validation error",
    "data": null,
    "error": {
        "title": "Title is required",
        "description": "Description too long"
    }
}
```

## 🛠️ Development Workflow

### **1. Thêm tính năng mới:**

1. **Tạo Entity** → `app/entities/`
2. **Tạo DTO** → `app/dto/`
3. **Tạo Repository Interface** → `app/interfaces/repositories/`
4. **Implement Repository** → `app/repository/`
5. **Tạo Service Interface** → `app/interfaces/services/`
6. **Implement Service** → `app/services/`
7. **Tạo Controller** → `app/controllers/`
8. **Update DI Container** → `pkg/di/container.go`
9. **Add Routes** → `pkg/routes/`

### **2. Testing:**

```bash
# Run tests
go test ./...

# Run specific test
go test ./app/services -v

# Run with coverage
go test -cover ./...
```

### **3. Database Migration:**

```bash
# Create migration
migrate create -ext sql -dir platform/migrations -seq create_new_table

# Run migrations
migrate -path platform/migrations -database "postgres://..." up

# Rollback
migrate -path platform/migrations -database "postgres://..." down 1
```

## 🌍 Environment Variables

Tạo file `.env`:

```bash
# Database
DB_TYPE=pgx
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=fiber_db
DB_SSL_MODE=disable

# JWT
JWT_SECRET_KEY=your-secret-key
JWT_SECRET_KEY_EXPIRE_MINUTES_COUNT=15
JWT_REFRESH_KEY=your-refresh-key
JWT_REFRESH_KEY_EXPIRE_HOURS_COUNT=720

# Redis
REDIS_HOST=localhost  
REDIS_PORT=6379
REDIS_PASSWORD=

# Server
STAGE_STATUS=dev
```

## 🚀 Chạy ứng dụng

```bash
# Development
go run main.go

# Build và run
go build -o app
./app

# Với Docker
docker-compose up -d
```

## 📚 Tài liệu bổ sung

- **API Documentation:** [`docs/API_DOCS.md`](backend/docs/API_DOCS.md)
- **Redis Integration:** [`docs/REDIS_INTEGRATION.md`](backend/docs/REDIS_INTEGRATION.md)  
- **Business Logic:** [`app/BUSINESS_LOGIC.md`](backend/app/BUSINESS_LOGIC.md)
- **Platform Level:** [`platform/PLATFORM_LEVEL.md`](backend/platform/PLATFORM_LEVEL.md)

## 🎯 Các nguyên tắc quan trọng

1. **Separation of Concerns**: Mỗi layer có trách nhiệm riêng biệt
2. **Dependency Injection**: Inject dependencies thay vì hard-code
3. **Interface Driven**: Sử dụng interfaces để decouple code
4. **Error Handling**: Handle errors ở mỗi layer phù hợp
5. **Validation**: Validate input ở controller và service layer
6. **Security**: Implement authentication, authorization đúng cách
7. **Testing**: Write tests cho từng layer
8. **Documentation**: Document APIs với Swagger

---

🎉 **Happy Coding!** Hy vọng hướng dẫn này giúp bạn hiểu rõ cấu trúc và cách thức hoạt động của dự án Go Fiber này!
