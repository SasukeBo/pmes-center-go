package test

import (
	"fmt"
	"github.com/SasukeBo/configer"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/router"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/gavv/httpexpect"
	"net/http"
	"net/url"
	"strings"
)

var host string
var AccessTokenCookie string

type Object map[string]interface{}

type Tester struct {
	E       *httpexpect.Expect
	Headers map[string]string
}

type Request struct {
	*httpexpect.Request
}

func (r *Request) GQLObject() *httpexpect.Object {
	return r.Expect().Status(http.StatusOK).JSON().Object()
}

// set Tester header
func (t *Tester) SetHeader(key string, value string) {
	t.Headers[key] = value
}

// send a POST Request with form Data
func (t *Tester) POST(path string, variables interface{}, pathargs ...interface{}) *Request {
	rr := t.E.POST(path, pathargs...).WithHeaders(t.Headers).WithHeader("Cookie", AccessTokenCookie).WithForm(variables)
	return &Request{rr}
}

// send a GET Request with query
func (t *Tester) GET(path string, variables interface{}, pathargs ...interface{}) *Request {
	rr := t.E.GET(path, pathargs...).WithHeaders(t.Headers).WithHeader("Cookie", AccessTokenCookie).WithQueryObject(variables)
	return &Request{rr}
}

// API1Admin post a api/v1/admin Request
func (t *Tester) API1Admin(query string, variables interface{}) *Request {
	return t.api("/api/v1/admin", query, variables)
}

// API1Admin post a api/v1 Request
func (t *Tester) API1(query string, variables interface{}) *Request {
	return t.api("/api/v1", query, variables)
}

func (t *Tester) api(path, query string, variables interface{}) *Request {
	payload := map[string]interface{}{
		"operationName": "",
		"query":         query,
		"variables":     variables,
	}

	rr := t.E.POST(path).WithHeaders(t.Headers).WithHeader("Cookie", AccessTokenCookie).WithJSON(payload)
	return &Request{rr}
}

// send a POST Request with form Data
func (t *Tester) Upload(path string, pathargs ...interface{}) *Request {
	rr := t.E.POST(path, pathargs...).WithHeaders(t.Headers).WithHeader("Cookie", AccessTokenCookie)
	return &Request{rr}
}

// new a Tester
func NewTester(t httpexpect.LoggerReporter) *Tester {
	tst := &Tester{}
	tst.E = httpexpect.New(t, host)
	tst.Headers = make(map[string]string)
	return tst
}

// Login 测试环境下登录系统
func Login(account, password string, remember bool) {
	client := &http.Client{}
	data := url.Values{}
	data.Set("account", account)
	data.Set("password", password)
	data.Set("remember", fmt.Sprint(remember))

	res, err := client.PostForm(fmt.Sprintf("%s%s", host, "/auth/login"), data)
	if err != nil {
		panic(err)
	}
	setCookies := strings.Split(res.Header.Get("Set-Cookie"), ";")
	AccessTokenCookie = setCookies[0]
}

func init() {
	orm.DB.LogMode(false)
	tearDown()
	setup()
	host = fmt.Sprintf("http://localhost:%v", configer.GetEnv("port"))
	go router.Start()
	Login(Data.User.Account, UserPasswd, true)
	orm.DB.LogMode(true)
}

func tearDown() {
	var tables = []string{
		"bar_code_rules",
		"decode_templates",
		"devices",
		"import_records",
		"materials",
		"material_versions",
		"points",
		"products",
		"system_configs",
		"users",
		"user_logins",
	}

	for _, name := range tables {
		cleanTable(name)
	}
	orm.GenerateDefaultConfig()
}

func cleanTable(tbName string) {
	orm.DB.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1 = 1", tbName))
}

// generate fake Data

var Data struct {
	User     *orm.User
	Admin    *orm.User
	Material *orm.Material
	Device   *orm.Device
}

const (
	UserAccount  = "test_user"
	UserPasswd   = "test_passwd"
	AdminAccount = "test_admin"
	AdminPasswd  = "test_admin_passwd"
)

func setup() {
	Data.User = &orm.User{
		IsAdmin:  false,
		Account:  UserAccount,
		Password: util.Encrypt(UserPasswd),
	}
	orm.Create(Data.User)
	Data.Admin = &orm.User{
		IsAdmin:  true,
		Account:  AdminAccount,
		Password: util.Encrypt(AdminPasswd),
	}
	orm.Create(Data.Admin)
	Data.Material = &orm.Material{
		Name:          "mock_material",
		CustomerCode:  "mock_material_customer_code",
		ProjectRemark: "mock_material_project_remark",
	}
	orm.Create(Data.Material)
	Data.Device = &orm.Device{
		Name:       "mock_device",
		Remark:     "mock_device",
		MaterialID: Data.Material.ID,
	}
	orm.Create(Data.Device)
}
