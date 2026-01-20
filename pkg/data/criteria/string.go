package criteria

type stringCriteria string

func (s stringCriteria) GetQuery() string {
	return string(s)
}

func (s stringCriteria) GetValues() []any {
	return nil
}
