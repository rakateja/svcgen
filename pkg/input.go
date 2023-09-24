package pkg

type EntityInput struct {
	ID       string             `yaml:"id"`
	IsParent bool               `yaml:"isParent"`
	Fields   []FieldInput       `yaml:"fields"`
	Childs   []EntityChildInput `yaml:"childs"`
}

type EntityChildInput struct {
	Type string `yaml:"type"`
}

type FieldInput struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`
}
