package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Chamistery/subpub/pkg/subpub"
)

func main() {
	bus := subpub.NewSubPub()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	defer bus.Close(context.Background())

	// подписываем два подписчика
	bus.Subscribe("greeting", func(msg interface{}) {
		fmt.Printf("Subscriber1 got: %v\n", msg)
	})
	bus.Subscribe("greeting", func(msg interface{}) {
		fmt.Printf("Subscriber2 got: %v\n", msg)
	})

	// публикуем несколько сообщений
	for i := 1; i <= 5; i++ {
		bus.Publish("greeting", fmt.Sprintf("hello %d", i))
		time.Sleep(100 * time.Millisecond)
	}

	// позволяем доставить данные и закрываем шину
	<-ctx.Done()
	fmt.Println("Shutting down...")
}
