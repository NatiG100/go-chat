package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"example.com/go-chat/internal/core"
	"example.com/go-chat/internal/core/usecases"
	"example.com/go-chat/internal/drivers"
)

type Handler struct {
	repos core.Repositories
	rds   *drivers.RedisClient
	jwt   *drivers.JWTManager
	authU *usecases.AuthUsecase
	chatU *usecases.ChatUsecase
}

func NewHandler(repos core.Repositories, rds *drivers.RedisClient, jwt *drivers.JWTManager) *Handler {
	return &Handler{repos: repos, rds: rds, jwt: jwt, authU: usecases.NewAuthUsecase(repos, jwt), chatU: usecases.NewChatUsecase(repos, rds)}
}

func (h *Handler) SignUp(c *gin.Context) {
	var body struct{ Username, Email, Password string }
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.authU.SignUp(c.Request.Context(), body.Username, body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *Handler) Login(c *gin.Context) {
	var body struct{ Email, Password string }
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, user, err := h.authU.Login(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (h *Handler) Me(c *gin.Context) {
	idI, _ := c.Get("user_id")
	id := idI.(uuid.UUID)
	u, err := h.authU.Me(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) CreateGroup(c *gin.Context) {
	var body struct{ Name string }
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	idI, _ := c.Get("user_id")
	owner := idI.(uuid.UUID)
	g := &core.Group{Name: body.Name, OwnerID: owner}
	if err := h.repos.GroupRepo().CreateGroup(c.Request.Context(), g); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.repos.GroupRepo().AddGroupMember(c.Request.Context(), g.ID, owner)
	c.JSON(http.StatusCreated, g)
}
func (h *Handler) MyGroups(c *gin.Context){
	
	idI, _ := c.Get("user_id")
	uid := idI.(uuid.UUID)
	groups,err := h.repos.GroupRepo().MyGroups(c.Request.Context(),uid)
	if err!=nil{
		c.JSON(http.StatusNotFound, gin.H{"message":err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)

}

func (h *Handler) JoinGroup(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	idI, _ := c.Get("user_id")
	uid := idI.(uuid.UUID)
	if err := h.repos.GroupRepo().AddGroupMember(c.Request.Context(), gid, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) ListGroupMembers(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	members, err := h.repos.GroupRepo().ListGroupMembers(c.Request.Context(), gid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *Handler) GetPrivateHistory(c *gin.Context) {
	other := c.Query("user_id")
	if other == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing user_id"})
		return
	}
	otherID, err := uuid.Parse(other)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	idI, _ := c.Get("user_id")
	selfID := idI.(uuid.UUID)
	limit := 50
	msgs, err := h.chatU.GetPrivateHistory(c.Request.Context(), selfID, otherID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

func (h *Handler) GetGroupHistory(c *gin.Context) {
	gid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	msgs, err := h.chatU.GetGroupHistory(c.Request.Context(), gid, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}
