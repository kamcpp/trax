package apiv1

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/trax"
)

// listSagaInstancesAllowlist is the per-table contract consulted by this
// handler. Unknown sort fields are rejected with 400; empty sort_fields falls
// back to the store's declared default.
var listSagaInstancesAllowlist = common.ListingAllowlist{
	SortableFields:    trax.SagaInstancesSortableFields,
	SearchableColumns: trax.SagaInstancesSearchColumns,
	Default:           trax.SagaInstancesDefaultSort,
}

// @version v1
// @router /saga-instances/list [post]
// @summary List saga instances for a cluster with the unified pagination/search/sort contract.
// @schemes
// @tags saga
// @produce json
// @param request body listSagaInstancesRequest true "page_nr/page_size, search, sort_fields, cluster_id"
// @success 200 object listSagaInstancesResponse
// @failure 400 "invalid request"
// @failure 500 "internal server error; check server logs"
func postSagaInstances(c *gin.Context) {
	var req listSagaInstancesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := fmt.Sprintf("invalid request body: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	if req.ClusterId == "" {
		errMsg := "cluster ID is required"
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(400, errMsg)
		return
	}

	opts, err := buildSagaListQueryOptions(&req)
	if err != nil {
		common.L.Error(err.Error(), common.F(c)...)
		c.JSON(400, err.Error())
		return
	}

	sagaInstances, totalCount, err := traxStore.ListSagaInstancesPaginated(c, req.ClusterId, opts)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list saga instances for cluster %q: %v", req.ClusterId, err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(500, errMsg)
		return
	}

	instances := make([]sagaInstanceResponse, 0, len(sagaInstances))
	for _, instance := range sagaInstances {
		instanceResp, err := convertSagaInstanceToResponse(instance)
		if err != nil {
			errMsg := fmt.Sprintf("failed to convert saga instance to response: %v", err)
			common.L.Error(errMsg, common.F(c)...)
			c.JSON(500, errMsg)
			return
		}
		instances = append(instances, *instanceResp)
	}

	resp := listSagaInstancesResponse{
		SagaInstances: instances,
		TotalCount:    &totalCount,
	}
	c.JSON(200, resp)
}

// buildSagaListQueryOptions translates the wire-level request into a
// *common.QueryOptions. Defaults page_nr=1, page_size=50; clamps page_size to
// [1, 500]. Unknown sort fields are rejected so callers learn early.
func buildSagaListQueryOptions(req *listSagaInstancesRequest) (*common.QueryOptions, error) {
	pageNr := 1
	pageSize := 50
	if req.PageNr != nil && *req.PageNr > 0 {
		pageNr = *req.PageNr
	}
	if req.PageSize != nil && *req.PageSize > 0 {
		pageSize = *req.PageSize
	}
	if pageSize > 500 {
		pageSize = 500
	}

	sortCols := make([]common.SortColumn, 0, len(req.SortFields))
	for _, sf := range req.SortFields {
		field := strings.TrimSpace(sf.Field)
		if field == "" {
			return nil, fmt.Errorf("sort_fields contains an entry with empty field")
		}
		col, ok := listSagaInstancesAllowlist.SortableFields[field]
		if !ok {
			return nil, fmt.Errorf("sort_fields[%q] is not a sortable column on saga_instances", field)
		}
		dir := "ASC"
		if strings.EqualFold(sf.Direction, "desc") {
			dir = "DESC"
		}
		sortCols = append(sortCols, common.SortColumn{Column: col, Direction: dir})
	}

	filters := map[string]string{}
	if req.State != "" {
		filters["state"] = req.State
	}
	if req.SagaTemplateId != "" {
		filters["saga_template_id"] = req.SagaTemplateId
	}
	if req.SagaSubmitterId != "" {
		filters["saga_submitter_id"] = req.SagaSubmitterId
	}

	return &common.QueryOptions{
		Offset:  (pageNr - 1) * pageSize,
		Limit:   pageSize,
		Search:  req.Search,
		SortBy:  sortCols,
		Filters: filters,
	}, nil
}
