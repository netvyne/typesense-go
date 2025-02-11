package typesense

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/netvyne/typesense-go/typesense/api"
	"github.com/netvyne/typesense-go/typesense/mocks"
	"github.com/stretchr/testify/assert"
)

type eqReaderMatcher struct {
	readerBytes []byte
}

func eqReader(r io.Reader) gomock.Matcher {
	allBytes, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return &eqReaderMatcher{readerBytes: allBytes}
}

func (m *eqReaderMatcher) Matches(x interface{}) bool {
	if _, ok := x.(io.Reader); !ok {
		return false
	}
	r := x.(io.Reader)
	allBytes, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return reflect.DeepEqual(allBytes, m.readerBytes)
}

func (m *eqReaderMatcher) String() string {
	return string(m.readerBytes)
}

func TestDocumentsImportWithOneDocument(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	expectedBody := strings.NewReader(`{"id":"123","companyName":"Stark Industries","numEmployees":5215,"country":"USA"}` + "\n")
	expectedResultString := `{"success": true}`
	expectedResult := []*api.ImportDocumentResponse{
		{Success: true},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", eqReader(expectedBody)).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(expectedResultString)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	documents := []interface{}{
		createNewDocument(),
	}
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	result, err := client.Collection("companies").Documents().Import(documents, params)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestDocumentsImportWithEmptyListReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	documents := []interface{}{}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportWithOneDocumentAndInvalidResultJsonReturnsError(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	expectedBody := strings.NewReader(`{"id":"123","companyName":"Stark Industries","numEmployees":5215,"country":"USA"}` + "\n")
	expectedResultString := `{"success": invalid_json,}`

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", eqReader(expectedBody)).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(expectedResultString)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	documents := []interface{}{
		createNewDocument(),
	}
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportWithInvalidInputDataReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	documents := []interface{}{
		func() {},
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportOnApiClientErrorReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", gomock.Any(), "application/octet-stream", gomock.Any()).
		Return(nil, errors.New("failed request")).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	documents := []interface{}{
		createNewDocument(),
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportOnHttpStatusErrorCodeReturnsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", gomock.Any(), "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(strings.NewReader("Internal server error")),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	documents := []interface{}{
		createNewDocument(),
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportWithTwoDocuments(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	expectedBody := strings.NewReader(`{"id":"123","companyName":"Stark Industries","numEmployees":5215,"country":"USA"}` +
		"\n" + `{"id":"125","companyName":"Stark Industries","numEmployees":5215,"country":"USA"}` + "\n")
	expectedResultString := `{"success": true}` + "\n" + `{"success": false, "error": "Bad JSON.", "document": "[bad doc"}`
	expectedResult := []*api.ImportDocumentResponse{
		{Success: true},
		{Success: false, Error: "Bad JSON.", Document: "[bad doc"},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", eqReader(expectedBody)).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(expectedResultString)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	documents := []interface{}{
		createNewDocument("123"),
		createNewDocument("125"),
	}
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	result, err := client.Collection("companies").Documents().Import(documents, params)

	assert.Nil(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestDocumentsImportWithActionOnly(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: defaultImportBatchSize,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"success": true}`)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	documents := []interface{}{
		createNewDocument(),
	}
	params := &api.ImportDocumentsParams{
		Action: "create",
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.Nil(t, err)
}

func TestDocumentsImportWithBatchSizeOnly(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 10,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"success": true}`)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	documents := []interface{}{
		createNewDocument(),
	}
	params := &api.ImportDocumentsParams{
		BatchSize: 10,
	}
	_, err := client.Collection("companies").Documents().Import(documents, params)
	assert.Nil(t, err)
}

func TestDocumentsImportJsonl(t *testing.T) {
	expectedBytes := []byte(`{"success": true}`)
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	expectedBody := createDocumentStream()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", eqReader(expectedBody)).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer(expectedBytes)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	importBody := createDocumentStream()
	result, err := client.Collection("companies").Documents().ImportJsonl(importBody, params)
	assert.Nil(t, err)

	resultBytes, err := ioutil.ReadAll(result)
	assert.Nil(t, err)
	assert.Equal(t, string(expectedBytes), string(resultBytes))
}

func TestDocumentsImportJsonlOnApiClientErrorReturnsError(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			gomock.Any(), gomock.Any(), "application/octet-stream", gomock.Any()).
		Return(nil, errors.New("failed request")).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	importBody := createDocumentStream()
	_, err := client.Collection("companies").Documents().ImportJsonl(importBody, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportJsonlOnHttpStatusErrorCodeReturnsError(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			gomock.Any(), gomock.Any(), "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(strings.NewReader("Internal server error")),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 40,
	}
	importBody := createDocumentStream()
	_, err := client.Collection("companies").Documents().ImportJsonl(importBody, params)
	assert.NotNil(t, err)
}

func TestDocumentsImportJsonlWithActionOnly(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: defaultImportBatchSize,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"success": true}`)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		Action: "create",
	}
	importBody := createDocumentStream()
	_, err := client.Collection("companies").Documents().ImportJsonl(importBody, params)
	assert.Nil(t, err)
}

func TestDocumentsImportJsonlWithBatchSizeOnly(t *testing.T) {
	expectedParams := &api.ImportDocumentsParams{
		Action:    "create",
		BatchSize: 10,
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAPIClient := mocks.NewMockAPIClientInterface(ctrl)

	mockAPIClient.EXPECT().
		ImportDocumentsWithBody(gomock.Not(gomock.Nil()),
			"companies", expectedParams, "application/octet-stream", gomock.Any()).
		Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(strings.NewReader(`{"success": true}`)),
		}, nil).
		Times(1)

	client := NewClient(WithAPIClient(mockAPIClient))
	params := &api.ImportDocumentsParams{
		BatchSize: 10,
	}
	importBody := createDocumentStream()
	_, err := client.Collection("companies").Documents().ImportJsonl(importBody, params)
	assert.Nil(t, err)
}
