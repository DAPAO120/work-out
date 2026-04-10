package api

import (
	"Project001/internal/controller"
	"Project001/internal/service/article"
	"Project001/middleware"

	"github.com/gin-gonic/gin"
)

type ArticleRouter struct{}

func (dr *ArticleRouter) InitApiRouter(parent *gin.RouterGroup) {
	// 依赖注入
	articleCtrl := controller.NewArticleController(article.NewArticlePostService())
	// 私有路由使用jwt验证
	privateRouter := parent.Group("articleapi")
	privateRouter.Use(middleware.JWTAuth())
	{
		privateRouter.POST("/post", articleCtrl.CreatePost)
		privateRouter.GET("/list", articleCtrl.GetPostList)
		privateRouter.GET("/favorite", articleCtrl.GetFavoritePosts)

		privateRouter.POST("/favorite", articleCtrl.FavoritePost)
		privateRouter.POST("/deleteFavorite", articleCtrl.CancelFavorite)

		privateRouter.GET("/deletePost", articleCtrl.DeletePost)

		privateRouter.GET("/featured", articleCtrl.GetFeaturedPosts)
		privateRouter.GET("/setFeatured", articleCtrl.SetFeatured)
		privateRouter.GET("/deleteFeatured", articleCtrl.CancelFeatured)

		privateRouter.GET("/detail", articleCtrl.GetPostDetail)

		privateRouter.POST("/comment", articleCtrl.CreateComment)

		privateRouter.GET("/search", articleCtrl.SearchPosts)

		privateRouter.GET("/deleteComment", articleCtrl.DeleteComment)
		privateRouter.GET("/getPostsByUserId", articleCtrl.GetUserPosts)
	}

}
