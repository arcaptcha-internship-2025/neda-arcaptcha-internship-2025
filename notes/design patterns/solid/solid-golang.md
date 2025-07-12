# SOLID principles in Go with Examples

Go is not a classic object-oriented language (no inheritance, no classes), but you can still apply SOLID principles using interfaces, composition, and structs. Here's how each principle works in Go with code examples.

## Single Responsibility Principle (SRP)

A package or struct should have one and only one reason to change

### Example:

```go
package invoice

type Invoice struct {
    ID    int
    Total float64
}

// responsible only for calculations
func (i *Invoice) CalculateTotal() float64 {
    return i.Total
}

// separate module for printing
package printer

import "fmt"

type Printer struct{}

func (p *Printer) PrintInvoice(invoiceID int) {
    fmt.Printf("Printing Invoice #%d
", invoiceID)
}
```

each package/struct has one responsibility: business logic vs. output logic

## Open/Closed Principle (OCP)

code should be open for extension but closed for modification

use interfaces to allow new behavior without changing existing code

### Example:

```go
package payment

type PaymentMethod interface {
    Pay(amount float64)
}

type CreditCard struct{}

func (c CreditCard) Pay(amount float64) {
    fmt.Printf("Paid %.2f with Credit Card
", amount)
}

type PayPal struct{}

func (p PayPal) Pay(amount float64) {
    fmt.Printf("Paid %.2f with PayPal
", amount)
}

type Checkout struct {
    method PaymentMethod
}

func (c *Checkout) SetMethod(m PaymentMethod) {
    c.method = m
}

func (c *Checkout) Process(amount float64) {
    c.method.Pay(amount)
}
```

You can add ApplePay, Crypto, etc., without modifying Checkout

## Liskov Substitution Principle (LSP)

subtypes must be substitutable for their base types

go uses interfaces naturally, so this fits well

### Example:

```go
package bird

type Flyer interface {
    Fly()
}

type Sparrow struct{}

func (s Sparrow) Fly() {
    fmt.Println("Sparrow flying")
}

type Penguin struct{}

// penguin can't fly so don't make it implement Flyer
```

don’t force Penguin to implement Flyer — instead, group birds by behavior

## I — Interface Segregation Principle (ISP)

clients shouldn't depend on methods they don't use

split large interfaces into smaller, specific ones

### Example:

```go
package machine

type Printer interface {
    Print()
}

type Scanner interface {
    Scan()
}

type MultiFunctionPrinter struct{}

func (m MultiFunctionPrinter) Print() {
    fmt.Println("Printing...")
}

func (m MultiFunctionPrinter) Scan() {
    fmt.Println("Scanning...")
}

type SimplePrinter struct{}

func (s SimplePrinter) Print() {
    fmt.Println("Just printing.")
}
```

aach device only implements what it needs.

## Dependency Inversion Principle (DIP)

high-level modules should not depend on low-level modules. both should depend on abstractions

use interfaces to decouple code

### Example:

```go
package logger

type Logger interface {
    Log(message string)
}

type ConsoleLogger struct{}

func (c ConsoleLogger) Log(message string) {
    fmt.Println("LOG:", message)
}

package service

import "yourapp/logger"

type UserService struct {
    logger logger.Logger
}

func NewUserService(l logger.Logger) *UserService {
    return &UserService{logger: l}
}

func (s *UserService) CreateUser(name string) {
    s.logger.Log("User created: " + name)
}
```

now you can inject FileLogger, ConsoleLogger, or RemoteLogger without modifying UserService

## Final Thoughts

| Principle | Key Technique in Go           |
|----------|-------------------------------|
| SRP      | Split into packages/files      |
| OCP      | Interfaces + composition       |
| LSP      | Design interfaces carefully    |
| ISP      | Create small interfaces        |
| DIP      | Always program to interfaces   |

go’s philosophy of composition over inheritance fits well with SOLID

## Resources

- https://go.dev/blog/interfaces
- Clean Architecture by Robert C. Martin
- https://go-proverbs.github.io/
