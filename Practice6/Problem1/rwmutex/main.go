package main

import (
"fmt"
"sync"
)

func main() {
unsafeMap := make(map[string]int)
var mu sync.RWMutex
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
wg.Add(1)
go func(val int) {
defer wg.Done()
mu.Lock()
unsafeMap["key"] = val
mu.Unlock()
}(i)
}

wg.Wait()

mu.RLock()
value := unsafeMap["key"]
mu.RUnlock()

fmt.Printf("Value: %d\n", value)
}
