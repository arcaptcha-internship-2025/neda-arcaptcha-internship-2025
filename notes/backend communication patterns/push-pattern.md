# push notifications and push model in go 

## introduction

the push model is a communication pattern where the server actively sends data to clients without them explicitly requesting it. unlike traditional request-response patterns where clients must poll for updates, push notifications enable real-time communication by maintaining persistent connections and instantly delivering updates when events occur.

## understanding the push model

### what is the push model?

in the push model, servers proactively send data to connected clients. the only requirement is an established connection between client and server. this pattern excels in scenarios requiring:

- real-time messaging and chat applications
- live notifications and alerts  
- real-time dashboards and monitoring
- collaborative applications
- gaming and interactive systems

### how push works

1. **connection establishment**: client establishes a long-lived connection to the server
2. **event occurrence**: when something happens on the server (new message, update, etc.)
3. **immediate push**: server instantly pushes data to all connected clients
4. **bidirectional communication**: many push implementations support two-way communication

### push vs polling comparison

| aspect | push model | polling model |
|--------|------------|---------------|
| latency | immediate updates | delayed by polling interval |
| server load | constant connection overhead | repeated request overhead |
| network efficiency | high (only sends when needed) | low (constant checking) |
| real-time capability | excellent | poor to moderate |
| implementation complexity | moderate to high | simple |

## push technologies overview

### websockets
- full-duplex communication over single tcp connection
- low overhead after initial handshake
- supported by all modern browsers
- ideal for real-time bidirectional communication

### server-sent events (sse)
- unidirectional push from server to client
- built on http, simpler than websockets
- automatic reconnection handling
- perfect for live feeds and notifications

### message queues
- asynchronous messaging between services
- supports publish-subscribe patterns
- examples: rabbitmq, apache kafka, redis pub/sub

### push notification services
- apple push notification service (apns)
- firebase cloud messaging (fcm)  
- web push protocol

## implementing push in go

### websocket implementation

```go
package main

import (
    "log"
    "net/http"
    "time"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // allow all origins in development
    },
}

type hub struct {
    clients    map[*client]bool
    broadcast  chan []byte
    register   chan *client
    unregister chan *client
}

type client struct {
    hub  *hub
    conn *websocket.Conn
    send chan []byte
}

func newHub() *hub {
    return &hub{
        clients:    make(map[*client]bool),
        broadcast:  make(chan []byte),
        register:   make(chan *client),
        unregister: make(chan *client),
    }
}

func (h *hub) run() {
    for {
        select {
        case client := <-h.register:
            h.clients[client] = true
            log.println("client connected")

        case client := <-h.unregister:
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                log.println("client disconnected")
            }

        case message := <-h.broadcast:
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
        }
    }
}

func (c *client) writePump() {
    ticker := time.newticker(54 * time.second)
    defer func() {
        ticker.stop()
        c.conn.close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.setwritedeadline(time.now().add(10 * time.second))
            if !ok {
                c.conn.writemessage(websocket.closemessage, []byte{})
                return
            }

            if err := c.conn.writemessage(websocket.textmessage, message); err != nil {
                log.println("write error:", err)
                return
            }

        case <-ticker.c:
            c.conn.setwritedeadline(time.now().add(10 * time.second))
            if err := c.conn.writemessage(websocket.pingmessage, nil); err != nil {
                return
            }
        }
    }
}

func (c *client) readpump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.close()
    }()

    c.conn.setreaddeadline(time.now().add(60 * time.second))
    c.conn.setponghandler(func(string) error {
        c.conn.setreaddeadline(time.now().add(60 * time.second))
        return nil
    })

    for {
        _, message, err := c.conn.readmessage()
        if err != nil {
            break
        }
        
        // broadcast message to all clients
        c.hub.broadcast <- message
    }
}

func servewebsocket(hub *hub, w http.responsewriter, r *http.request) {
    conn, err := upgrader.upgrade(w, r, nil)
    if err != nil {
        log.println("websocket upgrade error:", err)
        return
    }

    client := &client{
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 256),
    }

    client.hub.register <- client

    go client.writepump()
    go client.readpump()
}

func main() {
    hub := newhub()
    go hub.run()

    http.handlefunc("/ws", func(w http.responsewriter, r *http.request) {
        servewebsocket(hub, w, r)
    })

    log.println("websocket server starting on :8080")
    log.fatal(http.listenandserve(":8080", nil))
}
```

### server-sent events implementation

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"
)

type eventstream struct {
    clients map[chan string]bool
    addclient chan chan string
    removeclient chan chan string
    messages chan string
}

func neweventstream() *eventstream {
    return &eventstream{
        clients: make(map[chan string]bool),
        addclient: make(chan chan string),
        removeclient: make(chan chan string),
        messages: make(chan string),
    }
}

func (es *eventstream) run() {
    for {
        select {
        case client := <-es.addclient:
            es.clients[client] = true
            log.println("client connected to sse")

        case client := <-es.removeclient:
            delete(es.clients, client)
            close(client)
            log.println("client disconnected from sse")

        case message := <-es.messages:
            for client := range es.clients {
                select {
                case client <- message:
                default:
                    delete(es.clients, client)
                    close(client)
                }
            }
        }
    }
}

func (es *eventstream) servehttp(w http.responsewriter, r *http.request) {
    w.header().set("content-type", "text/event-stream")
    w.header().set("cache-control", "no-cache")
    w.header().set("connection", "keep-alive")
    w.header().set("access-control-allow-origin", "*")

    client := make(chan string)
    es.addclient <- client

    defer func() {
        es.removeclient <- client
    }()

    notify := r.context().done()
    go func() {
        <-notify
        es.removeclient <- client
    }()

    for {
        select {
        case message := <-client:
            fmt.fprintf(w, "data: %s\n\n", message)
            w.(http.flusher).flush()
        case <-notify:
            return
        }
    }
}

func main() {
    stream := neweventstream()
    go stream.run()

    // simulate sending periodic messages
    go func() {
        ticker := time.newticker(2 * time.second)
        defer ticker.stop()
        
        counter := 0
        for range ticker.c {
            counter++
            message := fmt.sprintf("update %d at %s", counter, time.now().format("15:04:05"))
            stream.messages <- message
        }
    }()

    http.handle("/events", stream)
    
    // serve static html for testing
    http.handlefunc("/", func(w http.responsewriter, r *http.request) {
        html := `
<!doctype html>
<html>
<head>
    <title>server-sent events demo</title>
</head>
<body>
    <h1>server-sent events</h1>
    <div id="messages"></div>
    <script>
        const eventsource = new eventsource('/events');
        eventsource.onmessage = function(event) {
            const div = document.createelement('div');
            div.innerhtml = event.data;
            document.getelementbyid('messages').appendchild(div);
        };
    </script>
</body>
</html>`
        w.header().set("content-type", "text/html")
        w.write([]byte(html))
    })

    log.println("sse server starting on :8080")
    log.fatal(http.listenandserve(":8080", nil))
}
```

### redis pub/sub implementation

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

type redispubsub struct {
    client *redis.client
    ctx    context.context
}

func newredispubsub(addr string) *redispubsub {
    rdb := redis.newclient(&redis.options{
        addr: addr,
    })

    return &redispubsub{
        client: rdb,
        ctx:    context.background(),
    }
}

func (r *redispubsub) publish(channel, message string) error {
    return r.client.publish(r.ctx, channel, message).err()
}

func (r *redispubsub) subscribe(channels ...string) *redis.pubsub {
    return r.client.subscribe(r.ctx, channels...)
}

func main() {
    pubsub := newredispubsub("localhost:6379")

    // subscriber goroutine
    go func() {
        subscriber := pubsub.subscribe("notifications", "alerts")
        defer subscriber.close()

        ch := subscriber.channel()
        for msg := range ch {
            log.printf("received message from %s: %s", msg.channel, msg.payload)
        }
    }()

    // publisher goroutine
    go func() {
        ticker := time.newticker(3 * time.second)
        defer ticker.stop()

        counter := 0
        for range ticker.c {
            counter++
            message := fmt.sprintf("notification %d", counter)
            
            if err := pubsub.publish("notifications", message); err != nil {
                log.printf("failed to publish: %v", err)
            } else {
                log.printf("published: %s", message)
            }
        }
    }()

    select {} // keep main running
}
```

### push notification service integration

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type fcmmessage struct {
    to   string                 `json:"to"`
    data map[string]interface{} `json:"data"`
}

type pushnotificationservice struct {
    serverkey string
    fcmurl    string
}

func newpushservice(serverkey string) *pushnotificationservice {
    return &pushnotificationservice{
        serverkey: serverkey,
        fcmurl:    "https://fcm.googleapis.com/fcm/send",
    }
}

func (p *pushnotificationservice) sendnotification(devicetoken string, data map[string]interface{}) error {
    message := fcmmessage{
        to:   devicetoken,
        data: data,
    }

    jsondata, err := json.marshal(message)
    if err != nil {
        return err
    }

    req, err := http.newrequest("post", p.fcmurl, bytes.newbuffer(jsondata))
    if err != nil {
        return err
    }

    req.header.set("authorization", "key="+p.serverkey)
    req.header.set("content-type", "application/json")

    client := &http.client{}
    resp, err := client.do(req)
    if err != nil {
        return err
    }
    defer resp.body.close()

    if resp.statuscode != http.statusok {
        return fmt.errorf("fcm request failed with status: %d", resp.statuscode)
    }

    return nil
}

func main() {
    pushservice := newpushservice("your-fcm-server-key")

    data := map[string]interface{}{
        "title": "new message",
        "body":  "you have received a new message",
    }

    err := pushservice.sendnotification("device-token-here", data)
    if err != nil {
        log.printf("failed to send notification: %v", err)
    } else {
        log.println("notification sent successfully")
    }
}
```

## advanced patterns

### fan-out messaging

```go
type fanoutmanager struct {
    subscribers []chan message
    mutex       sync.rwmutex
}

func (f *fanoutmanager) subscribe() chan message {
    f.mutex.lock()
    defer f.mutex.unlock()
    
    ch := make(chan message, 100)
    f.subscribers = append(f.subscribers, ch)
    return ch
}

func (f *fanoutmanager) broadcast(msg message) {
    f.mutex.rlock()
    defer f.mutex.runlock()
    
    for _, sub := range f.subscribers {
        select {
        case sub <- msg:
        default:
            // subscriber channel is full, skip
        }
    }
}
```

### connection pooling

```go
type connectionpool struct {
    connections chan *websocket.conn
    factory     func() (*websocket.conn, error)
    maxsize     int
}

func newconnectionpool(maxsize int, factory func() (*websocket.conn, error)) *connectionpool {
    return &connectionpool{
        connections: make(chan *websocket.conn, maxsize),
        factory:     factory,
        maxsize:     maxsize,
    }
}

func (p *connectionpool) get() (*websocket.conn, error) {
    select {
    case conn := <-p.connections:
        return conn, nil
    default:
        return p.factory()
    }
}

func (p *connectionpool) put(conn *websocket.conn) {
    select {
    case p.connections <- conn:
    default:
        conn.close()
    }
}
```

## best practices

### connection management
- implement proper connection lifecycle management
- handle reconnection logic for client disconnections
- use connection pooling for high-throughput scenarios
- implement heartbeat/ping-pong mechanisms

### security considerations
- validate and sanitize all incoming messages
- implement authentication and authorization
- use wss:// (websocket secure) in production
- implement rate limiting to prevent abuse

### scalability patterns
- use message queues for decoupling
- implement horizontal scaling with load balancers
- consider using sticky sessions for websockets
- implement proper backpressure handling

### error handling
- implement graceful degradation
- provide fallback mechanisms
- log connection events and errors
- handle partial message delivery scenarios

### monitoring and observability
- track connection counts and message rates
- monitor memory usage and connection leaks
- implement health checks for push services
- use metrics for performance optimization

## common use cases

### real-time chat application
- websockets for bidirectional communication
- message broadcasting to multiple clients
- presence detection and typing indicators
- message persistence and history

### live notifications system
- server-sent events for one-way notifications
- push notifications for mobile devices
- email/sms fallback for offline users
- notification preferences and filtering

### collaborative editing
- operational transformation for conflict resolution
- real-time cursor position sharing
- version control and change tracking
- user presence and activity indicators

### live dashboards and monitoring
- real-time metrics and data visualization
- alert notifications for threshold breaches
- automatic data refresh without page reload
- multi-user dashboard sharing

## performance optimization

### message batching
```go
type messagebatcher struct {
    messages []message
    timer    *time.timer
    flush    func([]message)
    batchsize int
    timeout   time.duration
}

func (b *messagebatcher) add(msg message) {
    b.messages = append(b.messages, msg)
    
    if len(b.messages) >= b.batchsize {
        b.flushbatch()
    } else if b.timer == nil {
        b.timer = time.aftertimer(b.timeout, b.flushbatch)
    }
}

func (b *messagebatcher) flushbatch() {
    if len(b.messages) > 0 {
        b.flush(b.messages)
        b.messages = b.messages[:0]
    }
    
    if b.timer != nil {
        b.timer.stop()
        b.timer = nil
    }
}
```

### memory management
- use object pooling for frequently allocated structures
- implement proper cleanup for disconnected clients
- monitor goroutine leaks in connection handlers
- optimize json serialization/deserialization

### network optimization
- compress messages when beneficial
- implement message deduplication
- use binary protocols for high-frequency updates
- optimize tcp buffer sizes

## testing strategies

### unit testing
```go
func testwebsockethub(t *testing.t) {
    hub := newhub()
    go hub.run()
    
    // create mock clients
    client1 := &client{
        hub:  hub,
        send: make(chan []byte, 1),
    }
    
    client2 := &client{
        hub:  hub,
        send: make(chan []byte, 1),
    }
    
    // register clients
    hub.register <- client1
    hub.register <- client2
    
    // test message broadcasting
    testmessage := []byte("test message")
    hub.broadcast <- testmessage
    
    // verify both clients received message
    select {
    case received := <-client1.send:
        assert.equal(t, testmessage, received)
    case <-time.after(time.second):
        t.error("client1 did not receive message")
    }
    
    select {
    case received := <-client2.send:
        assert.equal(t, testmessage, received)
    case <-time.after(time.second):
        t.error("client2 did not receive message")
    }
}
```

### integration testing
- test end-to-end message flow
- verify connection handling under load
- test reconnection scenarios
- validate message ordering and delivery

### load testing
- simulate high connection counts
- test message throughput limits  
- verify memory usage under stress
- test graceful degradation scenarios

## deployment considerations

### containerization
```dockerfile
from golang:1.21-alpine as builder
workdir /app
copy go.mod go.sum ./
run go mod download
copy . .
run go build -o push-server main.go

from alpine:latest
run apk --no-cache add ca-certificates
workdir /root/
copy --from=builder /app/push-server .
cmd ["./push-server"]
```

### kubernetes deployment
```yaml
apiversion: apps/v1
kind: deployment
metadata:
  name: push-server
spec:
  replicas: 3
  selector:
    matchlabels:
      app: push-server
  template:
    metadata:
      labels:
        app: push-server
    spec:
      containers:
      - name: push-server
        image: push-server:latest
        ports:
        - containerport: 8080
        env:
        - name: redis_url
          value: "redis:6379"
---
apiversion: v1
kind: service
metadata:
  name: push-server-service
spec:
  selector:
    app: push-server
  ports:
  - port: 80
    targetport: 8080
  type: loadbalancer
```

### monitoring and logging
- implement structured logging
- use distributed tracing for debugging
- monitor connection metrics and performance
- set up alerting for critical failures

## troubleshooting

### common issues
- connection timeouts and reconnection loops
- memory leaks from unclosed connections
- message ordering issues in distributed setups
- rate limiting and backpressure problems

### debugging techniques
- use websocket debugging tools
- implement detailed connection logging
- monitor goroutine counts and memory usage
- test with simulated network failures

### performance bottlenecks
- identify cpu-intensive message processing
- optimize json marshaling/unmarshaling
- reduce memory allocations in hot paths
- profile network i/o performance

## conclusion

the push model is essential for building modern real-time applications. go provides excellent support for implementing push notifications through websockets, server-sent events, and message queue integrations. key considerations include proper connection management, security, scalability, and monitoring.

successful push implementations require careful attention to connection lifecycle management, error handling, and performance optimization. by following the patterns and best practices outlined in this guide, you can build robust and scalable push notification systems in go.

## references

1. [understanding the push model in backend systems: a practical guide](https://medium.com/@mayank66jain/understanding-the-push-model-in-backend-systems-a-practical-guide-bffc9ff280d9)
2. [communication design patterns for backend development](https://www.freecodecamp.org/news/communication-design-patterns-for-backend-development/)
