package test

var currentUserGQL = ` query { currentUser { id account isAdmin } } `
var loginGQL = ` mutation($input: LoginInput!) { login(loginInput: $input) { id account isAdmin } }`
