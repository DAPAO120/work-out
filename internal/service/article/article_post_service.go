package article

import (
	"Project001/global"
	"Project001/internal/model"
	impl "Project001/internal/service/article/Impl"
	"time"
)

type (
	IArticlePostService interface {
		// 获取帖子列表
		GetPostList(userID int64, cursorTime time.Time,
			cursorID int64, pageSize int) ([]model.Post, error)
		// 获取收藏列表
		GetFavoritePosts(userID int64, cursorTime time.Time,
			cursorID int64, pageSize int) ([]model.Post, error)
		// 收藏
		FavoritePost(userID, postID int64) error
		//取消收藏
		CancelFavorite(userID, postID int64) error
		//删除帖子
		DeletePost(userID int64, postID int64, isAdmin bool) error
		//管理员设置精品
		SetPostFeatured(postID int64) error
		//取消精品
		CancelFeatured(postID int64) error
		//精品帖子列表
		GetFeaturedPosts(cursorTime time.Time,
			cursorID int64, pageSize int) ([]model.Post, error)
		//帖子详情
		GetPostDetail(userID int64, postID int64, cursorTime time.Time,
			cursorID int64, pageSize int) (*model.Post, []model.PostComment, bool, interface{}, error) //创建评论
		CreateComment(userID, postID, parentID int64, content string, images []model.CommentImage) error
		//新建帖子
		CreatePost(userID int64, content string, title string, images []string) error
		//搜索
		SearchPosts(req model.SearchPostsModel) ([]model.Post, error)
		// 删除评论
		DeleteComment(userID int64, commentID int64, isAdmin bool) error
		// 获取用户的所有帖子
		GetUserPosts(userID int64, targetUserID int64, cursorTime time.Time, cursorID int64, pageSize int) ([]model.Post, error)
		//新建帖子上传图片

		//创建评论上传图片
	}
)

func NewArticlePostService() IArticlePostService {
	return &impl.ArticlePostServiceImpl{
		DB:    global.DB,
		Redis: global.Redis,
	}
}
