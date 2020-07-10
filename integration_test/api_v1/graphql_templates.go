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
var analyzeMaterialGQL = `
query($input: AnalyzeMaterialInput!) {
	response: analyzeMaterial(analyzeInput: $input) {
		xAxisData
		seriesData
	}
}
`
var materialYieldTopGQL = `
query($duration: [Time]!, $limit: Int!) {
	response: materialYieldTop(duration: $duration, limit: $limit) {
		xAxisData
		seriesData
	}
}
`
var sizeUnYieldTopGQL = `
query($groupInput: GroupAnalyzeInput!) {
	response: sizeUnYieldTop(groupInput: $groupInput) {
		xAxisData
		seriesData
		seriesAmountData
	}
}
`
