package test

var currentUserGQL = ` query { currentUser { id account isAdmin } } `
var loginGQL = ` mutation($input: LoginInput!) { login(loginInput: $input) { id account isAdmin } }`
var createMaterialGQL = `
mutation($input: MaterialCreateInput!) {
	response: addMaterial(input: $input) {
		id
		name
		customerCode
		projectRemark
	}
}
`