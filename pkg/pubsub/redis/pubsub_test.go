package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/pkg/pubsub"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testPayload struct {
	Message string `json:"message"`
	Value   int    `json:"value"`
}

var testRedis *redis.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisC, err := testcontainers.Run(
		ctx, "redis:7-alpine",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("6379/tcp"),
			wait.ForLog("Ready to accept connections"),
		),
	)
	if err != nil {
		log.Fatalf("failed to start redis container: %v", err)
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get redis host: %v", err)
	}
	port, err := redisC.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("failed to get redis port: %v", err)
	}

	testRedis = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	if err := testRedis.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	code := m.Run()

	testRedis.Close()
	redisC.Terminate(ctx)
	os.Exit(code)
}

func setupMultiplexor(t *testing.T, rdb *redis.Client, streamName string) (*RedisMultiplexor, context.Context) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = zerolog.Nop().WithContext(ctx)
	t.Cleanup(cancel)

	mux := NewRedisMultiplexor(rdb, streamName)
	mux.Start(ctx)

	return mux, ctx
}

func TestRedisPublisherPublish(t *testing.T) {
	streamName := "test-stream-publish"
	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	ctx := context.Background()

	channel := pubsub.Channel("comments")
	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("created"),
		Payload: testPayload{Message: "hello", Value: 1},
	}

	id, err := publisher.Publish(ctx, channel, event)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	streams, err := testRedis.XRange(ctx, streamName, "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, streams, 1)

	msg := streams[0]
	assert.Equal(t, string(id), msg.ID)
	assert.Equal(t, string(channel), msg.Values[pubsub.ChannelKey])
	assert.Equal(t, string(event.Type), msg.Values[pubsub.TypeKey])

	var payload testPayload
	err = json.Unmarshal([]byte(msg.Values[pubsub.PayloadKey].(string)), &payload)
	require.NoError(t, err)
	assert.Equal(t, event.Payload, payload)
}

func TestRedisPublisherPublishContextCancelled(t *testing.T) {
	publisher := NewRedisPublisher[testPayload](testRedis, "test-stream-cancel")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("created"),
		Payload: testPayload{Message: "hello", Value: 1},
	}

	_, err := publisher.Publish(ctx, pubsub.Channel("comments"), event)
	assert.Error(t, err)
}

func TestE2EPublishSubscribe(t *testing.T) {
	streamName := "test-stream-e2e"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channel := pubsub.Channel("comments")
	sub, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)
	defer sub.Close()

	time.Sleep(200 * time.Millisecond)

	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("created"),
		Payload: testPayload{Message: "hello", Value: 42},
	}

	id, err := publisher.Publish(ctx, channel, event)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	select {
	case received := <-sub.C():
		assert.Equal(t, event.Type, received.Type)
		assert.Equal(t, event.Payload, received.Payload)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestE2EDifferentChannels(t *testing.T) {
	streamName := "test-stream-channels"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channelA := pubsub.Channel("channel-a")
	channelB := pubsub.Channel("channel-b")

	subA, err := subscriber.Subscribe(ctx, channelA)
	require.NoError(t, err)
	defer subA.Close()

	subB, err := subscriber.Subscribe(ctx, channelB)
	require.NoError(t, err)
	defer subB.Close()

	time.Sleep(200 * time.Millisecond)

	eventA := pubsub.Event[testPayload]{
		Type:    pubsub.Type("event-a"),
		Payload: testPayload{Message: "for a", Value: 1},
	}

	_, err = publisher.Publish(ctx, channelA, eventA)
	require.NoError(t, err)

	select {
	case received := <-subA.C():
		assert.Equal(t, eventA.Type, received.Type)
		assert.Equal(t, eventA.Payload, received.Payload)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event on channel A")
	}

	select {
	case received := <-subB.C():
		t.Fatalf("subscriber B should not receive event for channel A, got %+v", received)
	case <-time.After(500 * time.Millisecond):
	}

	eventB := pubsub.Event[testPayload]{
		Type:    pubsub.Type("event-b"),
		Payload: testPayload{Message: "for b", Value: 2},
	}

	_, err = publisher.Publish(ctx, channelB, eventB)
	require.NoError(t, err)

	select {
	case received := <-subB.C():
		assert.Equal(t, eventB.Type, received.Type)
		assert.Equal(t, eventB.Payload, received.Payload)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event on channel B")
	}
}

func TestE2EMultipleSubscribers(t *testing.T) {
	streamName := "test-stream-fanout"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channel := pubsub.Channel("broadcast")

	sub1, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)
	defer sub1.Close()

	sub2, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)
	defer sub2.Close()

	time.Sleep(200 * time.Millisecond)

	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("broadcast"),
		Payload: testPayload{Message: "to all", Value: 99},
	}

	_, err = publisher.Publish(ctx, channel, event)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(2)

	var received1, received2 pubsub.Event[testPayload]

	go func() {
		defer wg.Done()
		select {
		case received1 = <-sub1.C():
		case <-time.After(5 * time.Second):
			t.Error("timeout waiting for event on sub1")
		}
	}()

	go func() {
		defer wg.Done()
		select {
		case received2 = <-sub2.C():
		case <-time.After(5 * time.Second):
			t.Error("timeout waiting for event on sub2")
		}
	}()

	wg.Wait()

	assert.Equal(t, event.Type, received1.Type)
	assert.Equal(t, event.Payload, received1.Payload)
	assert.Equal(t, event.Type, received2.Type)
	assert.Equal(t, event.Payload, received2.Payload)
}

func TestSubscriptionClose(t *testing.T) {
	streamName := "test-stream-close"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channel := pubsub.Channel("comments")
	sub, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("created"),
		Payload: testPayload{Message: "before close", Value: 1},
	}

	_, err = publisher.Publish(ctx, channel, event)
	require.NoError(t, err)

	select {
	case <-sub.C():
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for event before close")
	}

	err = sub.Close()
	assert.NoError(t, err)

	event2 := pubsub.Event[testPayload]{
		Type:    pubsub.Type("created"),
		Payload: testPayload{Message: "after close", Value: 2},
	}

	_, err = publisher.Publish(context.Background(), channel, event2)
	require.NoError(t, err)

	select {
	case received, ok := <-sub.C():
		if ok {
			t.Fatalf("received event after close: %+v", received)
		}
	case <-time.After(500 * time.Millisecond):
	}
}

func TestSubscriptionContextCancellation(t *testing.T) {
	streamName := "test-stream-cancel-sub"
	mux, _ := setupMultiplexor(t, testRedis, streamName)

	subCtx, cancel := context.WithCancel(context.Background())
	subCtx = zerolog.Nop().WithContext(subCtx)

	subscriber := NewMuxSubscriber[testPayload](mux)
	sub, err := subscriber.Subscribe(subCtx, pubsub.Channel("comments"))
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	cancel()
	time.Sleep(200 * time.Millisecond)

	assert.ErrorIs(t, sub.Err(), context.Canceled)
}

func TestMuxSubscriberCustomBufferSize(t *testing.T) {
	streamName := "test-stream-buffer"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channel := pubsub.Channel("comments")
	customSize := 5

	bufOpt := func(opts *pubsub.SubOptions) {
		opts.BufferSize = customSize
	}

	sub, err := subscriber.Subscribe(ctx, channel, bufOpt)
	require.NoError(t, err)
	defer sub.Close()

	time.Sleep(200 * time.Millisecond)

	for i := 0; i < customSize; i++ {
		event := pubsub.Event[testPayload]{
			Type:    pubsub.Type("tick"),
			Payload: testPayload{Message: fmt.Sprintf("msg-%d", i), Value: i},
		}
		_, err := publisher.Publish(ctx, channel, event)
		require.NoError(t, err)
	}

	receivedCount := 0
	timeout := time.After(10 * time.Second)

	for receivedCount < customSize {
		select {
		case <-sub.C():
			receivedCount++
		case <-timeout:
			t.Fatalf("received %d events, expected %d", receivedCount, customSize)
		}
	}

	assert.Equal(t, customSize, receivedCount)
}

func TestRawEventFormat(t *testing.T) {
	streamName := "test-stream-raw"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)

	channel := pubsub.Channel("raw-channel")
	rawCh, unsubscribe := mux.subscribe(string(channel))
	defer unsubscribe()

	time.Sleep(200 * time.Millisecond)

	event := pubsub.Event[testPayload]{
		Type:    pubsub.Type("raw-event"),
		Payload: testPayload{Message: "raw", Value: 7},
	}

	_, err := publisher.Publish(ctx, channel, event)
	require.NoError(t, err)

	payloadBytes, _ := json.Marshal(event.Payload)

	select {
	case raw := <-rawCh:
		assert.Equal(t, string(channel), raw.Channel)
		assert.Equal(t, string(event.Type), raw.Type)
		assert.JSONEq(t, string(payloadBytes), string(raw.Payload))
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for raw event")
	}
}

func TestMuxSubscriberUnsubscribeRemovesListener(t *testing.T) {
	streamName := "test-stream-unsub"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	subscriber := NewMuxSubscriber[testPayload](mux)
	channel := pubsub.Channel("unsub-channel")

	sub, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	mux.mu.RLock()
	listenerCount := len(mux.listeners[string(channel)])
	mux.mu.RUnlock()
	assert.Equal(t, 1, listenerCount)

	sub.Close()
	time.Sleep(200 * time.Millisecond)

	mux.mu.RLock()
	listenerCount = len(mux.listeners[string(channel)])
	mux.mu.RUnlock()
	assert.Equal(t, 0, listenerCount)
}

func TestMultiplePublishEvents(t *testing.T) {
	streamName := "test-stream-multi"
	mux, ctx := setupMultiplexor(t, testRedis, streamName)

	publisher := NewRedisPublisher[testPayload](testRedis, streamName)
	subscriber := NewMuxSubscriber[testPayload](mux)

	channel := pubsub.Channel("multi")
	sub, err := subscriber.Subscribe(ctx, channel)
	require.NoError(t, err)
	defer sub.Close()

	time.Sleep(200 * time.Millisecond)

	eventCount := 10
	for i := 0; i < eventCount; i++ {
		event := pubsub.Event[testPayload]{
			Type:    pubsub.Type("multi-event"),
			Payload: testPayload{Message: fmt.Sprintf("msg-%d", i), Value: i},
		}
		_, err := publisher.Publish(ctx, channel, event)
		require.NoError(t, err)
	}

	for i := 0; i < eventCount; i++ {
		select {
		case received := <-sub.C():
			assert.Equal(t, pubsub.Type("multi-event"), received.Type)
			assert.Equal(t, i, received.Payload.Value)
			assert.Equal(t, fmt.Sprintf("msg-%d", i), received.Payload.Message)
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}
}
