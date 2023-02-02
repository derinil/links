package account_test

import (
	"context"
	"errors"
	"testing"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/generic"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type (
	MockReader struct{ mock.Mock }
	MockWriter struct{ mock.Mock }
)

func (r *MockReader) Get(ctx context.Context, cmd *account.GetCmd) (*account.Account, error) {
	args := r.Called(ctx, cmd)
	return args.Get(0).(*account.Account), args.Error(1)
}

func (w *MockWriter) SaveAccount(ctx context.Context, a *account.Account) error {
	args := w.Called(ctx, a)
	return args.Error(0)
}

func TestGet(t *testing.T) {
	var (
		defaultAccount = &account.Account{
			DBStruct: generic.DBStruct{
				ID: uuid.New(),
			}}
		funnyErr = errors.New("haha")
	)

	testCases := []struct {
		name       string
		cmd        *account.GetCmd
		a          *account.Account
		readerErr  error
		handlerErr error
		id         uuid.UUID
	}{
		{
			name: "get by id",
			cmd: &account.GetCmd{
				ID: defaultAccount.ID,
			},
			a:  defaultAccount,
			id: defaultAccount.ID,
		},
		{
			name: "get by handle",
			cmd: &account.GetCmd{
				Handle: "asd",
			},
			a:  defaultAccount,
			id: defaultAccount.ID,
		},
		{
			name: "get shallow by handle",
			cmd: &account.GetCmd{
				Handle:  "asd",
				Shallow: true,
			},
			a:  defaultAccount,
			id: defaultAccount.ID,
		},
		{
			name: "reader error out",
			cmd: &account.GetCmd{
				Handle: "asd",
			},
			readerErr:  funnyErr,
			handlerErr: funnyErr,
			id:         uuid.New(),
		},
		{
			name: "account not found",
			cmd: &account.GetCmd{
				Handle: "asd",
			},
			handlerErr: account.ErrAccountNotFound,
			id:         uuid.New(),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				reader         = new(MockReader)
				writer         = new(MockWriter)
				accountHandler = account.NewHandler(reader, writer)
			)

			reader.On("Get", ctx, c.cmd).Return(c.a, c.readerErr).Once()

			a, err := accountHandler.Get(ctx, c.cmd)
			require.ErrorIs(t, err, c.handlerErr)

			reader.AssertExpectations(t)

			if c.handlerErr != nil {
				return
			}

			require.Equal(t, c.id, a.ID)
		})
	}
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		name       string
		cmd        *account.CreateCmd
		err        error
		errStr     string
		a          *account.Account
		exists     *account.Account
		skipReader bool
		skipWriter bool
	}{
		{
			name: "create normal account",
			cmd: &account.CreateCmd{
				Name:     "Big Account!!!!",
				Handle:   "enemy",
				Password: "hashedpassword:)",
			},
			a: account.New("Big Account!!!!", "enemy", "hashedpassword:)"),
		},
		{
			name: "handle taken",
			cmd: &account.CreateCmd{
				Name:     "Big Account!!!!",
				Handle:   "wretch",
				Password: "hashedpassword:)",
			},
			exists:     account.New("irrelevant", "wretch", "irrelevant"),
			a:          account.New("Big Account!!!!", "wretch", "hashedpassword:)"),
			err:        account.ErrHandleTaken,
			skipWriter: true,
		},
		{
			name: "invalid handle",
			cmd: &account.CreateCmd{
				Name:     "Big Account!!!!",
				Handle:   "wretch...",
				Password: "hashedpassword:)",
			},
			errStr:     "Account.Handle",
			skipReader: true,
			skipWriter: true,
		},
		{
			name: "invalid name",
			cmd: &account.CreateCmd{
				Name:     "toomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacterstoomanycharacters",
				Handle:   "wretch...",
				Password: "hashedpassword:)",
			},
			errStr:     "Account.Name",
			skipReader: true,
			skipWriter: true,
		},
		{
			name: "sanitized name and handle",
			cmd: &account.CreateCmd{
				Name:     " wrapped by space ",
				Handle:   " same ",
				Password: "hashedpassword:)",
			},
			a: account.New("wrapped by space", "same", "hashedpassword:)"),
		},
		{
			name: "id conflict",
			cmd: &account.CreateCmd{
				Name:     " wrapped by space ",
				Handle:   " same ",
				Password: "hashedpassword:)",
			},
			err: errors.New("id_conflict"),
			a:   account.New("wrapped by space", "same", "hashedpassword:)"),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				reader         = new(MockReader)
				writer         = new(MockWriter)
				accountHandler = account.NewHandler(reader, writer)
			)

			if !c.skipReader {
				reader.On("Get", ctx, mock.MatchedBy(func(cmd *account.GetCmd) bool {
					return cmd.Handle == c.a.Handle
				})).Return(c.exists, nil).Once()
			}

			if !c.skipWriter {
				writer.On("SaveAccount", ctx, mock.MatchedBy(func(a *account.Account) bool {
					return a.Name == c.a.Name &&
						a.Handle == c.a.Handle &&
						a.Password == c.a.Password
				})).Return(c.err).Once()
			}

			a, err := accountHandler.Create(ctx, c.cmd)

			reader.AssertExpectations(t)
			writer.AssertExpectations(t)

			if c.err != nil || c.errStr != "" || err != nil {
				if c.err != nil {
					require.ErrorIs(t, err, c.err)
				} else {
					require.Contains(t, err.Error(), c.errStr)
				}
				return
			}

			require.Equal(t, c.a.Name, a.Name)
			require.Equal(t, c.a.Handle, a.Handle)
			require.Equal(t, c.a.Password, a.Password)
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		defaultAccount     = account.New("name", "handle", "password")
		defaultAccountWith = func(handle, css string, links []account.Link) *account.Account {
			a := *defaultAccount

			a.Handle = handle
			a.CSS = css
			a.Links = links

			return &a
		}
		copy = func(a *account.Account) *account.Account {
			b := *a
			return &b
		}
	)

	testCases := []struct {
		name       string
		cmd        *account.UpdateCmd
		expected   *account.Account
		exists     *account.Account
		err        error
		errStr     string
		skipReader bool
		skipWriter bool
	}{
		{
			name: "account not found",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
			},
			skipReader: true,
			skipWriter: true,
			err:        account.ErrAccountNotFound,
		},
		{
			name: "empty update",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
			},
			expected: copy(defaultAccount),
			exists:   copy(defaultAccount),
		},
		{
			name: "valid update",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
				Handle:    "newhandle",
				CSS:       "body { color: white; }",
				Links: []account.LinkScaffold{
					{
						Title: "Link",
						Link:  "https://example.com",
					},
				},
			},
			expected: defaultAccountWith("newhandle", "body { color: white; }", []account.Link{
				*account.NewLink(defaultAccount.ID, "Link", "https://example.com", 0),
			}),
			exists: copy(defaultAccount),
		},
		{
			name: "invalid link",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
				Handle:    "newhandle",
				CSS:       "body { color: white; }",
				Links: []account.LinkScaffold{
					{
						Title: "Link",
						Link:  "examplecom",
					},
				},
			},
			expected: defaultAccountWith("newhandle", "body { color: white; }", []account.Link{
				*account.NewLink(defaultAccount.ID, "Link", "https://example.com", 0),
			}),
			errStr:     "Link.Link",
			skipWriter: true,
			skipReader: true,
			exists:     copy(defaultAccount),
		},
		{
			name: "harmful css",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
				Handle:    "newhandle",
				CSS:       "body { <script>alert(owned);</script>color: white; }",
				Links: []account.LinkScaffold{
					{
						Title: "Link",
						Link:  "examplecom",
					},
				},
			},
			expected: defaultAccountWith("newhandle", "body { color: white; }", []account.Link{
				*account.NewLink(defaultAccount.ID, "Link", "https://example.com", 0),
			}),
			errStr:     "Account.CSS",
			skipWriter: true,
			skipReader: true,
			exists:     copy(defaultAccount),
		},
		{
			name: "valid update with links and no new handle",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
				CSS:       "body { color: white; }",
				Links: []account.LinkScaffold{
					{
						Title: "Link",
						Link:  "https://example.com",
					},
					{
						Title: "Link 2",
						Link:  "https://google.com",
					},
				},
			},
			expected: defaultAccountWith("handle", "body { color: white; }", []account.Link{
				*account.NewLink(defaultAccount.ID, "Link", "https://example.com", 0),
				*account.NewLink(defaultAccount.ID, "Link 2", "https://google.com", 1),
			}),
			exists: copy(defaultAccount),
		},
		{
			name: "update existing link",
			cmd: &account.UpdateCmd{
				AccountID: defaultAccount.ID,
				Links: []account.LinkScaffold{
					{
						Title: "New Title for Link",
						Link:  "https://example.com",
					},
				},
			},
			expected: defaultAccountWith("handle", "", []account.Link{
				*account.NewLink(defaultAccount.ID, "New Title for Link", "https://example.com", 0),
			}),
			exists: defaultAccountWith("handle", "", []account.Link{
				*account.NewLink(defaultAccount.ID, "Link", "https://example.com", 0),
			}),
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			var (
				ctx            = context.Background()
				reader         = new(MockReader)
				writer         = new(MockWriter)
				accountHandler = account.NewHandler(reader, writer)
			)

			getFirst := reader.On("Get", ctx, mock.MatchedBy(func(cmd *account.GetCmd) bool {
				return cmd.ID == c.cmd.AccountID
			})).Return(c.exists, nil).Once()

			if !c.skipReader {
				reader.On("Get", ctx, mock.MatchedBy(func(cmd *account.GetCmd) bool {
					return cmd.Handle == c.expected.Handle
				})).Return(c.exists, nil).Once().NotBefore(getFirst)
			}

			if !c.skipWriter {
				writer.On("SaveAccount", ctx, mock.MatchedBy(func(a *account.Account) bool {
					return a.Name == c.expected.Name &&
						a.Handle == c.expected.Handle &&
						a.Password == c.expected.Password
				})).Return(c.err).Once()
			}

			a, err := accountHandler.Update(ctx, c.cmd)

			cancel := false

			if c.err != nil || c.errStr != "" || err != nil {
				if c.err != nil {
					require.ErrorIs(t, err, c.err)
				} else if c.errStr != "" {
					require.Contains(t, err.Error(), c.errStr)
				} else {
					require.Nil(t, err)
				}
				cancel = true
			}

			reader.AssertExpectations(t)
			writer.AssertExpectations(t)

			if cancel {
				return
			}

			require.Equal(t, c.expected.Name, a.Name)
			require.Equal(t, c.expected.Handle, a.Handle)
			require.Equal(t, c.expected.Password, a.Password)
			require.Equal(t, len(a.Links), len(c.expected.Links))
			for i := range c.expected.Links {
				el := &c.expected.Links[i]
				require.Equal(t, el.Title, a.Links[i].Title)
				require.Equal(t, el.Link, a.Links[i].Link)
			}
		})
	}
}
