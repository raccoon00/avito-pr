package http

import (
	"github.com/gin-gonic/gin"
	"github.com/raccoon00/avito-pr/internal/service"
)

func Run(s *service.Service) {
	r := gin.Default()

	gs := GinService{srv: s}

	r.GET("/team/add", gs.TeamAdd)

	r.Run()
}
