package config

// Struct representing the yaml configuration file passed as a parameter to the program
type YamlConfig struct {
	vars []gitlabVar `yaml:"vars"`
}

type gitlabVar struct {
	key   string `yaml:"key"`
	value string `yaml:"value"`
}

/*
{
    "key": "NEW_VARIABLE",
    "value": "new value",
    "protected": false,
    "variable_type": "env_var",
    "masked": false,
    "environment_scope": "*"
}
*/
