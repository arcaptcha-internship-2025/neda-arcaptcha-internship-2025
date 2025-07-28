**adapter pattern in go**

the adapter pattern is a structural design pattern that allows objects with incompatible interfaces to work together. it acts as a bridge between the client and a class with a different interface.

---

**when to use**

- when you want to use a existing class, but its interface does not match what you need.
- when you want to reuse legacy code with a new system.
- when integrating third-party libraries that cannot be changed.

---

**real-world analogy**

think of a power plug adapter. if your laptop has a us plug and the socket is eu-style, you use an adapter to make it work. the adapter does not change the devices—it just makes them compatible.

---

**types of adapters**

- **object adapter**: uses composition (preferred in go).
- **class adapter**: uses inheritance (not applicable in go due to lack of inheritance).

---

**example 1: legacy printer adapter**

```go
type Printer interface {
    Print(string) string
}

type LegacyPrinter struct{}

func (l *LegacyPrinter) PrintLegacy(msg string) string {
    return "legacy printer: " + msg
}

type PrinterAdapter struct {
    LegacyPrinter *LegacyPrinter
}

func (p *PrinterAdapter) Print(msg string) string {
    if p.LegacyPrinter != nil {
        return p.LegacyPrinter.PrintLegacy(msg)
    }
    return ""
}
```

---

**example 2: logger adapter**

a system wants to use a new logger interface, but an old logging library is still in use.

```go
// new logging interface expected
type Logger interface {
    LogInfo(message string)
}

// old logger with a different method
type OldLogger struct{}

func (o *OldLogger) WriteLog(msg string) {
    fmt.Println("old logger:", msg)
}

// adapter to match the new interface
type LoggerAdapter struct {
    Old *OldLogger
}

func (l *LoggerAdapter) LogInfo(message string) {
    l.Old.WriteLog(message)
}
```

---

**example 3: http handler adapter**

adapting a custom handler function to match `http.Handler` interface.

```go
func MyCustomHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("custom handler"))
}

// adapter to convert to http.Handler
func Adapter(fn func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(fn)
}

func main() {
    http.Handle("/custom", Adapter(MyCustomHandler))
    http.ListenAndServe(":8080", nil)
}
```

---

**example 4: payment gateway adapter**

adapting different payment gateways to a common interface.

```go
// common payment interface
type PaymentProcessor interface {
    Pay(amount float64) string
}

// paypal sdk
type PayPal struct{}

func (p *PayPal) SendPayment(amount float64) string {
    return fmt.Sprintf("paid %0.2f with paypal", amount)
}

// stripe sdk
type Stripe struct{}

func (s *Stripe) MakePayment(amount float64) string {
    return fmt.Sprintf("paid %0.2f with stripe", amount)
}

// paypal adapter
type PayPalAdapter struct {
    PayPal *PayPal
}

func (p *PayPalAdapter) Pay(amount float64) string {
    return p.PayPal.SendPayment(amount)
}

// stripe adapter
type StripeAdapter struct {
    Stripe *Stripe
}

func (s *StripeAdapter) Pay(amount float64) string {
    return s.Stripe.MakePayment(amount)
}
```

---

**advantages**

- promotes code reusability.
- helps in migrating from old interfaces to new ones without rewriting the whole system.
- decouples the client from concrete implementations.

---

**limitations**

- can lead to unnecessary complexity if overused.
- adapter logic can become complex if interfaces differ significantly.

---

**conclusion**

the adapter pattern is useful when dealing with mismatched interfaces. in go, it is implemented using composition. by introducing adapters, systems can grow and evolve while keeping old components usable.

---

**references**

- [refactoring guru - adapter pattern](https://refactoring.guru/design-patterns/adapter)  
- [refactoring guru - go example](https://refactoring.guru/design-patterns/adapter/go/example)  
- [dev.to - understanding the adapter pattern in go](https://dev.to/kittipat1413/understanding-the-adapter-pattern-in-go-2mln)  
- [dev.to - adapter design pattern in go](https://dev.to/ansu/adapter-design-pattern-in-go-3hbm)
