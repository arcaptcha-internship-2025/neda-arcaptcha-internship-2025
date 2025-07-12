# factory method design pattern in go - comprehensive guide

## what is it

the factory method pattern is a **creational design pattern** that defines an interface or method for creating objects, letting subclasses or specific implementations decide which concrete types to instantiate. the factory method pattern provides an interface for creating objects without specifying their exact class types and delegates the responsibility of object creation to subclasses.

in go, since there's no classic oop inheritance, we achieve this through interfaces and factory functions. it's impossible to implement the classic factory method pattern in go due to lack of oop features such as classes and inheritance, however, we can still implement the basic version of the pattern, the simple factory.

---

## why use it

- **abstracts creation**: keeps client code decoupled from concrete pizza types
- **supports open/closed principle**: supports the open/closed principle, allowing for the addition of new product types without modifying existing client code
- **simplifies unit testing**: with the use of interfaces and dependency injection, it becomes easier to write unit tests for the client code
- **centralizes creation logic**: useful for caching, pooling, or conditional instantiation
- **promotes loose coupling**: by decoupling object creation from the client code, it becomes easier to introduce new product variations or extend the existing ones

---

## core components

1. **product interface**  
   the abstract interface representing the objects the factory creates

2. **concrete product**  
   the specific implementations of the product interface

3. **creator (factory)**  
   the interface or class responsible for creating objects

4. **concrete creator**  
   the class that implements the factory interface and produces specific instances of the product

---

## basic go implementation

### simple factory approach

```go
package main

import (
    "fmt"
    "errors"
)

// product interface
type Gun interface {
    SetName(name string)
    SetPower(power int)
    GetName() string
    GetPower() int
}

// concrete product - base gun
type gun struct {
    name  string
    power int
}

func (g *gun) SetName(name string) { g.name = name }
func (g *gun) GetName() string     { return g.name }
func (g *gun) SetPower(power int)  { g.power = power }
func (g *gun) GetPower() int       { return g.power }

// concrete products
type AK47 struct {
    gun
}

func newAK47() Gun {
    return &AK47{
        gun: gun{
            name:  "AK47 gun",
            power: 4,
        },
    }
}

type Musket struct {
    gun
}

func newMusket() Gun {
    return &Musket{
        gun: gun{
            name:  "Musket gun", 
            power: 1,
        },
    }
}

// simple factory function
func GetGun(gunType string) (Gun, error) {
    switch gunType {
    case "ak47":
        return newAK47(), nil
    case "musket":
        return newMusket(), nil
    default:
        return nil, fmt.Errorf("wrong gun type passed: %s", gunType)
    }
}

// usage
func main() {
    ak47, _ := GetGun("ak47")
    musket, _ := GetGun("musket")
    
    fmt.Printf("gun: %s, power: %d\n", ak47.GetName(), ak47.GetPower())
    fmt.Printf("gun: %s, power: %d\n", musket.GetName(), musket.GetPower())
}
```

---

## advanced go variations

### 1. factory with configuration

in the real world, each payment gateway may require specific configuration parameters. here's how to handle different configurations:

```go
package main

import (
    "errors"
    "fmt"
)

// product interface
type PaymentGateway interface {
    ProcessPayment(amount float64) error
}

// concrete products
type PayPalGateway struct {
    ClientID     string
    ClientSecret string
}

func (pg *PayPalGateway) ProcessPayment(amount float64) error {
    fmt.Printf("processing paypal payment of $%.2f with clientid: %s\n", 
        amount, pg.ClientID)
    return nil
}

type StripeGateway struct {
    APIKey string
}

func (sg *StripeGateway) ProcessPayment(amount float64) error {
    fmt.Printf("processing stripe payment of $%.2f with apikey: %s\n", 
        amount, sg.APIKey)
    return nil
}

// configuration structs
type PayPalConfig struct {
    ClientID     string
    ClientSecret string
}

type StripeConfig struct {
    APIKey string
}

// gateway types
type PaymentGatewayType int

const (
    PayPalGateway PaymentGatewayType = iota
    StripeGateway
)

// factory with configuration
func NewPaymentGateway(gwType PaymentGatewayType, config interface{}) (PaymentGateway, error) {
    switch gwType {
    case PayPalGateway:
        paypalConfig, ok := config.(PayPalConfig)
        if !ok {
            return nil, errors.New("invalid config for paypal gateway")
        }
        return &PayPalGateway{
            ClientID:     paypalConfig.ClientID,
            ClientSecret: paypalConfig.ClientSecret,
        }, nil
        
    case StripeGateway:
        stripeConfig, ok := config.(StripeConfig)
        if !ok {
            return nil, errors.New("invalid config for stripe gateway")
        }
        return &StripeGateway{
            APIKey: stripeConfig.APIKey,
        }, nil
        
    default:
        return nil, errors.New("unsupported payment gateway type")
    }
}

func main() {
    // create paypal gateway
    paypal, err := NewPaymentGateway(PayPalGateway, PayPalConfig{
        ClientID:     "paypal-client-id",
        ClientSecret: "paypal-client-secret",
    })
    if err != nil {
        panic(err)
    }
    paypal.ProcessPayment(100.00)
    
    // create stripe gateway
    stripe, err := NewPaymentGateway(StripeGateway, StripeConfig{
        APIKey: "stripe-api-key",
    })
    if err != nil {
        panic(err)
    }
    stripe.ProcessPayment(150.50)
}
```

### 2. functional options pattern

functional options allow for a more flexible configuration by using variadic functions:

```go
package main

import (
    "errors"
    "fmt"
)

// payment gateway option function type
type PaymentGatewayOption func(PaymentGateway) error

// option functions
func WithClientID(clientID string) PaymentGatewayOption {
    return func(pg PaymentGateway) error {
        if paypal, ok := pg.(*PayPalGateway); ok {
            paypal.ClientID = clientID
            return nil
        }
        return errors.New("invalid option for this gateway")
    }
}

func WithClientSecret(clientSecret string) PaymentGatewayOption {
    return func(pg PaymentGateway) error {
        if paypal, ok := pg.(*PayPalGateway); ok {
            paypal.ClientSecret = clientSecret
            return nil
        }
        return errors.New("invalid option for this gateway")
    }
}

func WithAPIKey(apiKey string) PaymentGatewayOption {
    return func(pg PaymentGateway) error {
        if stripe, ok := pg.(*StripeGateway); ok {
            stripe.APIKey = apiKey
            return nil
        }
        return errors.New("invalid option for this gateway")
    }
}

// factory with functional options
func NewPaymentGateway(gwType PaymentGatewayType, opts ...PaymentGatewayOption) (PaymentGateway, error) {
    var pg PaymentGateway
    
    switch gwType {
    case PayPalGateway:
        pg = &PayPalGateway{}
    case StripeGateway:
        pg = &StripeGateway{}
    default:
        return nil, errors.New("unsupported payment gateway type")
    }
    
    for _, opt := range opts {
        if err := opt(pg); err != nil {
            return nil, err
        }
    }
    
    return pg, nil
}

func main() {
    // create paypal with options
    paypal, err := NewPaymentGateway(
        PayPalGateway,
        WithClientID("paypal-client-id"),
        WithClientSecret("paypal-client-secret"),
    )
    if err != nil {
        panic(err)
    }
    paypal.ProcessPayment(100.00)
    
    // create stripe with options
    stripe, err := NewPaymentGateway(
        StripeGateway,
        WithAPIKey("stripe-api-key"),
    )
    if err != nil {
        panic(err)
    }
    stripe.ProcessPayment(150.50)
}
```

### 3. separate factory interfaces

the factory design pattern offers several benefits, including increased flexibility:

```go
package main

import "fmt"

// product interface
type Car interface {
    Drive() string
}

// concrete products
type Sedan struct{}
func (s *Sedan) Drive() string { return "driving a sedan car" }

type SUV struct{}
func (s *SUV) Drive() string { return "driving an suv car" }

// factory interface
type CarFactory interface {
    CreateCar() Car
}

// concrete factories
type SedanFactory struct{}
func (sf *SedanFactory) CreateCar() Car { return &Sedan{} }

type SUVFactory struct{}
func (sf *SUVFactory) CreateCar() Car { return &SUV{} }

func main() {
    // create sedan using factory
    sedanFactory := &SedanFactory{}
    sedan := sedanFactory.CreateCar()
    fmt.Println(sedan.Drive())
    
    // create suv using factory
    suvFactory := &SUVFactory{}
    suv := suvFactory.CreateCar()
    fmt.Println(suv.Drive())
}
```

### 4. function-based factories

more idiomatic go approach using function types:

```go
package main

import "fmt"

// product interface
type Robot interface {
    GetType() string
    Work() string
}

// concrete products
type TeacherRobot struct{}
func (tr *TeacherRobot) GetType() string { return "teacher" }
func (tr *TeacherRobot) Work() string { return "teaching students" }

type FighterRobot struct{}
func (fr *FighterRobot) GetType() string { return "fighter" }
func (fr *FighterRobot) Work() string { return "fighting enemies" }

// factory function type
type RobotFactory func() Robot

// factory functions
func TeacherRobotFactory() Robot { return &TeacherRobot{} }
func FighterRobotFactory() Robot { return &FighterRobot{} }

// factory registry
var robotFactories = map[string]RobotFactory{
    "teacher": TeacherRobotFactory,
    "fighter": FighterRobotFactory,
}

func GetRobotFactory(robotType string) (RobotFactory, bool) {
    factory, exists := robotFactories[robotType]
    return factory, exists
}

func main() {
    // using factory functions
    teacherFactory, exists := GetRobotFactory("teacher")
    if exists {
        robot := teacherFactory()
        fmt.Printf("robot type: %s, work: %s\n", robot.GetType(), robot.Work())
    }
    
    fighterFactory, exists := GetRobotFactory("fighter")
    if exists {
        robot := fighterFactory()
        fmt.Printf("robot type: %s, work: %s\n", robot.GetType(), robot.Work())
    }
}
```

---

## practical use cases

### 1. database connection factory

```go
package main

import (
    "database/sql"
    "fmt"
)

// database interface
type Database interface {
    Connect() error
    Query(query string) (interface{}, error)
    Close() error
}

// concrete database implementations
type MySQL struct {
    connectionString string
}

func (m *MySQL) Connect() error {
    fmt.Println("connecting to mysql with:", m.connectionString)
    return nil
}

func (m *MySQL) Query(query string) (interface{}, error) {
    fmt.Println("executing mysql query:", query)
    return nil, nil
}

func (m *MySQL) Close() error {
    fmt.Println("closing mysql connection")
    return nil
}

type PostgreSQL struct {
    connectionString string
}

func (p *PostgreSQL) Connect() error {
    fmt.Println("connecting to postgresql with:", p.connectionString)
    return nil
}

func (p *PostgreSQL) Query(query string) (interface{}, error) {
    fmt.Println("executing postgresql query:", query)
    return nil, nil
}

func (p *PostgreSQL) Close() error {
    fmt.Println("closing postgresql connection")
    return nil
}

// database factory
type DatabaseType string

const (
    MySQLType      DatabaseType = "mysql"
    PostgreSQLType DatabaseType = "postgresql"
)

func NewDatabase(dbType DatabaseType, connectionString string) (Database, error) {
    switch dbType {
    case MySQLType:
        return &MySQL{connectionString: connectionString}, nil
    case PostgreSQLType:
        return &PostgreSQL{connectionString: connectionString}, nil
    default:
        return nil, fmt.Errorf("unsupported database type: %s", dbType)
    }
}

func main() {
    // create mysql database
    mysqlDB, err := NewDatabase(MySQLType, "user:password@tcp(localhost:3306)/dbname")
    if err != nil {
        panic(err)
    }
    mysqlDB.Connect()
    mysqlDB.Query("SELECT * FROM users")
    mysqlDB.Close()
    
    // create postgresql database
    postgresDB, err := NewDatabase(PostgreSQLType, "postgres://user:password@localhost/dbname")
    if err != nil {
        panic(err)
    }
    postgresDB.Connect()
    postgresDB.Query("SELECT * FROM users")
    postgresDB.Close()
}
```

### 2. logger factory

```go
package main

import (
    "fmt"
    "log"
    "os"
)

// logger interface
type Logger interface {
    Info(message string)
    Error(message string)
    Debug(message string)
}

// concrete loggers
type ConsoleLogger struct{}

func (cl *ConsoleLogger) Info(message string) {
    fmt.Printf("[INFO] %s\n", message)
}

func (cl *ConsoleLogger) Error(message string) {
    fmt.Printf("[ERROR] %s\n", message)
}

func (cl *ConsoleLogger) Debug(message string) {
    fmt.Printf("[DEBUG] %s\n", message)
}

type FileLogger struct {
    filename string
}

func (fl *FileLogger) Info(message string) {
    fl.writeToFile(fmt.Sprintf("[INFO] %s", message))
}

func (fl *FileLogger) Error(message string) {
    fl.writeToFile(fmt.Sprintf("[ERROR] %s", message))
}

func (fl *FileLogger) Debug(message string) {
    fl.writeToFile(fmt.Sprintf("[DEBUG] %s", message))
}

func (fl *FileLogger) writeToFile(message string) {
    file, err := os.OpenFile(fl.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Printf("error opening file: %v", err)
        return
    }
    defer file.Close()
    
    if _, err := file.WriteString(message + "\n"); err != nil {
        log.Printf("error writing to file: %v", err)
    }
}

// logger factory
type LoggerType string

const (
    ConsoleLoggerType LoggerType = "console"
    FileLoggerType    LoggerType = "file"
)

type LoggerConfig struct {
    Type     LoggerType
    Filename string // used for file logger
}

func NewLogger(config LoggerConfig) (Logger, error) {
    switch config.Type {
    case ConsoleLoggerType:
        return &ConsoleLogger{}, nil
    case FileLoggerType:
        if config.Filename == "" {
            return nil, fmt.Errorf("filename is required for file logger")
        }
        return &FileLogger{filename: config.Filename}, nil
    default:
        return nil, fmt.Errorf("unsupported logger type: %s", config.Type)
    }
}

func main() {
    // create console logger
    consoleLogger, err := NewLogger(LoggerConfig{Type: ConsoleLoggerType})
    if err != nil {
        panic(err)
    }
    consoleLogger.Info("this is console info")
    consoleLogger.Error("this is console error")
    
    // create file logger
    fileLogger, err := NewLogger(LoggerConfig{
        Type:     FileLoggerType,
        Filename: "app.log",
    })
    if err != nil {
        panic(err)
    }
    fileLogger.Info("this is file info")
    fileLogger.Error("this is file error")
}
```

---
#### DataStore Interface
```go
type DataStore interface {
  Name() string
  FindUserNameById(id int64) (string, error)
}
```

#### Implementations
- **PostgreSQLDataStore**: real SQL-backed `DataStore`
- **MemoryDataStore**: in‑memory, using `map[int64]string` + `sync.RWMutex`

#### Factory Type and Functions
```go
type DataStoreFactory func(conf map[string]string) (DataStore, error)

func NewPostgreSQLDataStore(conf map[string]string) (DataStore, error) { … }
func NewMemoryDataStore(conf map[string]string) (DataStore, error) { … }
```

#### Registry & Creator
```go
var datastoreFactories = make(map[string]DataStoreFactory)

func Register(name string, factory DataStoreFactory) { … }
func init() {
  Register("postgres", NewPostgreSQLDataStore)
  Register("memory", NewMemoryDataStore)
}

func CreateDatastore(conf map[string]string) (DataStore, error) {
  engine := conf["DATASTORE"] // default “memory”
  factory := datastoreFactories[engine] // error if not found
  return factory(conf)
}
```

> This pattern decouples client code, allows selecting implementations at runtime, and enables unit testing via mocks—all while encouraging runtime error checking for unregistered factories.

#### Testability & Integration Testing
- Client code works with interface rather than concretion.
- Option to inject `MockDataStore` during testing.
- Recommended: maintain real DB integration tests too.

---

## when to use

- **runtime object creation**: when you don't know at compile time what concrete type you'll need
- **decoupling**: when you want client code decoupled from concrete types
- **complex creation logic**: when creating objects involves conditional logic, pooling, or shared setup
- **configuration-based creation**: creating objects based on runtime conditions or configuration
- **testing**: when you need to easily swap implementations for testing
- **plugin architectures**: when supporting extensible plugin systems

---

## advantages & trade-offs

### advantages
- **loose coupling**: encourages loose coupling between client code and the created objects
- **extensibility**: increased flexibility by decoupling object creation from the client code
- **single responsibility**: provides a centralized point of control for object creation
- **open/closed principle**: supports the open/closed principle, allowing for the addition of new product types without modifying existing client code
- **testability**: improved testability with the use of interfaces and dependency injection

### trade-offs
- **complexity**: can lead to an explosion of subclasses if there are many variations of products
- **indirection**: increases complexity in the codebase, especially when dealing with multiple factories and product types
- **runtime errors**: may introduce runtime errors if the factory method is not implemented properly
- **go-specific**: in go, no inheritance means relying on interfaces and functions instead of traditional oop patterns

---

## best practices in go

1. **use interfaces wisely**: define minimal interfaces that focus on behavior
2. **prefer composition**: use struct embedding rather than trying to simulate inheritance
3. **handle errors properly**: always return errors from factory functions
4. **consider functional options**: for complex configurations, use functional options pattern
5. **keep it simple**: don't over-engineer - sometimes a simple factory function is enough
6. **use type safety**: leverage go's type system to catch errors at compile time
7. **document your factories**: clearly document what each factory creates and when to use it

---

## summary

the factory method pattern in go provides a powerful way to create objects while maintaining loose coupling and flexibility. while go doesn't support traditional oop inheritance, we can achieve similar benefits through interfaces, factory functions, and go's composition features.

key takeaways:
- **simple factory functions** work well for basic use cases
- **configuration-based factories** handle complex creation scenarios
- **functional options** provide flexible configuration
- **interface-based design** enables easy testing and extensibility
- **composition over inheritance** aligns with go's design philosophy

the pattern is particularly valuable when building scalable applications that need to support multiple implementations, configurations, or when you want to abstract creation logic from business logic.

---

## references

- medium.com (eshika shah): exploring the factory method design pattern
- dev.to (kittipat): understanding the factory method pattern in go  
- refactoring.guru: factory method in go example
- medium.com (swabhav techlabs): implementing the factory design pattern in golang
- factory vs simple factory vs abstract factory: https://refactoring.guru/design-patterns/factory-comparison  
- soham kamani's factory patterns guide: https://www.sohamkamani.com/golang/2018-06-20-golang-factory-patterns/
- matthew brown's factory method guide: https://matthewbrown.io/2016/01/23/factory-pattern-in-golang

---


