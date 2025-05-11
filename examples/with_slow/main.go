package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Chamistery/subpub/pkg/subpub"
)

func main() {
	bus := subpub.NewSubPub()
	defer bus.Close(context.Background())

	// Подписчик 1 — медленный
	bus.Subscribe("topic-slow", func(msg interface{}) {
		fmt.Printf("[Slow Subscriber] received: %v, processing...", msg)
		time.Sleep(200 * time.Millisecond)
		fmt.Println("[Slow Subscriber] done processing")
	})

	// Подписчик 2 — быстрый
	bus.Subscribe("topic-slow", func(msg interface{}) {
		fmt.Printf("[Fast Subscriber] received: %v", msg)
	})

	// Публикуем несколько сообщений
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("message %d", i)
		fmt.Printf("Publishing: %s", msg)
		bus.Publish("topic-slow", msg)
		time.Sleep(50 * time.Millisecond)
	}

	// Ждём, чтобы все обработчики завершили работу
	time.Sleep(1 * time.Second)
	fmt.Println("Done.")
}
