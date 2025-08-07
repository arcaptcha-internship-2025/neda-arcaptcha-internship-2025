package image

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockImage struct {
	mock.Mock
}

func (m *MockImage) SaveImage(ctx context.Context, image []byte, filename string) (string, error) {
	args := m.Called(ctx, image, filename)
	return args.String(0), args.Error(1)
}

func (m *MockImage) GetImageURL(ctx context.Context, filename string) (string, error) {
	args := m.Called(ctx, filename)
	return args.String(0), args.Error(1)
}

func (m *MockImage) DeleteImage(ctx context.Context, filename string) error {
	args := m.Called(ctx, filename)
	return args.Error(0)
}

func NewMockImage() *MockImage {
	return &MockImage{}
}

func (m *MockImage) ExpectSaveImage(ctx context.Context, image []byte, filename string, returnObjectKey string, returnError error) *mock.Call {
	return m.On("SaveImage", ctx, image, filename).Return(returnObjectKey, returnError)
}

func (m *MockImage) ExpectSaveImageOnce(ctx context.Context, image []byte, filename string, returnObjectKey string, returnError error) *mock.Call {
	return m.On("SaveImage", ctx, image, filename).Return(returnObjectKey, returnError).Once()
}

func (m *MockImage) ExpectSaveImageWithAnyData(ctx context.Context, filename string, returnObjectKey string, returnError error) *mock.Call {
	return m.On("SaveImage", ctx, mock.AnythingOfType("[]uint8"), filename).Return(returnObjectKey, returnError)
}

func (m *MockImage) ExpectSaveImageWithAnyArgs(returnObjectKey string, returnError error) *mock.Call {
	return m.On("SaveImage", mock.Anything, mock.Anything, mock.Anything).Return(returnObjectKey, returnError)
}

func (m *MockImage) ExpectGetImageURL(ctx context.Context, filename string, returnURL string, returnError error) *mock.Call {
	return m.On("GetImageURL", ctx, filename).Return(returnURL, returnError)
}

func (m *MockImage) ExpectGetImageURLOnce(ctx context.Context, filename string, returnURL string, returnError error) *mock.Call {
	return m.On("GetImageURL", ctx, filename).Return(returnURL, returnError).Once()
}

func (m *MockImage) ExpectGetImageURLWithAnyArgs(returnURL string, returnError error) *mock.Call {
	return m.On("GetImageURL", mock.Anything, mock.Anything).Return(returnURL, returnError)
}

func (m *MockImage) ExpectGetImageURLEmpty(ctx context.Context) *mock.Call {
	return m.On("GetImageURL", ctx, "").Return("", nil)
}

func (m *MockImage) ExpectDeleteImage(ctx context.Context, filename string, returnError error) *mock.Call {
	return m.On("DeleteImage", ctx, filename).Return(returnError)
}

func (m *MockImage) ExpectDeleteImageOnce(ctx context.Context, filename string, returnError error) *mock.Call {
	return m.On("DeleteImage", ctx, filename).Return(returnError).Once()
}

func (m *MockImage) ExpectDeleteImageWithAnyArgs(returnError error) *mock.Call {
	return m.On("DeleteImage", mock.Anything, mock.Anything).Return(returnError)
}

func (m *MockImage) ExpectDeleteImageEmpty(ctx context.Context) *mock.Call {
	return m.On("DeleteImage", ctx, "").Return(nil)
}

func (m *MockImage) ExpectSaveImageSizeLimit(ctx context.Context, image []byte, filename string) *mock.Call {
	return m.On("SaveImage", ctx, image, filename).Return("", mock.MatchedBy(func(err error) bool {
		return err != nil && err.Error() == "image size exceeds 10MB limit"
	}))
}

func (m *MockImage) ExpectSaveImageUnsupportedType(ctx context.Context, image []byte, filename string, ext string) *mock.Call {
	expectedError := mock.MatchedBy(func(err error) bool {
		return err != nil && err.Error() == "unsupported file type: "+ext
	})
	return m.On("SaveImage", ctx, image, filename).Return("", expectedError)
}

func (m *MockImage) ExpectSaveImageTimes(times int, ctx context.Context, image []byte, filename string, returnObjectKey string, returnError error) *mock.Call {
	return m.On("SaveImage", ctx, image, filename).Return(returnObjectKey, returnError).Times(times)
}

func (m *MockImage) ExpectGetImageURLTimes(times int, ctx context.Context, filename string, returnURL string, returnError error) *mock.Call {
	return m.On("GetImageURL", ctx, filename).Return(returnURL, returnError).Times(times)
}

func (m *MockImage) ExpectDeleteImageTimes(times int, ctx context.Context, filename string, returnError error) *mock.Call {
	return m.On("DeleteImage", ctx, filename).Return(returnError).Times(times)
}

func (m *MockImage) ExpectNoImageCalls(t mock.TestingT) {
	m.AssertNotCalled(t, "SaveImage")
	m.AssertNotCalled(t, "GetImageURL")
	m.AssertNotCalled(t, "DeleteImage")
}

func (m *MockImage) ExpectSaveImageCalledWith(t mock.TestingT, ctx context.Context, image []byte, filename string) {
	m.AssertCalled(t, "SaveImage", ctx, image, filename)
}

func (m *MockImage) ExpectGetImageURLCalledWith(t mock.TestingT, ctx context.Context, filename string) {
	m.AssertCalled(t, "GetImageURL", ctx, filename)
}

func (m *MockImage) ExpectDeleteImageCalledWith(t mock.TestingT, ctx context.Context, filename string) {
	m.AssertCalled(t, "DeleteImage", ctx, filename)
}
