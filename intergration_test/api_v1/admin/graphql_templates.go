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
var pointImportParseGQL = ` mutation($file: Upload!) { response: parseImportPoints(file: $file) { id name upperLimit nominal lowerLimit } } `
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
      status
      errorMessage
      originErrorMessage
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
var deleteDecodeTemplateGQL = ` mutation($id: Int!) { response: deleteDecodeTemplate(id: $id) } `
var changeDefaultTemplateGQL = ` mutation($id: Int!, $isDefault: Boolean!) { response: changeDefaultTemplate(id: $id, isDefault: $isDefault) } `
var deleteMaterialGQL = ` mutation($id: Int!) { response: deleteMaterial(id: $id) } `
var updateMaterialGQL = `
mutation($input: MaterialUpdateInput!) {
  response: updateMaterial(input: $input) {
    id
    name
    customerCode
    projectRemark
    createdAt
    updatedAt
  }
}`
var savePointsGQL = `
mutation($materialID: Int!, $saveItems: [PointCreateInput]!, $deleteItems: [Int!]!) {
  response: savePoints(materialID: $materialID, saveItems: $saveItems, deleteItems: $deleteItems)
}`
var listMaterialPointGQL = `
query($materialID: Int!) {
  response: listMaterialPoints( materialID: $materialID ) {
  	id
  	name
  	upperLimit
  	nominal
  	lowerLimit
  }
}`
var saveDeviceGQL = `
mutation($input: DeviceInput!) {
  response: saveDevice(input: $input) {
    id
    uuid
    name
    remark
    ip
    material {
      id
    }
    deviceSupplier
    isRealtime
    address
  }
}`
