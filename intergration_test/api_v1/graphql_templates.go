package api_v1

var currentUserGQL = ` query { currentUser { id account isAdmin } } `
var materialsGQL = `
query($page: Int!, $limit: Int!, $search: String) {
	response: materials(page: $page, limit: $limit, search: $search) {
		total
		materials {
			id
			name
			ok
			ng
			customerCode
			projectRemark
		}
	}
}
`
var materialGQL = `
query($id: Int!) {
	response: material(id: $id) {
		id
		name
		ok
		ng
		customerCode
		projectRemark
	}
}
`

