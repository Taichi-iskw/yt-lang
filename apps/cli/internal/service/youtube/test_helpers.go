package youtube

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Taichi-iskw/yt-lang/internal/model"
	"github.com/Taichi-iskw/yt-lang/internal/service/common"
)

// mockCmdRunner is a mock implementation of CmdRunner for testing
type mockCmdRunner struct {
	mock.Mock
}

func (m *mockCmdRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	arguments := m.Called(ctx, name, args)
	return arguments.Get(0).([]byte), arguments.Error(1)
}

func (m *mockCmdRunner) Start(ctx context.Context, name string, args ...string) (common.Process, error) {
	arguments := m.Called(ctx, name, args)
	if arguments.Get(0) == nil {
		return nil, arguments.Error(1)
	}
	return arguments.Get(0).(common.Process), arguments.Error(1)
}

// mockChannelRepository is a mock implementation of ChannelRepository for testing
type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) Create(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) GetByURL(ctx context.Context, url string) (*model.Channel, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*model.Channel), args.Error(1)
}

func (m *mockChannelRepository) Update(ctx context.Context, channel *model.Channel) error {
	args := m.Called(ctx, channel)
	return args.Error(0)
}

func (m *mockChannelRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockChannelRepository) List(ctx context.Context, limit, offset int) ([]*model.Channel, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Channel), args.Error(1)
}

// mockVideoRepository is a mock implementation of VideoRepository for testing
type mockVideoRepository struct {
	mock.Mock
}

func (m *mockVideoRepository) Create(ctx context.Context, video *model.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *mockVideoRepository) CreateBatch(ctx context.Context, videos []*model.Video) error {
	args := m.Called(ctx, videos)
	return args.Error(0)
}

func (m *mockVideoRepository) UpsertBatch(ctx context.Context, videos []*model.Video) error {
	args := m.Called(ctx, videos)
	return args.Error(0)
}

func (m *mockVideoRepository) GetByID(ctx context.Context, id string) (*model.Video, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Video), args.Error(1)
}

func (m *mockVideoRepository) GetByChannelID(ctx context.Context, channelID string, limit, offset int) ([]*model.Video, error) {
	args := m.Called(ctx, channelID, limit, offset)
	return args.Get(0).([]*model.Video), args.Error(1)
}

func (m *mockVideoRepository) Update(ctx context.Context, video *model.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *mockVideoRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockVideoRepository) List(ctx context.Context, limit, offset int) ([]*model.Video, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*model.Video), args.Error(1)
}
