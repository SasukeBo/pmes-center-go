package errormap

func init() {
	registerArg("decode_template", langMap{
		ZH_CN: "解析模板",
		EN:    "the decode template",
	})
	registerArg("material", langMap{
		ZH_CN: "料号",
		EN:    "the material",
	})
	registerArg("material_devices", langMap{
		ZH_CN: "该料号的设备",
		EN:    "devices of the material",
	})
	registerArg("material_import_records", langMap{
		ZH_CN: "该料号的导入记录",
		EN:    "import records of the material",
	})
	registerArg("material_decode_templates", langMap{
		ZH_CN: "该料号的解析模板",
		EN:    "decode templates of the material",
	})
	registerArg("material_default_decode_template", langMap{
		ZH_CN: "该料号的默认解析模板",
		EN:    "the default decode template of the material",
	})
}
