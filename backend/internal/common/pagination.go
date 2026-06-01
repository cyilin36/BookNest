package common

import "strconv"

func ParsePagination(pageStr, sizeStr string) (int, int) {
	page := 1
	pageSize := 20
	if n, err := strconv.Atoi(pageStr); err == nil && n > 0 {
		page = n
	}
	if n, err := strconv.Atoi(sizeStr); err == nil && n > 0 {
		pageSize = n
		if pageSize > 100 {
			pageSize = 100
		}
	}
	return page, pageSize
}

func NormalizeSortOrder(sortBy, order string, allowed map[string]struct{}, defaultSort string) (string, string, bool) {
	if sortBy == "" {
		sortBy = defaultSort
	} else if _, ok := allowed[sortBy]; !ok {
		return "", "", false
	}
	switch order {
	case "asc", "desc":
	case "":
		order = "desc"
	default:
		return "", "", false
	}
	return sortBy, order, true
}
