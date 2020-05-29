package test

var currentUserGQL = ` query { currentUser { id account admin } } `
var loginGQL = ` mutation($input: LoginInput!) { login(loginInput: $input) { id account admin } }`
