package subpub

import (
	"context"
	"errors"
	"sync"
)

// MessageHandler is a callback function that processes messages delivered
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscrib
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscriber on the given
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the given subject.
	Publish(subject string, msg interface{}) error

	// Close will shutdown sub-pub system.
	// May be blocked by data delivery until the context is canceled.
	Close(ctx context.Context) error
}

func NewSubPub() SubPub {
	return &bus{
		subs: make(map[string]map[*subscriber]struct{}),
	}
}

type bus struct {
	mu      sync.RWMutex
	subs    map[string]map[*subscriber]struct{}
	closing bool
	wg      sync.WaitGroup
}

type subscriber struct {
	cb     MessageHandler
	mu     sync.Mutex
	cond   *sync.Cond
	queue  []interface{}
	closed bool
}

type subscription struct {
	b       *bus
	subject string
	s       *subscriber
	once    sync.Once
}

func (b *bus) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	b.mu.RLock()
	if b.closing {
		b.mu.RUnlock()
		return nil, errors.New("bus is closing")
	}
	b.mu.RUnlock()

	s := &subscriber{cb: cb}
	s.cond = sync.NewCond(&s.mu)

	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()
		return nil, errors.New("bus is closing")
	}
	if b.subs[subject] == nil {
		b.subs[subject] = make(map[*subscriber]struct{})
	}
	b.subs[subject][s] = struct{}{}
	b.wg.Add(1)
	b.mu.Unlock()

	go s.start(&b.wg)

	return &subscription{b: b, subject: subject, s: s}, nil
}

func (b *bus) Publish(subject string, msg interface{}) error {
	b.mu.RLock()
	if b.closing {
		b.mu.RUnlock()
		return errors.New("bus is closing")
	}
	subs := b.subs[subject]
	b.mu.RUnlock()

	for s := range subs {
		s.enqueue(msg)
	}
	return nil
}

func (b *bus) Close(ctx context.Context) error {
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()
		return errors.New("already closing")
	}
	b.closing = true

	for _, m := range b.subs {
		for s := range m {
			s.close()
		}
	}
	b.mu.Unlock()

	ch := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(ch)
	}()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *subscriber) start(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		s.mu.Lock()
		for !s.closed && len(s.queue) == 0 {
			s.cond.Wait()
		}
		if s.closed && len(s.queue) == 0 {
			s.mu.Unlock()
			return
		}
		msg := s.queue[0]
		s.queue = s.queue[1:]
		s.mu.Unlock()

		s.cb(msg)
	}
}

func (s *subscriber) enqueue(msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.queue = append(s.queue, msg)
	s.cond.Signal()
}

func (s *subscriber) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	s.cond.Signal()
}

func (x *subscription) Unsubscribe() {
	x.once.Do(func() {
		x.b.mu.Lock()
		if subs := x.b.subs[x.subject]; subs != nil {
			delete(subs, x.s)
		}
		x.b.mu.Unlock()
		x.s.close()
	})
}
