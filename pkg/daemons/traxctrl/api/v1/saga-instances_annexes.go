package apiv1

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kamcpp/trax/pkg/common"
	"github.com/kamcpp/trax/pkg/trax"
)

// createSagaAnnexRequest is the JSON shape for
// POST /saga-instances/:sagaInstanceId/annexes. Bytes ride in
// `content_data_base64` because the rest of the traxctrl API is
// JSON; multipart support can come later if payloads grow large.
type createSagaAnnexRequest struct {
	ClusterId         string `json:"cluster_id" binding:"required"`
	Iid               string `json:"iid" binding:"required"`
	ContentType       string `json:"content_type"`
	ContentDataBase64 string `json:"content_data_base64" binding:"required"`
	Notes             string `json:"notes,omitempty"`
}

// sagaAnnexResponse is the metadata projection — bytes are NEVER
// included on this shape (download endpoint serves them).
type sagaAnnexResponse struct {
	Iid            string `json:"iid"`
	SagaInstanceId string `json:"saga_instance_id"`
	ContentType    string `json:"content_type"`
	ContentLength  int64  `json:"content_length"`
	Notes          string `json:"notes,omitempty"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

type listSagaAnnexesResponse struct {
	SagaInstanceId string              `json:"saga_instance_id"`
	Annexes        []sagaAnnexResponse `json:"annexes"`
}

// @version v1
// @router /saga-instances/{sagaInstanceId}/annexes [post]
// @summary Upload a binary annex attached to a saga
// @description Trax owns the bytes of any binary payload tied to
// @description a saga (csdmsggw is the primary writer for batch-
// @description issue-security-units uploads). Body is JSON with
// @description content_data_base64; idempotent on iid.
// @tags saga
// @accept json
// @produce json
// @param sagaInstanceId path string true "Saga Instance ID"
// @param request body createSagaAnnexRequest true "Annex payload"
// @success 200 object sagaAnnexResponse
// @failure 400 "missing required fields / bad base64"
// @failure 404 "saga instance not found"
// @failure 500 "internal server error; check server logs"
func postSagaAnnex(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	if sagaInstanceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga instance ID is required"})
		return
	}
	var req createSagaAnnexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	bytes, err := base64.StdEncoding.DecodeString(req.ContentDataBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("content_data_base64 is not valid base64: %v", err)})
		return
	}

	annex := &trax.SagaAnnex{
		Iid:            req.Iid,
		ClusterId:      req.ClusterId,
		SagaInstanceId: sagaInstanceId,
		ContentType:    req.ContentType,
		ContentLength:  int64(len(bytes)),
		ContentData:    bytes,
		Notes:          req.Notes,
	}
	if err := traxStore.CreateSagaAnnex(c, annex); err != nil {
		if err == trax.ErrSagaInstanceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "saga instance not found"})
			return
		}
		errMsg := fmt.Sprintf("failed to create saga annex: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	c.JSON(http.StatusOK, sagaAnnexResponse{
		Iid:            annex.Iid,
		SagaInstanceId: annex.SagaInstanceId,
		ContentType:    annex.ContentType,
		ContentLength:  annex.ContentLength,
		Notes:          annex.Notes,
	})
}

// @version v1
// @router /saga-instances/{sagaInstanceId}/annexes [get]
// @summary List annex metadata for a saga
// @tags saga
// @produce json
// @param sagaInstanceId path string true "Saga Instance ID"
// @param cluster_id query string true "Cluster ID"
// @success 200 object listSagaAnnexesResponse
// @failure 400 "missing required query parameter"
// @failure 500 "internal server error"
func getSagaAnnexes(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	clusterId := c.Query("cluster_id")
	if sagaInstanceId == "" || clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga instance ID and cluster_id are required"})
		return
	}
	rows, err := traxStore.ListSagaAnnexes(c, clusterId, sagaInstanceId)
	if err != nil {
		errMsg := fmt.Sprintf("failed to list saga annexes: %v", err)
		common.L.Error(errMsg, common.F(c)...)
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}
	resp := listSagaAnnexesResponse{SagaInstanceId: sagaInstanceId}
	for _, a := range rows {
		resp.Annexes = append(resp.Annexes, sagaAnnexResponse{
			Iid:            a.Iid,
			SagaInstanceId: a.SagaInstanceId,
			ContentType:    a.ContentType,
			ContentLength:  a.ContentLength,
			Notes:          a.Notes,
			CreatedAt:      a.CreatedAt,
			UpdatedAt:      a.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, resp)
}

// @version v1
// @router /saga-instances/{sagaInstanceId}/annexes/{annexIid} [get]
// @summary Download annex bytes
// @description Cross-saga retrieval is rejected — annexIid must
// @description belong to sagaInstanceId.
// @tags saga
// @produce application/octet-stream
// @param sagaInstanceId path string true "Saga Instance ID"
// @param annexIid path string true "Annex IID"
// @param cluster_id query string true "Cluster ID"
// @success 200 {file} file "Raw annex bytes; Content-Type from stored row"
// @failure 400 "missing required parameter"
// @failure 404 "annex not found, or annex doesn't belong to saga"
// @failure 500 "internal server error"
func getSagaAnnexBytes(c *gin.Context) {
	sagaInstanceId := c.Param("sagaInstanceId")
	annexIid := c.Param("annexIid")
	clusterId := c.Query("cluster_id")
	if sagaInstanceId == "" || annexIid == "" || clusterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saga instance ID, annex iid, and cluster_id are required"})
		return
	}
	annex, err := traxStore.GetSagaAnnexBytes(c, clusterId, annexIid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "annex not found"})
		return
	}
	if annex.SagaInstanceId != sagaInstanceId {
		// Don't leak existence on other sagas — surface as 404.
		c.JSON(http.StatusNotFound, gin.H{"error": "annex not found"})
		return
	}
	contentType := annex.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, annex.ContentData)
}
