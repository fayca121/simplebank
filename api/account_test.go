package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	mockdb "github.com/fayca121/simplebank/db/mock"
	db "github.com/fayca121/simplebank/db/sqlc"
	"github.com/fayca121/simplebank/token"
	"github.com/fayca121/simplebank/util"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetAccountAPI(t *testing.T) {
	user := randomUser()
	account := randomAccount(user.Username)

	testcases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			accountId: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InvalidOwner",
			accountId: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(util.RandomOwner(), time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	//build stubs
	for i := range testcases {
		tc := testcases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)
			// start test server and send request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {

	user := randomUser()
	account := randomAccount(user.Username)

	testCases := []struct {
		name          string
		request       *createAccountRequest
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStub     func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			request: &createAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name: "InternalError",
			request: &createAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				arg := db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				}

				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidBodyRequest",
			request: &createAccountRequest{
				Owner:    account.Owner,
				Currency: "DA",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(user.Username, time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidOwner",
			request: &createAccountRequest{
				Owner:    account.Owner,
				Currency: account.Currency,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				generatedToken, payload, err := tokenMaker.CreateToken(util.RandomOwner(), time.Minute)
				require.NoError(t, err)
				require.NotEmpty(t, payload)
				authorizationHeader := fmt.Sprintf("%s %s", authorizationTypeBearer, generatedToken)
				request.Header.Add(authorizationHeaderKey, authorizationHeader)
			},
			buildStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStub(store)
			// start test server and send request
			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()
			body, err := json.Marshal(tc.request)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, "/accounts/", bytes.NewReader(body))
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func randomUser() db.User {
	return db.User{
		Username:          util.RandomOwner(),
		FullName:          util.RandomOwner(),
		Email:             util.RandomEmail(),
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, gotAccount, account)
}
