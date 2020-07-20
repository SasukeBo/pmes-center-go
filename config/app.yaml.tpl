# app.yaml
env: dev

# 用于存储上传文件的根目录
file_cache_path: your_download_file_cache_path

# 检测尺寸导入模板Token，用于在数据库中检索该文件对象
points_import_template_token: points_import_template_token

# 默认解析模板中产品属性列 Index:Token:Label:Type
default_product_attribute_index: "C:2dcode:2D条码号:String;D:line:线体号:String;E:fixture:冶具号:String;F:tool:模号:String"
# 默认检测项第一列序号
default_point_begin_index: H

# 缓存持续时间，用于配置缓存中单个数据的存活时间，单位秒
cache_expired_time:
