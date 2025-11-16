package http

import (
	"github.com/gin-gonic/gin"
	"github.com/raccoon00/avito-pr/internal/service"
)

func Run(s *service.Service) {
	r := gin.Default()

	gs := GinService{srv: s}

	r.POST("/team/add", gs.TeamAdd)
	r.GET("/team/get", gs.TeamGet)
	r.POST("/users/setIsActive", gs.SetUserIsActive)
	r.POST("/pullRequest/create", gs.CreatePullRequest)
	r.POST("/pullRequest/reassign", gs.ReassignReviewer)
	r.POST("/pullRequest/merge", gs.MergePullRequest)
	r.GET("/users/getReview", gs.GetUserReviews)

	r.Run()
}
