package admin

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
var productScrollFetchGQL = `
query($input: ProductSearch!, $limit: Int!, $offset: Int!) {
	response: productScrollFetch(searchInput: $input, limit: $limit, offset: $offset) {
		data {
			id
			materialID
			deviceID
			qualified
			createdAt
			attribute
			pointValues
		}
		total
	}
}
`
var pointImportGQL = ` mutation($file: Upload!, $materialID: Int!) { response: importPoints(file: $file, materialID: $materialID) { id name upperLimit nominal lowerLimit } } `
