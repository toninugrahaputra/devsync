package main

import (
	"devsync/internal/handler"
	"devsync/internal/middleware"
	"devsync/internal/repository"
	"devsync/internal/service"
	"devsync/pkg/database"
	"devsync/pkg/mailer"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Loads .env into the process environment when present; a real deployment
	// that sets env vars directly (no .env file on disk) hits the ignored error
	// path here and keeps working unchanged.
	_ = godotenv.Load()

	db := database.InitDB()
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	boardRepo := repository.NewBoardRepository(db)
	listRepo := repository.NewListRepository(db)
	cardRepo := repository.NewCardRepository(db)
	labelRepo := repository.NewLabelRepository(db)
	checklistRepo := repository.NewChecklistRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	inviteRepo := repository.NewInviteRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// Services
	authService := service.NewAuthService(userRepo)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	boardService := service.NewBoardService(workspaceRepo, boardRepo, listRepo, cardRepo, labelRepo, checklistRepo)
	listService := service.NewListService(boardRepo, listRepo)
	cardService := service.NewCardService(boardRepo, listRepo, cardRepo, labelRepo, checklistRepo, commentRepo, attachmentRepo, activityRepo, userRepo)
	labelService := service.NewLabelService(boardRepo, labelRepo)
	checklistService := service.NewChecklistService(boardRepo, cardRepo, checklistRepo, activityRepo)
	commentService := service.NewCommentService(boardRepo, cardRepo, commentRepo)
	inviteService := service.NewInviteService(workspaceRepo, userRepo, inviteRepo, mailer.New())

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userRepo)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	boardHandler := handler.NewBoardHandler(boardService)
	listHandler := handler.NewListHandler(listService)
	cardHandler := handler.NewCardHandler(cardService)
	labelHandler := handler.NewLabelHandler(labelService)
	checklistHandler := handler.NewChecklistHandler(checklistService)
	commentHandler := handler.NewCommentHandler(commentService)
	inviteHandler := handler.NewInviteHandler(inviteService)

	r := gin.Default()

	// CORS middleware (basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Static("/uploads", "./uploads")

	api := r.Group("/api/v1")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/google", authHandler.GoogleLogin)
		api.GET("/invites/:token", inviteHandler.Preview)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/users", userHandler.Search)
			protected.PUT("/users/me", userHandler.UpdateProfile)
			protected.POST("/users/me/avatar", userHandler.UploadAvatar)

			protected.GET("/workspaces", workspaceHandler.ListMine)
			protected.POST("/workspaces", workspaceHandler.Create)
			protected.GET("/workspaces/:id", workspaceHandler.Get)
			protected.PUT("/workspaces/:id", workspaceHandler.Update)
			protected.DELETE("/workspaces/:id", workspaceHandler.Delete)
			protected.GET("/workspaces/:id/members", workspaceHandler.ListMembers)
			protected.POST("/workspaces/:id/members", workspaceHandler.AddMember)
			protected.DELETE("/workspaces/:id/members/:userId", workspaceHandler.RemoveMember)
			protected.POST("/workspaces/:id/invites", inviteHandler.Create)
			protected.GET("/workspaces/:id/boards", boardHandler.ListByWorkspace)
			protected.POST("/workspaces/:id/boards", boardHandler.Create)

			protected.GET("/boards/:id", boardHandler.GetDetail)
			protected.PUT("/boards/:id", boardHandler.Update)
			protected.DELETE("/boards/:id", boardHandler.Delete)
			protected.POST("/boards/:id/members", boardHandler.AddMember)
			protected.DELETE("/boards/:id/members/:userId", boardHandler.RemoveMember)
			protected.GET("/boards/:id/labels", labelHandler.ListByBoard)
			protected.POST("/boards/:id/labels", labelHandler.Create)
			protected.POST("/boards/:id/lists", listHandler.Create)
			protected.GET("/boards/:id/archived-cards", cardHandler.ListArchived)

			protected.PUT("/labels/:id", labelHandler.Update)
			protected.DELETE("/labels/:id", labelHandler.Delete)

			protected.PUT("/lists/:id", listHandler.Rename)
			protected.PUT("/lists/:id/move", listHandler.Move)
			protected.DELETE("/lists/:id", listHandler.Archive)
			protected.POST("/lists/:id/cards", cardHandler.Create)

			protected.GET("/cards/:id", cardHandler.GetDetail)
			protected.PUT("/cards/:id", cardHandler.Update)
			protected.PUT("/cards/:id/move", cardHandler.Move)
			protected.DELETE("/cards/:id", cardHandler.Archive)
			protected.PUT("/cards/:id/restore", cardHandler.Restore)
			protected.DELETE("/cards/:id/permanent", cardHandler.PermanentDelete)
			protected.POST("/cards/:id/members/:userId", cardHandler.AddMember)
			protected.DELETE("/cards/:id/members/:userId", cardHandler.RemoveMember)
			protected.POST("/cards/:id/labels/:labelId", cardHandler.AddLabel)
			protected.DELETE("/cards/:id/labels/:labelId", cardHandler.RemoveLabel)
			protected.POST("/cards/:id/checklists", checklistHandler.Create)
			protected.GET("/cards/:id/comments", commentHandler.ListByCard)
			protected.POST("/cards/:id/comments", commentHandler.Create)
			protected.POST("/cards/:id/attachments", cardHandler.AddAttachment)
			protected.DELETE("/attachments/:id", cardHandler.DeleteAttachment)

			protected.PUT("/checklists/:id", checklistHandler.Update)
			protected.DELETE("/checklists/:id", checklistHandler.Delete)
			protected.POST("/checklists/:id/items", checklistHandler.CreateItem)

			protected.PUT("/checklist-items/:id", checklistHandler.UpdateItem)
			protected.DELETE("/checklist-items/:id", checklistHandler.DeleteItem)

			protected.DELETE("/comments/:id", commentHandler.Delete)

			protected.POST("/invites/:token/accept", inviteHandler.Accept)
		}
	}

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
