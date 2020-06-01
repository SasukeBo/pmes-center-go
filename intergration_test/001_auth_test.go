package test

import (
	"net/http"
	"strings"
	"testing"
)

func TestAuthenticate(t *testing.T) {
	tester := newTester(t)
	// login
	ret := tester.POST("/api/login", object{
		"account":  testAdminAccount,
		"password": testAdminPasswd,
	}).Expect()
	ret.Status(http.StatusOK)
	setCookie := strings.Split(ret.Header("Set-Cookie").Raw(), ";")
	accessTokenCookie = setCookie[0]
	ret.JSON().Object().Value("status").Equal("ok")

	// current user
	ret1 := tester.API1(currentUserGQL, object{}).GQLObject().Path("$.data.currentUser").Object()
	ret1.Value("account").Equal(testAdminAccount)

	// logout
	ret2 := tester.GET("/api/logout", object{}).Expect()
	ret2.Status(http.StatusOK)
	tester.API1(currentUserGQL, object{}).GQLObject().Path("$.errors").Array().First().Object().Path("$.extensions.code").Equal(http.StatusForbidden)
}
