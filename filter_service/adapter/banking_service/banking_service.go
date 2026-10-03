package banking_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "banking-service"

const DefaultBaseURL = "http://banking-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Persons(ctx context.Context) ([]*model.BankPerson, error) {
	path := "/persons"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankPerson
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Person(ctx context.Context, id int) (*model.BankPerson, error) {
	path := "/persons/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankPerson
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Branches(ctx context.Context) ([]*model.BankBranch, error) {
	path := "/branches"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankBranch
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Branch(ctx context.Context, id int) (*model.BankBranch, error) {
	path := "/branches/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankBranch
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Customers(ctx context.Context) ([]*model.BankCustomer, error) {
	path := "/customers"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankCustomer
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Customer(ctx context.Context, id int) (*model.BankCustomer, error) {
	path := "/customers/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankCustomer
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Employees(ctx context.Context) ([]*model.BankEmployee, error) {
	path := "/employees"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankEmployee
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Employee(ctx context.Context, id int) (*model.BankEmployee, error) {
	path := "/employees/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankEmployee
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Accounts(ctx context.Context) ([]*model.BankAccount, error) {
	path := "/accounts"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankAccount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Account(ctx context.Context, id int) (*model.BankAccount, error) {
	path := "/accounts/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankAccount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Loans(ctx context.Context) ([]*model.BankLoan, error) {
	path := "/loans"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankLoan
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Loan(ctx context.Context, id int) (*model.BankLoan, error) {
	path := "/loans/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankLoan
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LoanPayments(ctx context.Context) ([]*model.BankLoanPayment, error) {
	path := "/loan-payments"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankLoanPayment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LoanPayment(ctx context.Context, id int) (*model.BankLoanPayment, error) {
	path := "/loan-payments/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankLoanPayment
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transactions(ctx context.Context) ([]*model.BankTransaction, error) {
	path := "/transactions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.BankTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transaction(ctx context.Context, id int) (*model.BankTransaction, error) {
	path := "/transactions/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.BankTransaction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
