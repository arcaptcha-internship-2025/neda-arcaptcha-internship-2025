# builder design pattern in go

## what is the builder pattern?

the builder pattern is a creational design pattern that lets you construct complex objects step by step. it allows you to produce different types and representations of an object using the same construction code.

## when to use builder pattern

- when creating complex objects with many optional parameters
- when object construction requires multiple steps
- when you want to create different representations of the same object
- when constructor would have too many parameters (telescoping constructor problem)

## key components

1. **builder interface** - defines steps to build the product
2. **concrete builder** - implements building steps and assembles the product
3. **product** - the complex object being built
4. **director** (optional) - defines order of construction steps

## problem it solves

consider creating a house object with many optional features:

```go
// bad approach - too many parameters
func NewHouse(walls int, doors int, windows int, hasGarage bool, hasPool bool, hasGarden bool) *House {
    // constructor nightmare
}
```

## basic implementation

### product
```go
type House struct {
    walls    int
    doors    int
    windows  int
    hasGarage bool
    hasPool   bool
    hasGarden bool
}

func (h *House) String() string {
    return fmt.Sprintf("house with %d walls, %d doors, %d windows, garage: %t, pool: %t, garden: %t",
        h.walls, h.doors, h.windows, h.hasGarage, h.hasPool, h.hasGarden)
}
```

### builder interface
```go
type HouseBuilder interface {
    SetWalls(walls int) HouseBuilder
    SetDoors(doors int) HouseBuilder  
    SetWindows(windows int) HouseBuilder
    SetGarage(hasGarage bool) HouseBuilder
    SetPool(hasPool bool) HouseBuilder
    SetGarden(hasGarden bool) HouseBuilder
    Build() *House
}
```

### concrete builder
```go
type ConcreteHouseBuilder struct {
    house *House
}

func NewHouseBuilder() *ConcreteHouseBuilder {
    return &ConcreteHouseBuilder{
        house: &House{},
    }
}

func (b *ConcreteHouseBuilder) SetWalls(walls int) HouseBuilder {
    b.house.walls = walls
    return b
}

func (b *ConcreteHouseBuilder) SetDoors(doors int) HouseBuilder {
    b.house.doors = doors
    return b
}

func (b *ConcreteHouseBuilder) SetWindows(windows int) HouseBuilder {
    b.house.windows = windows
    return b
}

func (b *ConcreteHouseBuilder) SetGarage(hasGarage bool) HouseBuilder {
    b.house.hasGarage = hasGarage
    return b
}

func (b *ConcreteHouseBuilder) SetPool(hasPool bool) HouseBuilder {
    b.house.hasPool = hasPool
    return b
}

func (b *ConcreteHouseBuilder) SetGarden(hasGarden bool) HouseBuilder {
    b.house.hasGarden = hasGarden
    return b
}

func (b *ConcreteHouseBuilder) Build() *House {
    return b.house
}
```

## usage example

```go
func main() {
    // building a luxury house
    luxuryHouse := NewHouseBuilder().
        SetWalls(4).
        SetDoors(3).
        SetWindows(12).
        SetGarage(true).
        SetPool(true).
        SetGarden(true).
        Build()
    
    fmt.Println(luxuryHouse)
    
    // building a simple house
    simpleHouse := NewHouseBuilder().
        SetWalls(4).
        SetDoors(2).
        SetWindows(6).
        Build()
    
    fmt.Println(simpleHouse)
}
```

## advanced example - car builder

### product
```go
type Car struct {
    engine      string
    transmission string
    seats       int
    color       string
    features    []string
}

func (c *Car) String() string {
    return fmt.Sprintf("car: %s engine, %s transmission, %d seats, %s color, features: %v",
        c.engine, c.transmission, c.seats, c.color, c.features)
}
```

### builder
```go
type CarBuilder struct {
    car *Car
}

func NewCarBuilder() *CarBuilder {
    return &CarBuilder{
        car: &Car{
            features: make([]string, 0),
        },
    }
}

func (b *CarBuilder) SetEngine(engine string) *CarBuilder {
    b.car.engine = engine
    return b
}

func (b *CarBuilder) SetTransmission(transmission string) *CarBuilder {
    b.car.transmission = transmission
    return b
}

func (b *CarBuilder) SetSeats(seats int) *CarBuilder {
    b.car.seats = seats
    return b
}

func (b *CarBuilder) SetColor(color string) *CarBuilder {
    b.car.color = color
    return b
}

func (b *CarBuilder) AddFeature(feature string) *CarBuilder {
    b.car.features = append(b.car.features, feature)
    return b
}

func (b *CarBuilder) Build() *Car {
    return b.car
}
```

### director (optional)
```go
type CarDirector struct {
    builder *CarBuilder
}

func NewCarDirector(builder *CarBuilder) *CarDirector {
    return &CarDirector{builder: builder}
}

func (d *CarDirector) BuildSportsCar() *Car {
    return d.builder.
        SetEngine("V8").
        SetTransmission("manual").
        SetSeats(2).
        SetColor("red").
        AddFeature("racing stripes").
        AddFeature("sport exhaust").
        Build()
}

func (d *CarDirector) BuildFamilyCar() *Car {
    return d.builder.
        SetEngine("V6").
        SetTransmission("automatic").
        SetSeats(7).
        SetColor("blue").
        AddFeature("safety package").
        AddFeature("entertainment system").
        Build()
}
```

## real-world example - http client builder

```go
type HTTPClient struct {
    baseURL string
    timeout time.Duration
    headers map[string]string
    client  *http.Client
}

type HTTPClientBuilder struct {
    client *HTTPClient
}

func NewHTTPClientBuilder() *HTTPClientBuilder {
    return &HTTPClientBuilder{
        client: &HTTPClient{
            headers: make(map[string]string),
            timeout: 30 * time.Second,
        },
    }
}

func (b *HTTPClientBuilder) SetBaseURL(url string) *HTTPClientBuilder {
    b.client.baseURL = url
    return b
}

func (b *HTTPClientBuilder) SetTimeout(timeout time.Duration) *HTTPClientBuilder {
    b.client.timeout = timeout
    return b
}

func (b *HTTPClientBuilder) AddHeader(key, value string) *HTTPClientBuilder {
    b.client.headers[key] = value
    return b
}

func (b *HTTPClientBuilder) Build() *HTTPClient {
    b.client.client = &http.Client{
        Timeout: b.client.timeout,
    }
    return b.client
}

// usage
func createAPIClient() *HTTPClient {
    return NewHTTPClientBuilder().
        SetBaseURL("https://api.example.com").
        SetTimeout(10 * time.Second).
        AddHeader("Content-Type", "application/json").
        AddHeader("Authorization", "Bearer token123").
        Build()
}
```

## validation in builder

```go
type DatabaseConfig struct {
    host     string
    port     int
    database string
    username string
    password string
}

type DatabaseConfigBuilder struct {
    config *DatabaseConfig
}

func NewDatabaseConfigBuilder() *DatabaseConfigBuilder {
    return &DatabaseConfigBuilder{
        config: &DatabaseConfig{},
    }
}

func (b *DatabaseConfigBuilder) SetHost(host string) *DatabaseConfigBuilder {
    b.config.host = host
    return b
}

func (b *DatabaseConfigBuilder) SetPort(port int) *DatabaseConfigBuilder {
    b.config.port = port
    return b
}

func (b *DatabaseConfigBuilder) SetDatabase(database string) *DatabaseConfigBuilder {
    b.config.database = database
    return b
}

func (b *DatabaseConfigBuilder) SetCredentials(username, password string) *DatabaseConfigBuilder {
    b.config.username = username
    b.config.password = password
    return b
}

func (b *DatabaseConfigBuilder) Build() (*DatabaseConfig, error) {
    if b.config.host == "" {
        return nil, errors.New("host is required")
    }
    if b.config.port <= 0 {
        return nil, errors.New("valid port is required")
    }
    if b.config.database == "" {
        return nil, errors.New("database name is required")
    }
    
    return b.config, nil
}
```

## functional options pattern (go idiom)

an alternative approach in go using functional options:

```go
type Server struct {
    host    string
    port    int
    timeout time.Duration
    debug   bool
}

type ServerOption func(*Server)

func WithHost(host string) ServerOption {
    return func(s *Server) {
        s.host = host
    }
}

func WithPort(port int) ServerOption {
    return func(s *Server) {
        s.port = port
    }
}

func WithTimeout(timeout time.Duration) ServerOption {
    return func(s *Server) {
        s.timeout = timeout
    }
}

func WithDebug(debug bool) ServerOption {
    return func(s *Server) {
        s.debug = debug
    }
}

func NewServer(opts ...ServerOption) *Server {
    server := &Server{
        host:    "localhost",
        port:    8080,
        timeout: 30 * time.Second,
        debug:   false,
    }
    
    for _, opt := range opts {
        opt(server)
    }
    
    return server
}

// usage
func main() {
    server := NewServer(
        WithHost("0.0.0.0"),
        WithPort(9090),
        WithTimeout(60 * time.Second),
        WithDebug(true),
    )
    
    fmt.Printf("%+v\n", server)
}
```

## benefits

- **flexibility** - create different representations of objects
- **readable code** - method chaining makes construction clear
- **immutability** - can create immutable objects
- **validation** - can validate during construction
- **step-by-step construction** - complex objects built incrementally

## drawbacks

- **code complexity** - requires more classes and interfaces
- **memory overhead** - additional objects created during construction
- **overkill** - unnecessary for simple objects

## when not to use

- simple objects with few parameters
- when object construction is straightforward
- when immutability is not required
- performance-critical code where object creation overhead matters

## best practices

1. return the builder from each method for method chaining
2. validate required fields in build() method
3. consider using functional options pattern for simple cases
4. make builders reusable by resetting state
5. use interfaces for builders when multiple implementations needed
6. consider immutability of the final product

## testing builders

```go
func TestCarBuilder(t *testing.T) {
    car := NewCarBuilder().
        SetEngine("V6").
        SetTransmission("automatic").
        SetSeats(4).
        SetColor("black").
        AddFeature("bluetooth").
        Build()
    
    assert.Equal(t, "V6", car.engine)
    assert.Equal(t, "automatic", car.transmission)
    assert.Equal(t, 4, car.seats)
    assert.Equal(t, "black", car.color)
    assert.Contains(t, car.features, "bluetooth")
}
```

the builder pattern is particularly useful in go for creating complex configuration objects, api clients, and domain objects with many optional parameters. it provides a clean, readable way to construct objects while maintaining flexibility and validation capabilities.

## references

- [Builder Pattern - Refactoring Guru](https://refactoring.guru/design-patterns/builder)
- [Builder Design Pattern - Dev.to](https://dev.to/srishtikprasad/builder-design-pattern-3a7j)
- [Mastering the Builder Design Pattern - Medium](https://medium.com/@kalanamalshan98/mastering-the-builder-design-pattern-build-objects-like-a-pro-0b8b36d10383)
- [Builder Pattern in Go - Refactoring Guru](https://refactoring.guru/design-patterns/builder/go/example)
- [Builder Pattern in Go - Medium](https://medium.com/@josueparra2892/builder-pattern-in-go-56605f9e7387)
