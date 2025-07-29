package image

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockImageService struct {
	mock.Mock
}

func (m *MockImageService) SaveImage(ctx context.Context, image []byte, filename string) (string, error) {
	args := m.Called(ctx, image, filename)
	return args.String(0), args.Error(1)
}

func (m *MockImageService) GetImageURL(ctx context.Context, filename string) (string, error) {
	args := m.Called(ctx, filename)
	return args.String(0), args.Error(1)
}

func (m *MockImageService) DeleteImage(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}
