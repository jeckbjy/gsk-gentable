package conf

import "testing"

func TestConfig_Parse(t *testing.T) {
	data := `
auth = "oath.json"
type = "gdoc"
tasks "server" {
	source = "1MVNpnskNeN7"
	outputs = [
		{sheets=["aaa", "bbb"],data_formats=["csv"], code_formats=["go"]}
	]
}
tasks "client" {
	source = "1MVNpnskNeN7"
	outputs = [
		{sheets=["ccc", "ddd"],data_formats=["csv"], code_formats=["csharp"]}
	]
}
`
	cfg := Config{}
	if err := cfg.Unmarshal(data); err != nil {
		t.Error(err)
	} else {
		t.Log(cfg)
	}
}
