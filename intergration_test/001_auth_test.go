package test

import (
	"testing"
)

func TestAuthenticate(t *testing.T) {
	tester := newTester(t)
	ret := tester.API(loginGQL, object{
		"input": object{
			"account":  testAdminAccount,
			"password": testAdminPasswd,
		},
	}).GQLObject().Path("$.data.login").Object()
	ret.Value("admin").Equal(true)
	ret1 := tester.API(currentUserGQL, object{}).GQLObject().Path("$.data.currentUser").Object()
	ret1.Value("account").Equal(testAdminAccount)
}
