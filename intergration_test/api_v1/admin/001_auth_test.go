package admin

import (
	test "github.com/SasukeBo/ftpviewer/intergration_test"
	"net/http"
	"strings"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	tester := test.NewTester(t)
	// login
	ret := tester.POST("/api/login", test.Object{
		"account":  test.AdminAccount,
		"password": test.AdminPasswd,
	}).Expect()
	ret.Status(http.StatusOK)
	setCookie := strings.Split(ret.Header("Set-Cookie").Raw(), ";")
	test.AccessTokenCookie = setCookie[0]
	ret.JSON().Object().Value("status").Equal("ok")

	// current user
	ret1 := tester.API1Admin(currentUserGQL, test.Object{}).GQLObject().Path("$.data.currentUser").Object()
	ret1.Value("account").Equal(test.AdminAccount)

	// logout
	ret2 := tester.GET("/api/logout", test.Object{}).Expect()
	ret2.Status(http.StatusOK)
	tester.API1Admin(currentUserGQL, test.Object{}).GQLObject().Path("$.errors").Array().First().Object().Path("$.extensions.code").Equal(http.StatusForbidden)
}
