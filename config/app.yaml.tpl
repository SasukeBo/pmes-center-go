# app.yaml
env: dev

# 用于存储上传文件的根目录
file_cache_path: your_download_file_cache_path

# 检测尺寸导入模板Token，用于在数据库中检索该文件对象
points_import_template_token: points_import_template_token

# 解析模板中产品属性列
product_column_headers: "NO.:Integer;日期:Datetime;2D条码号:String;线体号:String;冶具号:String;模号:String;班别:String"

# 缓存持续时间，用于配置缓存中单个数据的存活时间，单位秒
cache_expired_time:
