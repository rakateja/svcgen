package pkg

type Entity struct {
	StructName    string
	TableName     string
	RootTableName string
	Timestamps    bool
	Fields        []EntityField
	Childs        []Entity
}

type EntityField struct {
	ID          string
	Type        string
	JsonTag     string
	SqlColumTag string
}

func FromInputList(input []EntityInput) []Entity {
	return []Entity{}
}
