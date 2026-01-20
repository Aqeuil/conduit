package criteria

import "fmt"

type notCriteria struct {
	Criteria
}

func (s notCriteria) GetQuery() string {
	return fmt.Sprintf(" NOT (%s)", s.Criteria.GetQuery())
}
