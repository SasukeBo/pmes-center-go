package admin

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
var listMaterialGQL = `
query($pattern: String, $page: Int!, $limit: Int!) {
	response: materials(pattern: $pattern, page: $page, limit: $limit) {
		total
		materials {
			id
			name
			createdAt
			updatedAt
			customerCode
			projectRemark
		}
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
var listImportRecordsGQL = `
query($materialID: Int!, $deviceID: Int, $page: Int!, $limit: Int!) {
  response: importRecords(
    materialID: $materialID
    deviceID: $deviceID
    page: $page
    limit: $limit
  ) {
    total
    importRecords {
      id
      fileName
      material { id }
      device { id }
      rowCount
      rowFinishedCount
      finished
      error
      fileSize
      user { id }
      importType
      decodeTemplate { id }
    }
  }
}
`
var saveDecodeTemplateGQL = `
mutation($input: DecodeTemplateInput!) {
  response: saveDecodeTemplate(input: $input) {
    id
    name
    material { id }
    user { id } 
    description
    dataRowIndex
    createdAtColumnIndex
    productColumns {
		name
		index
		type
	}
    pointColumns
    default
    createdAt
    updatedAt
  }
}
`
var listDecodeTemplateGQL = `
query($materialID: Int!) {
  response: listDecodeTemplate(materialID: $materialID) {
    id
    name
    material {
      id
    }
    user {
      id
    }
    description
    dataRowIndex
    createdAtColumnIndex
    productColumns {
      name
      index
      type
    }
    pointColumns
    default
    createdAt
    updatedAt
  }
}
`
var deleteDecodeTemplateGQL = `
mutation($id: Int!) {
  response: deleteDecodeTemplate(id: $id)
}
`
