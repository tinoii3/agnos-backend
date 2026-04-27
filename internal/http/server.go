package http

import (
	"errors"
	"net/http"
	"strings"

	"agnos-backend/internal/auth"
	"agnos-backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	contextClaimsKey = "claims"
)

type Server struct {
	store     store.Store
	jwtSecret string
}

func NewServer(store store.Store, jwtSecret string) *Server {
	return &Server{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

func (s *Server) Router() *gin.Engine {
	router := gin.Default()

	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/staff/create", s.handleCreateStaff)
	router.POST("/staff/login", s.handleLogin)

	protected := router.Group("/")
	protected.Use(s.JWTAuthMiddleware())
	protected.GET("/patient/search", s.handlePatientSearch)

	return router
}

type createStaffRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Hospital string `json:"hospital" binding:"required"`
}

func (s *Server) handleCreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure password"})
		return
	}

	err = s.store.CreateStaff(c.Request.Context(), req.Username, string(hash), req.Hospital)
	if err != nil {
		if store.IsUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create staff"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "staff created",
		"username": req.Username,
		"hospital": req.Hospital,
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

func (s *Server) handleLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff, err := s.store.GetStaffByUsername(c.Request.Context(), req.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.HashedPassword), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if staff.Hospital != req.Hospital {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(s.jwtSecret, staff.Username, staff.Hospital)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *Server) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		claims, err := auth.ParseToken(s.jwtSecret, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(contextClaimsKey, claims)
		c.Next()
	}
}

func (s *Server) handlePatientSearch(c *gin.Context) {
	claimsAny, ok := c.Get(contextClaimsKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth context"})
		return
	}

	claims, ok := claimsAny.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth context"})
		return
	}

	filters := store.PatientFilters{
		ID:          c.Query("id"),
		NationalID:  c.Query("national_id"),
		PassportID:  c.Query("passport_id"),
		FirstName:   c.Query("first_name"),
		MiddleName:  c.Query("middle_name"),
		LastName:    c.Query("last_name"),
		DateOfBirth: c.Query("date_of_birth"),
		PhoneNumber: c.Query("phone_number"),
		Email:       c.Query("email"),
	}

	patients, err := s.store.SearchPatients(c.Request.Context(), claims.Hospital, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search patients"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patients": patients})
}
