
# SOLID

## Solid Basics:

1. Single Responsibility Principle  
2. Open-Closed Principle  
3. Liskov Substitution Principle  
4. Interface Segregation Principle  
5. Dependency Inversion Principle  

---

## 1. Single Responsibility Principle

> "A class should have only one reason to change."

- A class should do one thing and do it well.  
- It keeps code modular and easier to understand/test.  
- When many different teams work on the same project and edit the same class for different reasons, it can lead to incompatible modules.  
- It makes version control easier.  
  - For example: A persistence class that handles database operations, if changed, clearly indicates the change is storage-related.  
- **Merge Conflicts:**  
  - These occur when different teams change the same file.  
  - Following SRP reduces such conflicts, as files will have a single reason to change.

---

## 2. Open/Closed Principle

> "Software entities should be open for extension, but closed for modification."

- You should be able to add new features without changing existing code.  
- A class can be both open (for extension) and closed (for modification).  
- Use abstraction (interfaces/abstract classes) and polymorphism.  
- Patterns that support this principle:  
  - Strategy pattern  
  - Plugins  
  - Inheritance  
  - Dependency Injection  

### Strategy Pattern

The strategy design pattern allows changing an algorithm/behavior at runtime without modifying the code that uses it.

**Structure:**  
- Uses a strategy to perform a task  
- Common interface for all algorithms  
- Different implementations of the strategy  

**Use cases:**  
- You have many variations of an algorithm → Keep logic modular  
- You want to switch behavior at runtime  
- Adhere to the Open/Closed Principle  

**Advantages:**  
- Removes if-else/switch-case blocks  
- Adds behavior via composition, not inheritance  
- Keeps code clean, flexible, and testable  
- Works great with dependency injection  

---

## 3. Liskov Substitution Principle

> "Subtypes must be substitutable for their base types."

- Any subclass should work perfectly in place of its parent class.  
- Prevents unexpected behaviors when using inheritance.  
- If class B is a subclass of class A, you should be able to pass an object of B to any method expecting A, and it should work correctly.  
- The child class should extend behavior but not narrow it down.  
- Violating this principle can lead to difficult-to-detect bugs.

---

## 4. Interface Segregation Principle

> "Clients should not be forced to depend on methods they do not use."

- Prefer small, focused interfaces over large, general-purpose ones.  
- Encourages cohesion and decoupling.  
- Many client-specific interfaces are better than one general-purpose interface.  
- Clients should not be forced to implement methods they don’t need.

---

## 5. Dependency Inversion Principle

> "High-level modules should not depend on low-level modules. Both should depend on abstractions."

- Use abstractions (interfaces) instead of direct dependencies.  
- **Inversion of Control:** Facilitates testing, modularity, and flexibility.

---

## SOLID in Go Examples

🎥 Video: [SOLID in Go](https://youtu.be/o_yTAosQUGc?si=M3djQFVy6plQjUIB)  
💻 Code: [GitHub - packagemain/solid](https://github.com/plutov/packagemain/tree/master/solid)

### 1. Single Responsibility Principle

Problem: Mixing database logic and survey logic in the same struct (`save()`).

**Fix:**  
Introduce a repository interface to move DB logic to another package/file.

---

### 2. Open/Closed Principle

Problem: Adding new export destinations requires modifying existing code, which may introduce bugs.

**Fix:**  
Use an `Exporter` interface.

```go
type Exporter interface {
    Export(*Survey) error
}

type S3Exporter struct{}
func (e *S3Exporter) Export(s *Survey) error {
    return nil
}

type GCSExporter struct{}
func (e *GCSExporter) Export(s *Survey) error {
    return nil
}
```

**Example Usage:**

```go
s := &Survey{}
var exporter Exporter

if useS3 {
    exporter = &S3Exporter{}
} else {
    exporter = &GCSExporter{}
}

err := ExportSurvey(s, exporter)
```

---

## Additional Resources

- [SOLID Principle Simplified – Medium](https://medium.com/@shubhadeepchat/solid-principle-simplified-b18b73b3e440)  
- *Dive Into Design Patterns* – Book
