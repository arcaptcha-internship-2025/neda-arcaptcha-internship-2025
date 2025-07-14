
# faking in go testing

---

## what is a fake?

a fake is an alternative implementation of a dependency that behaves like the real one but is simpler or more controllable. fakes are typically written by hand or generated, and are useful when:

- you want to simulate complex behavior (e.g., concurrent workers, db queries)
- you need stateful responses
- you prefer realistic behavior over rigid expectations (as with mocks)

---

## hand-writing fakes within structs

you can use function fields in a struct to override behavior during tests.

### example

```go
type saver struct {
    SaveFunc func(data string) error
}

func (s *saver) Save(data string) error {
    return s.SaveFunc(data)
}

func TestSaver(t *testing.T) {
    fake := &saver{
        SaveFunc: func(data string) error {
            if data == "" {
                return errors.New("empty data")
            }
            return nil
        },
    }

    err := fake.Save("")
    if err == nil {
        t.Error("expected error for empty data")
    }
}
```

source: [johnsiilver on medium](https://medium.com/@johnsiilver/go-faking-method-calls-within-a-struct-2bf3e5af9e29)

---

## faking database methods

define an interface and create a fake implementation that satisfies it.

### interface and fake

```go
type db interface {
    query(id string) (string, error)
}

type fakeDB struct{}

func (f *fakeDB) query(id string) (string, error) {
    if id == "123" {
        return "user123", nil
    }
    return "", errors.New("not found")
}
```

### test usage

```go
func TestGetUser(t *testing.T) {
    fake := &fakeDB{}
    result, err := fake.query("123")
    if err != nil || result != "user123" {
        t.Fail()
    }
}
```

source: [stackoverflow](https://stackoverflow.com/questions/49327497/faking-db-methods-in-golang)

---

## using channels in fakes

channels make fakes powerful when testing concurrent systems or pipelines.

### example: faking a worker

```go
type job struct {
    id int
}

type fakeQueue struct {
    jobs chan job
}

func (f *fakeQueue) enqueue(j job) {
    f.jobs <- j
}

func (f *fakeQueue) process() job {
    return <-f.jobs
}

func TestQueue(t *testing.T) {
    q := &fakeQueue{jobs: make(chan job, 1)}
    q.enqueue(job{id: 42})

    got := q.process()
    if got.id != 42 {
        t.Fail()
    }
}
```

source: [cloudbees blog](https://www.cloudbees.com/blog/creating-fakes-in-go-with-channels)

---

## generating fake data

use libraries like `gofakeit` to populate your fakes with realistic data.

### install

```bash
go get github.com/brianvoe/gofakeit/v6
```

### example

```go
import "github.com/brianvoe/gofakeit/v6"

func TestWithFakeData(t *testing.T) {
    name := gofakeit.Name()
    email := gofakeit.Email()
    t.Logf("fake name: %s, fake email: %s", name, email)
}
```

source: [ankit malik on dev.to](https://dev.to/ankitmalikg/golang-generate-fake-data-with-gofakeit-23gj)

---

## when to use fakes vs mocks

| criteria       | fake                                 | mock                                 |
|----------------|--------------------------------------|--------------------------------------|
| structure      | implemented manually or via channels | often generated with testify/mockery |
| behavior       | simulates real behavior              | focuses on interactions and calls    |
| use case       | integration-style tests              | unit tests                           |
| data handling  | maintains state                      | predefined expectations              |

---

## conclusion

fakes are a powerful way to simulate real behavior in tests, especially when dealing with stateful logic, concurrent systems, or complex workflows. go’s simple interface system and support for channels make it easy to create effective fakes.

combine fakes with real-looking data from libraries like `gofakeit` to write more comprehensive and realistic tests.

---

## references

- john siilver: [go faking method calls within a struct](https://medium.com/@johnsiilver/go-faking-method-calls-within-a-struct-2bf3e5af9e29)
- stackoverflow: [faking db methods in golang](https://stackoverflow.com/questions/49327497/faking-db-methods-in-golang)
- cloudbees: [creating fakes in go with channels](https://www.cloudbees.com/blog/creating-fakes-in-go-with-channels)
- ankit malik: [generate fake data with gofakeit](https://dev.to/ankitmalikg/golang-generate-fake-data-with-gofakeit-23gj)
