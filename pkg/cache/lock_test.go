package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestWithLockSerializesConcurrentWork(t *testing.T) {
	store := Store{Directory: t.TempDir()}
	key, err := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	if err != nil {
		t.Fatal(err)
	}
	var mutex sync.Mutex
	active, maximum := 0, 0
	work := func() error {
		mutex.Lock()
		active++
		if active > maximum {
			maximum = active
		}
		mutex.Unlock()
		time.Sleep(20 * time.Millisecond)
		mutex.Lock()
		active--
		mutex.Unlock()
		return nil
	}
	var group sync.WaitGroup
	for range 3 {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := store.WithLock(context.Background(), key, time.Minute, work); err != nil {
				t.Errorf("locked work failed: %v", err)
			}
		}()
	}
	group.Wait()
	if maximum != 1 {
		t.Fatalf("cache work overlapped: maximum=%d", maximum)
	}
}

func TestWithLockHonorsCancellation(t *testing.T) {
	store := Store{Directory: t.TempDir()}
	key, _ := Key(KeyInput{ScannerID: "pattern", ScannerVersion: "1"})
	entered := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = store.WithLock(context.Background(), key, time.Minute, func() error {
			close(entered)
			<-release
			return nil
		})
	}()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := store.WithLock(ctx, key, time.Minute, func() error { return nil })
	close(release)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock did not honor cancellation: %v", err)
	}
}
