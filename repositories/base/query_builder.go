package base

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueryBuilder[T any] struct {
	filter       map[string]interface{}
	preload      []string
	omit         []string
	sort         map[string]string
	limit        int
	page         int
	searchValue  string
	searchFields []string
	alias        map[string]string
	// Meili integration
	meiliIndex   string
	meiliFilters []string
	isMeili      bool
}

func NewQueryBuilder[T any]() *QueryBuilder[T] {
	return &QueryBuilder[T]{
		filter:  make(map[string]interface{}),
		preload: []string{},
		omit:    []string{},
		sort:    make(map[string]string),
		limit:   1000,
		page:    1,
		alias:   make(map[string]string),
	}
}

func (qb *QueryBuilder[T]) SetFilter(filter map[string]interface{}) {
	qb.filter = filter
}

func (qb *QueryBuilder[T]) SetPreload(preload []string) {
	qb.preload = preload
}

func (qb *QueryBuilder[T]) SetOmit(omit []string) {
	qb.omit = omit
}

func (qb *QueryBuilder[T]) SetSort(sort map[string]string) {
	qb.sort = sort
}

func (qb *QueryBuilder[T]) SetLimit(limit int) {
	qb.limit = limit
}

func (qb *QueryBuilder[T]) SetPage(page int) {
	qb.page = page
}

func (qb *QueryBuilder[T]) SetSearch(value string, fields []string) {
	qb.searchValue = value
	qb.searchFields = fields
}

func (qb *QueryBuilder[T]) SetAlias(alias map[string]string) {
	qb.alias = alias
}

func (qb *QueryBuilder[T]) SetMeili(index string, filters []string) {
	qb.meiliIndex = index
	qb.meiliFilters = filters
}

func (qb *QueryBuilder[T]) GetOmit() []string {
	return qb.omit
}

func (qb *QueryBuilder[T]) ApplySearch(query *gorm.DB) *gorm.DB {
	if qb.searchValue == "" || len(qb.searchFields) == 0 {
		return query
	}

	likePattern := "%" + qb.searchValue + "%"
	// Using tsquery_pattern to create full-text search queries
	// For example, 'con heo' (pig) would become 'con & heo' or 'con | heo' depending on the logic you want.
	// For approximate and prefix searches, you might consider adding ":*" to the end of each word if you want to search for prefixes,
	// or use `phraseto_tsquery` for more precise phrasing.
	// Here we use `plainto_tsquery` for simplicity, it will handle words and the default AND operator.
	tsQueryPattern := qb.searchValue

	orConditions := []string{}
	values := []interface{}{}

	var model T
	stmt := &gorm.Statement{DB: query}
	if err := stmt.Parse(&model); err != nil {
		return query
	}
	tableName := stmt.Schema.Table
	joinedTables := make(map[string]bool)
	sortableFields := []string{}
	seenFields := make(map[string]bool)

	for _, field := range qb.searchFields {
		if strings.Count(field, ":") == 5 {
			// Format: relationTable:relationFromKey:relationToKey:baseTableKey:targetTableKey:targetTable.col1,col2
			// Example search users by classes.name,code with relation table user_classes: "user_classes:user_id:class_id:id:id:classes.name,code"
			parts := strings.Split(field, ":")
			if len(parts) != 6 {
				continue
			}
			relationTable := parts[0]
			relationFromKey := parts[1]
			relationToKey := parts[2]
			baseTableKey := parts[3]
			targetTableKey := parts[4]
			targetFieldStr := parts[5]

			targetFieldParts := strings.Split(targetFieldStr, ".")
			if len(targetFieldParts) != 2 {
				continue
			}
			targetTable := targetFieldParts[0]
			targetColumns := strings.Split(targetFieldParts[1], ",")

			if !joinedTables[relationTable] {
				query = query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.%s",
					relationTable, relationTable, relationFromKey, tableName, baseTableKey))
				joinedTables[relationTable] = true
			}

			if !joinedTables[targetTable] {
				query = query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.%s",
					targetTable, targetTable, targetTableKey, relationTable, relationToKey))
				joinedTables[targetTable] = true
			}

			for _, col := range targetColumns {
				col = strings.TrimSpace(col)
				fullCol := fmt.Sprintf("%s.%s", targetTable, col)

				if strings.HasSuffix(col, "id") {
					if idInt, err := strconv.Atoi(qb.searchValue); err == nil {
						orConditions = append(orConditions, fmt.Sprintf("%s = ?", fullCol))
						values = append(values, idInt)
						if !seenFields[fullCol] {
							sortableFields = append(sortableFields, fullCol)
							seenFields[fullCol] = true
						}
						continue
					}
				}

				// LIKE
				orConditions = append(orConditions, fmt.Sprintf("unaccent(CAST(%s AS TEXT)) ILIKE unaccent(?)", fullCol))
				values = append(values, likePattern)

				// Full-Text Search
				// Use 'simple' or 'vietnamese' depending on your text search configuration
				orConditions = append(orConditions, fmt.Sprintf("to_tsvector('simple', unaccent(CAST(%s AS TEXT))) @@ plainto_tsquery('simple', unaccent(?))", fullCol))
				values = append(values, tsQueryPattern)

				if !seenFields[fullCol] {
					sortableFields = append(sortableFields, fullCol)
					seenFields[fullCol] = true
				}
			}

		} else if strings.Count(field, ":") == 3 {
			// Format: targetTable:targetFK:basePK:col1,col2
			// Example search users from user_addresses.address by realated user_id: "user_addresses:user_id:id:user_addresses.address,ward_code",
			// Example search users from districts.name,full_name by realated district_code: "districts:district_code:code:districts.name,full_name"
			parts := strings.Split(field, ":")
			if len(parts) != 4 {
				continue
			}

			targetTable := parts[0]
			targetFK := parts[1]
			basePK := parts[2]
			targetColumnsStr := parts[3]

			joinCondition := fmt.Sprintf("%s.%s = %s.%s", tableName, basePK, targetTable, targetFK)
			// Need to check if targetFK is the primary key of the original table,
			// otherwise the join condition may fail if that field is not present in the schema
			if stmt.Schema.LookUpField(targetFK) != nil {
				// This condition may need to be reconsidered depending on the actual relationship.
				// For now, keep your original logic intact.
				joinCondition = fmt.Sprintf("%s.%s = %s.%s", targetTable, targetFK, tableName, basePK)
			}

			if !joinedTables[targetTable] {
				query = query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", targetTable, joinCondition))
				joinedTables[targetTable] = true
			}

			targetColumns := strings.Split(targetColumnsStr, ",")
			for _, col := range targetColumns {
				col = strings.TrimSpace(col)
				fullCol := col
				if !strings.Contains(col, ".") {
					fullCol = fmt.Sprintf("%s.%s", targetTable, col)
				}

				// Search by id
				if strings.HasSuffix(col, "id") {
					if idInt, err := strconv.Atoi(qb.searchValue); err == nil {
						orConditions = append(orConditions, fmt.Sprintf("%s = ?", fullCol))
						values = append(values, idInt)
						if !seenFields[fullCol] {
							sortableFields = append(sortableFields, fullCol)
							seenFields[fullCol] = true
						}
						continue
					}
				}

				// LIKE
				orConditions = append(orConditions, fmt.Sprintf("unaccent(CAST(%s AS TEXT)) ILIKE unaccent(?)", fullCol))
				values = append(values, likePattern)

				// Full-Text Search
				orConditions = append(orConditions, fmt.Sprintf("to_tsvector('simple', unaccent(CAST(%s AS TEXT))) @@ plainto_tsquery('simple', unaccent(?))", fullCol))
				values = append(values, tsQueryPattern)

				if !seenFields[fullCol] {
					sortableFields = append(sortableFields, fullCol)
					seenFields[fullCol] = true
				}
			}
		} else {
			// Search by table field
			column := field
			if !strings.Contains(field, ".") {
				column = fmt.Sprintf("%s.%s", tableName, field)
			}

			// Search by id
			if strings.HasSuffix(field, "id") {
				if idInt, err := strconv.Atoi(qb.searchValue); err == nil {
					orConditions = append(orConditions, fmt.Sprintf("%s = ?", column))
					values = append(values, idInt)
					if !seenFields[column] {
						sortableFields = append(sortableFields, column)
						seenFields[column] = true
					}
					continue
				}
			}

			// LIKE
			orConditions = append(orConditions, fmt.Sprintf("unaccent(CAST(%s AS TEXT)) ILIKE unaccent(?)", column))
			values = append(values, likePattern)

			// Full-Text Search
			orConditions = append(orConditions, fmt.Sprintf("to_tsvector('simple', unaccent(CAST(%s AS TEXT))) @@ plainto_tsquery('simple', unaccent(?))", column))
			values = append(values, tsQueryPattern)

			if !seenFields[column] {
				sortableFields = append(sortableFields, column)
				seenFields[column] = true
			}
		}
	}

	// Order by keyword
	if len(orConditions) > 0 {
		query = query.Where("("+strings.Join(orConditions, " OR ")+")", values...)

		if len(sortableFields) > 0 {
			keywordLower := strings.ToLower(qb.searchValue)

			orderExprParts := make([]string, 0, len(sortableFields))
			orderVars := make([]interface{}, 0, len(sortableFields)*3)

			for index, field := range sortableFields {
				eqWeight := index + 1        // Top ranking for exact matches
				preWeight := index*30 + 10   // Ranking for prefix matches
				inWeight := index*30 + 20    // Rank for results containing keywords
				fuzzyWeight := index*30 + 25 // Ranking for fuzzy (full-text search) results
				elseWeight := index*30 + 30  // Lowest rating

				expr := fmt.Sprintf(
					`CASE
                        WHEN LOWER(CAST(%s AS TEXT)) = ? THEN %d
                        WHEN LOWER(CAST(%s AS TEXT)) LIKE ? THEN %d
                        WHEN LOWER(CAST(%s AS TEXT)) LIKE ? THEN %d
                        WHEN to_tsvector('simple', unaccent(CAST(%s AS TEXT))) @@ plainto_tsquery('simple', unaccent(?)) THEN %d
                        ELSE %d
                    END`,
					field, eqWeight, // Exact match
					field, preWeight, // Prefix match (e.g., "con heo%")
					field, inWeight, // Contains match (e.g., "%con heo%")
					field, fuzzyWeight, // Full-text search match
					elseWeight, // Fallback
				)

				orderExprParts = append(orderExprParts, expr)
				orderVars = append(orderVars, keywordLower, keywordLower+"%", "%"+keywordLower+"%", keywordLower)
			}

			fullOrder := "(" + strings.Join(orderExprParts, " + ") + ")"
			query = query.Order(clause.Expr{SQL: fullOrder, Vars: orderVars})
		}
	}

	return query
}

func (qb *QueryBuilder[T]) ApplyFilters(query *gorm.DB) *gorm.DB {
	if !qb.isMeili {
		query = qb.ApplySearch(query)
	}

	// Get table name for the model
	var model T
	stmt := &gorm.Statement{DB: query}
	var tableName string
	if err := stmt.Parse(&model); err == nil {
		tableName = stmt.Schema.Table
	}

	for key, value := range qb.filter {
		if strings.Contains(key, ".") {
			// Check if this is a table-qualified column name (e.g., "users.status")
			parts := strings.SplitN(key, ".", 2)
			firstPart := parts[0]

			// If the first part matches the table name, treat it as a table-qualified column
			if firstPart == tableName {
				// This is a table-qualified column name, not a relation filter
				columnName := parts[1]
				lowerKey := strings.ToLower(key)

				if strVal, ok := value.(string); ok {
					found := false

					// Check BETWEEN
					if strings.HasPrefix(strVal, "between:") {
						vals := strings.Split(strings.TrimPrefix(strVal, "between:"), ",")
						if len(vals) == 2 {
							start := strings.TrimSpace(vals[0])
							end := strings.TrimSpace(vals[1])
							if start != "" && end != "" {
								query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", lowerKey), start, end)
								found = true
							}
						}
					}

					// Check operators
					if !found {
						operators := []string{">=:", "<=:", "!=:", ">:", "<:", "==:"}
						for _, op := range operators {
							if strings.HasPrefix(strVal, op) {
								val := strings.TrimPrefix(strVal, op)
								val = strings.TrimSpace(val)

								if val != "" {
									if val == "NULL" {
										if op == "!=:" {
											query = query.Where(fmt.Sprintf("%s IS NOT NULL", lowerKey))
										} else if op == "==:" {
											query = query.Where(fmt.Sprintf("%s IS NULL", lowerKey))
										}
									} else {
										if strings.Contains(strings.ToLower(columnName), "_id") && val == "0" {
											query = query.Where(fmt.Sprintf("(%s = ? OR %s IS NULL)", lowerKey, lowerKey), val)
										} else {
											query = query.Where(fmt.Sprintf("%s %s ?", lowerKey, strings.TrimSuffix(op, ":")), val)
										}
									}
									found = true
									break
								}
							}
						}
					}

					// Check NOT IN
					if !found && strings.HasPrefix(strVal, "not_in:") {
						rawIds := strings.Split(strings.TrimPrefix(strVal, "not_in:"), ",")
						var ids []interface{}
						for _, idStr := range rawIds {
							idStr = strings.TrimSpace(idStr)
							if idStr != "" {
								ids = append(ids, idStr)
							}
						}
						if len(ids) > 0 {
							query = query.Where(lowerKey+" NOT IN ?", ids)
						}
						found = true
					}

					// Check IN
					if !found && strings.HasPrefix(strVal, "in:") {
						rawIds := strings.Split(strings.TrimPrefix(strVal, "in:"), ",")
						var ids []interface{}
						for _, idStr := range rawIds {
							idStr = strings.TrimSpace(idStr)
							if idStr != "" {
								ids = append(ids, idStr)
							}
						}
						query = query.Where(lowerKey+" IN ?", ids)
						found = true
					}

					// Default =
					if !found {
						query = query.Where(lowerKey+" = ?", strVal)
					}
				} else {
					query = query.Where(lowerKey+" = ?", value)
				}
				continue
			}

			// filter by releation table
			relation := firstPart

			if strVal, ok := value.(string); ok {
				// Check if this is IN clause with relation format: "in:1,2,3:user_classes:users:classes:user_id:class_id:id:id"
				if strings.HasPrefix(strVal, "in:") && strings.Count(strVal, ":") >= 7 {
					// Split by ":" to extract parts after "in:"
					allParts := strings.SplitN(strVal, ":", 9)
					if len(allParts) == 9 {
						// allParts[0] = "in"
						// allParts[1] = "1,2,3" (the IN values)
						// allParts[2-8] = relation info
						inValues := allParts[1]
						relationRefTable := allParts[2]
						currentTable := allParts[3]
						relationTable := allParts[4]
						currentRefKey := allParts[5]
						relationRefKey := allParts[6]
						currentKey := allParts[7]
						relationKey := allParts[8]

						// Parse IN values
						rawIds := strings.Split(inValues, ",")
						var ids []interface{}
						for _, idStr := range rawIds {
							idStr = strings.TrimSpace(idStr)
							if idStr != "" {
								ids = append(ids, idStr)
							}
						}

						if len(ids) > 0 {
							joinOn := fmt.Sprintf("%s.%s = %s.%s", relationRefTable, currentRefKey, currentTable, currentKey)
							whereClause := fmt.Sprintf("%s.%s IN ?", relationRefTable, relationRefKey)
							relationWhere := fmt.Sprintf("%s.%s IN ?", relationTable, relationKey)

							query = query.
								Joins("JOIN "+relationRefTable+" ON "+joinOn).
								Where(whereClause, ids).
								Preload(relation, func(db *gorm.DB) *gorm.DB {
									return db.Where(relationWhere, ids)
								})
						}
					}
				} else if strings.Count(strVal, ":") == 7 {
					// Example "1:user_courses:users:courses:user_id:course_id:id:id", Relation Courses
					valueParts := strings.SplitN(strVal, ":", 8)

					filterValue, relationRefTable, currentTable, relationTable :=
						valueParts[0], valueParts[1], valueParts[2], valueParts[3]
					currentRefKey, relationRefKey, currentKey, relationKey :=
						valueParts[4], valueParts[5], valueParts[6], valueParts[7]

					if filterValue == "0" {
						query = query.
							Joins(fmt.Sprintf("LEFT JOIN %s ON %s.%s = %s.%s",
								relationRefTable, relationRefTable, currentRefKey, currentTable, currentKey)).
							Where(fmt.Sprintf("%s.%s IS NULL", relationRefTable, relationRefKey))
					} else {
						// Has value
						joinOn := fmt.Sprintf("%s.%s = %s.%s", relationRefTable, currentRefKey, currentTable, currentKey)
						whereClause := fmt.Sprintf("%s.%s = ?", relationRefTable, relationRefKey)
						relationWhere := fmt.Sprintf("%s.%s = ?", relationTable, relationKey)

						query = query.
							Joins("JOIN "+relationRefTable+" ON "+joinOn).
							Where(whereClause, filterValue).
							Preload(relation, func(db *gorm.DB) *gorm.DB {
								return db.Where(relationWhere, filterValue)
							})
					}
				} else if strings.Count(strVal, ":") == 5 {
					// Example "01::users:user_address:id:user_id:province_code", Relation UserAddress
					parts := strings.SplitN(strVal, ":", 6)

					filterValue := parts[0]
					currentTable := parts[1]
					relationTable := parts[2]
					currentKey := parts[3]
					relationKey := parts[4]
					filterField := parts[5]

					joinOn := fmt.Sprintf("%s.%s = %s.%s", currentTable, currentKey, relationTable, relationKey)
					preloadWhere := fmt.Sprintf("%s.%s = ?", relationTable, filterField)

					if filterValue == "0" {
						query = query.
							Joins("LEFT JOIN " + relationTable + " ON " + joinOn).
							Where(fmt.Sprintf("%s.%s IS NULL", relationTable, filterField))
					} else {
						// Has value
						query = query.
							Joins("JOIN "+relationTable+" ON "+joinOn).
							Where(preloadWhere, filterValue).
							Preload(relation, func(db *gorm.DB) *gorm.DB {
								return db.Where(preloadWhere, filterValue)
							})
					}
				} else if strings.HasPrefix(strVal, "in:") {
					rawIds := strings.Split(strings.TrimPrefix(strVal, "in:"), ",")
					var ids []interface{}
					for _, idStr := range rawIds {
						idStr = strings.TrimSpace(idStr)
						if idStr != "" {
							ids = append(ids, idStr)
						}
					}

					query = query.Where(key+" IN ?", ids)
				}
			} else {
				// todo skip filter
				return query
			}
		} else {
			// filter by table field
			lowerKey := strings.ToLower(key)

			if strVal, ok := value.(string); ok {
				found := false

				// Check BETWEEN
				// Example "between:2006-01-02,2006-01-03"
				if strings.HasPrefix(strVal, "between:") {
					vals := strings.Split(strings.TrimPrefix(strVal, "between:"), ",")
					if len(vals) == 2 {
						start := strings.TrimSpace(vals[0])
						end := strings.TrimSpace(vals[1])
						if start != "" && end != "" {
							query = query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", lowerKey), start, end)
							found = true
						}
					}
				}

				// Check operators
				// Example ">=:2006-01-02"
				if !found {
					operators := []string{">=:", "<=:", "!=:", ">:", "<:", "==:"}
					for _, op := range operators {
						if strings.HasPrefix(strVal, op) {
							val := strings.TrimPrefix(strVal, op)
							val = strings.TrimSpace(val)

							if val != "" {
								if val == "NULL" {
									if op == "!=:" {
										query = query.Where(fmt.Sprintf("%s IS NOT NULL", lowerKey))
									} else if op == "==:" {
										query = query.Where(fmt.Sprintf("%s IS NULL", lowerKey))
									}
								} else {
									// If it is _id field and val == "0", then add special condition
									if strings.Contains(lowerKey, "_id") && val == "0" {
										query = query.Where(fmt.Sprintf("(%s = ? OR %s IS NULL)", lowerKey, lowerKey), val)
									} else {
										query = query.Where(fmt.Sprintf("%s %s ?", lowerKey, strings.TrimSuffix(op, ":")), val)
									}
								}
								found = true
								break
							}
						}
					}
				}

				// Check NOT IN
				// Example "not_in:1,2,3"
				if !found && strings.HasPrefix(strVal, "not_in:") {
					rawIds := strings.Split(strings.TrimPrefix(strVal, "not_in:"), ",")
					var ids []interface{}
					for _, idStr := range rawIds {
						idStr = strings.TrimSpace(idStr)
						if idStr != "" {
							ids = append(ids, idStr)
						}
					}

					if len(ids) > 0 {
						query = query.Where(lowerKey+" NOT IN ?", ids)
					}
					found = true
				}

				// Check IN
				// Example "in:1,2,3"
				if !found && strings.HasPrefix(strVal, "in:") {
					rawIds := strings.Split(strings.TrimPrefix(strVal, "in:"), ",")
					var ids []interface{}
					for _, idStr := range rawIds {
						idStr = strings.TrimSpace(idStr)
						if idStr != "" {
							ids = append(ids, idStr)
						}
					}

					query = query.Where(lowerKey+" IN ?", ids)
					found = true
				}

				// Default =
				if !found {
					query = query.Where(lowerKey+" = ?", strVal)
				}
			} else {
				query = query.Where(lowerKey+" = ?", value)
			}

		}
	}
	return query
}

func (qb *QueryBuilder[T]) ApplyFiltersAndPagination(query *gorm.DB) *gorm.DB {
	query = qb.ApplyFilters(query)
	offset := (qb.page - 1) * qb.limit
	return query.Limit(qb.limit).Offset(offset)
}

func (qb *QueryBuilder[T]) ApplySort(query *gorm.DB) *gorm.DB {
	if qb.searchValue != "" {
		return query
	}

	for _, v := range qb.sort {
		order := strings.ToLower(v)
		if order == "newest" {
			return query.Order("created_at DESC")
		}
		if order == "oldest" {
			return query.Order("created_at ASC")
		}
	}

	for field, order := range qb.sort {
		lowerField := strings.ToLower(field)
		query = query.Order(fmt.Sprintf("%s %s", lowerField, order))
	}

	return query
}

func (qb *QueryBuilder[T]) ApplyAlias(query *gorm.DB) *gorm.DB {
	selects := make([]string, 0, len(qb.alias))

	// rename field in table
	if len(qb.alias) > 0 {
		hasStar := false
		allSelect := "*"
		for field, alias := range qb.alias {
			if field == "*" {
				hasStar = true
				allSelect = alias
				continue
			}
			selects = append(selects, field+" AS "+alias)
		}

		if hasStar {
			selects = append([]string{allSelect}, selects...)
		}

		query = query.Select(strings.Join(selects, ", "))
	}

	return query
}

func (qb *QueryBuilder[T]) ApplyPreload(query *gorm.DB) *gorm.DB {
	for _, preload := range qb.preload {
		query = query.Preload(preload)
	}
	return query
}

func FilterProgram(query *gorm.DB, ctx *gin.Context) *gorm.DB {
	programId := ctx.Query("program_id")

	if programId != "" {
		query = query.Where("program_id = ?", programId)
	} else {
		byProgram := ctx.Query("by_program")

		if byProgram == "true" || byProgram == "1" || byProgram == "yes" {
			query = query.Where("program_id >= 1")
		} else {
			query = query.Where("program_id IS NULL OR program_id = 0")
		}
	}

	return query
}
