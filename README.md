# go-deq

Double Evented Queue


```go
func main() {

  unsub := queue2.Subscribe(func(s queue2.Item) {
    fmt.Println("got message", s.(string))
  })

  go func() {
    time.Sleep(10 * time.Second)
    unsub()
    queue2.Subscribe(func(s queue2.Item) {
      fmt.Println("[resub] got message", s.(string))
    })
  }()

  queue2.PCap = 5
  // call on queue2.LCap limit
  queue2.PC = func(c int) {
    fmt.Println("Primary queue limit reached: ", c)
  }

  queue2.SCap = 3
  // call on queue2.GCap limit
  queue2.SC = func(c int) {
    fmt.Println("Secondary queue limit reached: ", c)
  }
 

 // capture realtime traffic
  go func() {
    for {
      time.Sleep(1 * time.Second)
      queue2.WriteSecondary(fmt.Sprintf("global-%d", time.Now().Unix()))
    }
  }()

  // read from local file
  go func() {
    for i := 0; i <= 10; i++ {
      time.Sleep(500 * time.Millisecond)
      queue2.WritePrimary(fmt.Sprintf("local-%d", i))

    }
    // local file parse done, switch to secondary queue
    queue2.Switch()
  }()

  // run forever
  runtime.Goexit()
}
```


```go
func main() {
  var lt = queue2.New()

  lt.WritePrimary(func() { println("fn 1") })
  lt.WritePrimary(func() { println("fn 2") })
  lt.WritePrimary(func() { println("fn 3") })

  for i, e := lt.ProcessPrimary(); e == nil; i, e = lt.ProcessPrimary() {
    i.(func())()
  }
}

```
