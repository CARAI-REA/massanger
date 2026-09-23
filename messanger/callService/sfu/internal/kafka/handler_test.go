package kafka_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/CARAI-REA/messanger/callService/platform/pkg/logger"
	"github.com/stretchr/testify/require"

	"sfu/internal/kafka"
	"sfu/internal/model"
	"sfu/internal/service"
)

func TestMain(m *testing.M) {
	_ = logger.Init("error", false)
	os.Exit(m.Run())
}

type fakeSessions struct {
	mu         sync.Mutex
	closedRoom string
	closedUser struct{ room, user string }
}

func (f *fakeSessions) Connect(context.Context, service.Conn, string, string) error { return nil }
func (f *fakeSessions) Handle(context.Context, string, string, model.Envelope) error {
	return nil
}
func (f *fakeSessions) Disconnect(context.Context, service.Conn, string, string, string) error {
	return nil
}
func (f *fakeSessions) CloseRoom(_ context.Context, roomUUID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closedRoom = roomUUID
	return nil
}
func (f *fakeSessions) CloseUser(_ context.Context, roomUUID, userUUID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closedUser.room = roomUUID
	f.closedUser.user = userUUID
	return nil
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	b, err := os.ReadFile(filepath.Join(filepath.Dir(file), "fixtures", name))
	require.NoError(t, err)
	return b
}

func TestK1_RoomEnded(t *testing.T) {
	f := &fakeSessions{}
	h := kafka.NewHandler(f)
	require.NoError(t, h.Handle(context.Background(), fixture(t, "room_ended.json")))
	require.Equal(t, "room-1", f.closedRoom)
}

func TestK3_ParticipantLeft(t *testing.T) {
	f := &fakeSessions{}
	h := kafka.NewHandler(f)
	require.NoError(t, h.Handle(context.Background(), fixture(t, "participant_left.json")))
	require.Equal(t, "room-3", f.closedUser.room)
	require.Equal(t, "user-9", f.closedUser.user)
}

func TestK5_Malformed(t *testing.T) {
	f := &fakeSessions{}
	h := kafka.NewHandler(f)
	require.NoError(t, h.Handle(context.Background(), []byte(`{bad`)))
}
