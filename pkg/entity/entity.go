package entity

type Entity struct {
	ID       string        `yaml:"id"`
	IsParent bool          `yaml:"isParent"`
	Fields   []Field       `yaml:"fields"`
	Filter   []string      `yaml:"filter"`
	Childs   []EntityChild `yaml:"childs"`
}

type EntityChild struct {
	Type string `yaml:"type"`
}

type Field struct {
	ID   string `yaml:"id"`
	Type string `yaml:"type"`
}
