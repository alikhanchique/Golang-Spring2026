package main

import (
"fmt"
"sync"
)

func main() {
var safeMap sync.Map
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
wg.Add(1)
go func(val int) {
defer wg.Done()
safeMap.Store("key", val)
}(i)
}

wg.Wait()

value, _ := safeMap.Load("key")
fmt.Printf("Value: %d\n", value)
}
