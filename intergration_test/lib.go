package test

import (
	"encoding/json"
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/ftpviewer/orm"
	"github.com/SasukeBo/ftpviewer/router"
	"github.com/SasukeBo/ftpviewer/util"
	"github.com/gavv/httpexpect"
	"net/http"
	"net/url"
	"strings"
)

var host string
var accessTokenCookie string

type object map[string]interface{}

type tester struct {
	E       *httpexpect.Expect
	Headers map[string]string
}

type request struct {
	*httpexpect.Request
}

func (r *request) GQLObject() *httpexpect.Object {
	return r.Expect().Status(http.StatusOK).JSON().Object()
}

// set tester header
func (t *tester) SetHeader(key string, value string) {
	t.Headers[key] = value
}

// API post a api request
func (t *tester) API(query string, variables interface{}) *request {
	payload := map[string]interface{}{
		"operationName": "",
		"query":         query,
		"variables":     variables,
	}

	rr := t.E.POST("/api").WithHeaders(t.Headers).WithHeader("Cookie", accessTokenCookie).WithJSON(payload)
	return &request{rr}
}

// GET send a get request to url with query variables
func (t *tester) GET(url string, variables ...interface{}) *httpexpect.Request {
	return t.E.GET(url, variables...).WithHeaders(t.Headers)
}

// new a tester
func newTester(t httpexpect.LoggerReporter) *tester {
	tst := &tester{}
	tst.E = httpexpect.New(t, host)
	tst.Headers = make(map[string]string)
	return tst
}

// login 测试环境下登录系统
func login(account, password string) {
	client := &http.Client{}
	data := url.Values{}
	variables := object{
		"input": object{
			"account":  account,
			"password": password,
		},
	}
	content, _ := json.Marshal(variables)
	data.Set("operationName", "")
	data.Set("query", loginGQL)
	data.Set("variables", string(content))

	res, err := client.PostForm(fmt.Sprintf("%s%s", host, "/api"), data)
	if err != nil {
		panic(err)
	}
	setCookies := strings.Split(res.Header.Get("Set-Cookie"), ";")
	accessTokenCookie = setCookies[0]
}

func init() {
	tearDown()
	setup()
	host = fmt.Sprintf("http://localhost:%v", configer.GetEnv("port"))
	go router.Start()
	login(data.User.Username, testUserPasswd)
}

func tearDown() {
	orm.DB.Exec("DELETE FROM devices where 1 = 1")
	orm.DB.Exec("DELETE FROM files where 1 = 1")
	orm.DB.Exec("DELETE FROM materials where 1 = 1")
	orm.DB.Exec("DELETE FROM point_values where 1 = 1")
	orm.DB.Exec("DELETE FROM points where 1 = 1")
	orm.DB.Exec("DELETE FROM products where 1 = 1")
	orm.DB.Exec("DELETE FROM sizes where 1 = 1")
	orm.DB.Exec("DELETE FROM system_configs where 1 = 1")
	orm.DB.Exec("DELETE FROM users where 1 = 1")
}

// generate fake data

var data struct {
	User  *orm.User
	Admin *orm.User
}

const (
	testUserAccount  = "test_user"
	testUserPasswd   = "test_passwd"
	testAdminAccount = "test_admin"
	testAdminPasswd  = "test_admin_passwd"
)

func setup() {
	data.User = &orm.User{
		Admin:    false,
		Username: testUserAccount,
		Password: util.Encrypt(testUserPasswd),
	}
	orm.Create(data.User)
	data.Admin = &orm.User{
		Admin:    true,
		Username: testAdminAccount,
		Password: util.Encrypt(testAdminPasswd),
	}
	orm.Create(data.Admin)
}
