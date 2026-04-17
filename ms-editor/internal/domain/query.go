package domain

type Pagination struct {
	Offset int
	Limit  int
}

type ListResult[T any] struct {
	Items  []T
	Total  int
	Offset int
	Limit  int
}

type ObjectListQuery struct {
	Pagination
	Search string
}

type DeviceListQuery struct {
	Pagination
	ObjectID *string
	TypeID   *int
	Search   string
}

type TagListQuery struct {
	Pagination
	DeviceID   *string
	DataTypeID *int
	Search     string
}

type DiagramListQuery struct {
	Pagination
	ObjectID *string
	Search   string
}

type FigureListQuery struct {
	Pagination
	DiagramID  string
	TypeFilter *string
}
