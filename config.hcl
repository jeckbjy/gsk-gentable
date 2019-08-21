# 配置模板,仅供参考
type = "gdoc"
options {
	gopkg = "tbl"
}
tasks "server" {
	source = "xxx"
	outputs = [
	    {data="csv", code=["go"]},
	    {data="json", code=["go"]}
	]
}
