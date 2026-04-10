package impl

import (
	"Project001/common/cache"
	"Project001/internal/dao"
	"Project001/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type ArticlePostServiceImpl struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// 获取帖子列表
func (s *ArticlePostServiceImpl) GetPostList(userID int64, cursorTime time.Time,
	cursorID int64, pageSize int) ([]model.Post, error) {

	posts, err := dao.GetPostList(s.DB, cursorTime, cursorID, pageSize)

	if err != nil {
		return nil, err
	}

	var ids []int64

	for _, p := range posts {
		ids = append(ids, p.ID)
	}

	favMap := dao.GetUserFavoritePostIDs(s.DB, userID, ids)

	for i := range posts {

		if favMap[posts[i].ID] {
			posts[i].IsFavorite = true
		}
	}

	return posts, nil
}

// 获取收藏列表
func (s *ArticlePostServiceImpl) GetFavoritePosts(userID int64, cursorTime time.Time,
	cursorID int64, pageSize int) ([]model.Post, error) {

	posts, err := dao.GetFavoritePosts(s.DB, userID, cursorTime, cursorID, pageSize)

	if err != nil {
		return nil, err
	}

	for i := range posts {
		posts[i].IsFavorite = true
	}

	return posts, nil
}

// 收藏
func (s *ArticlePostServiceImpl) FavoritePost(userID, postID int64) error {
	key := fmt.Sprintf(cache.PostComment, postID)
	s.Redis.Del(key)
	key2 := fmt.Sprintf(cache.PostDetail, postID)
	s.Redis.Del(key2)
	return dao.CreateFavorite(s.DB, userID, postID)
}

// 取消收藏
func (s *ArticlePostServiceImpl) CancelFavorite(userID, postID int64) error {
	key := fmt.Sprintf(cache.PostComment, postID)
	s.Redis.Del(key)
	key2 := fmt.Sprintf(cache.PostDetail, postID)
	s.Redis.Del(key2)
	return dao.DeleteFavorite(s.DB, userID, postID)
}

// 删除帖子
func (s *ArticlePostServiceImpl) DeletePost(userID int64, postID int64, isAdmin bool) error {

	post, err := dao.GetPostDetail(s.DB, postID)

	if err != nil {
		return err
	}

	if post.UserID != userID && !isAdmin {
		return errors.New("no permission")
	}

	key := fmt.Sprintf(cache.PostComment, postID)
	s.Redis.Del(key)
	key2 := fmt.Sprintf(cache.PostDetail, postID)
	s.Redis.Del(key2)
	return dao.DeletePost(s.DB, postID)
}

// 管理员设置精品
func (s *ArticlePostServiceImpl) SetPostFeatured(postID int64) error {

	err := dao.SetFeatured(s.DB, postID, true)
	if err == nil {

		key := fmt.Sprintf(cache.PostComment, postID)

		s.Redis.Del(key)

		key2 := fmt.Sprintf(cache.PostDetail, postID)

		s.Redis.Del(key2)
	}
	return err
}

// 取消精品
func (s *ArticlePostServiceImpl) CancelFeatured(postID int64) error {
	key := fmt.Sprintf(cache.PostComment, postID)
	s.Redis.Del(key)
	key2 := fmt.Sprintf(cache.PostDetail, postID)
	s.Redis.Del(key2)
	return dao.CancelFeatured(s.DB, postID)
}

// 精品帖子列表
func (s *ArticlePostServiceImpl) GetFeaturedPosts(cursorTime time.Time,
	cursorID int64, pageSize int) ([]model.Post, error) {
	return dao.GetFeaturedPosts(s.DB, cursorTime, cursorID, pageSize)
}

// 获取帖子详情（包含评论分页）
func (s *ArticlePostServiceImpl) GetPostDetail(userID int64, postID int64, cursorTime time.Time, cursorID int64, pageSize int) (*model.Post, []model.PostComment, bool, interface{}, error) {

	// 先从缓存获取帖子详情
	key := fmt.Sprintf(cache.PostDetail, postID)
	var post model.Post

	val, _ := s.Redis.Get(key).Result()
	if val != "" {
		json.Unmarshal([]byte(val), &post)
	} else {
		// 从数据库获取帖子详情
		p, err := dao.GetPostDetail(s.DB, postID)
		if err != nil {
			return nil, nil, false, nil, err
		}

		var ids []int64
		ids = append(ids, postID)
		favMap := dao.GetUserFavoritePostIDs(s.DB, userID, ids)
		if favMap[postID] {
			p.IsFavorite = true
		}
		post = p

		// 缓存帖子详情
		data, _ := json.Marshal(post)
		s.Redis.Set(key, data, time.Hour)
	}

	// 获取评论列表（带分页）
	comments, err := dao.GetPostCommentsWithCursor(s.DB, postID, cursorTime, cursorID, pageSize)
	if err != nil {
		return &post, nil, false, nil, err
	}

	// 判断是否还有更多评论
	hasMore := len(comments) == pageSize

	// 构建下一个游标
	var nextCursor interface{} = nil
	if hasMore && len(comments) > 0 {
		last := comments[len(comments)-1]
		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	return &post, comments, hasMore, nextCursor, nil
}

// GetPostDetailOnly 只获取帖子详情（不包含评论）
// func (s *ArticlePostServiceImpl) GetPostDetailOnly(postID int64) (model.Post, error) {
// 	key := fmt.Sprintf(cache.PostDetail, postID)

// 	val, _ := s.Redis.Get(key).Result()
// 	if val != "" {
// 		var post model.Post
// 		json.Unmarshal([]byte(val), &post)
// 		return post, nil
// 	}

// 	post, err := dao.GetPostDetail(s.DB, postID)
// 	if err != nil {
// 		return post, err
// 	}

// 	data, _ := json.Marshal(post)
// 	s.Redis.Set(key, data, time.Hour)

// 	return post, nil
// }

// 创建评论
func (s *ArticlePostServiceImpl) CreateComment(userID, postID, parentID int64, content string, images []model.CommentImage) error {
	var postComment = &model.PostComment{
		UserID:   userID,
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
		Images:   images,
	}
	err := dao.CreateComment(s.DB, postComment)

	if err == nil {

		key := fmt.Sprintf(cache.PostComment, postID)

		s.Redis.Del(key)

		key2 := fmt.Sprintf(cache.PostDetail, postID)

		s.Redis.Del(key2)
	}

	return err
}

// 创建帖子
func (s *ArticlePostServiceImpl) CreatePost(userID int64, content string, title string, images []string) error {
	post := model.Post{
		UserID:     userID,
		Content:    content,
		Title:      title,
		ImageCount: len(images),
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&post).Error; err != nil {
			return err
		}

		var imageList []model.PostImage

		for i, url := range images {
			imageList = append(imageList, model.PostImage{
				PostID:   post.ID,
				ImageURL: url,
				Sort:     i,
			})
		}

		if len(imageList) > 0 {
			if err := tx.Create(&imageList).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *ArticlePostServiceImpl) SearchPosts(req model.SearchPostsModel) ([]model.Post, error) {
	// 参数校验
	if req.Keyword == "" {
		return []model.Post{}, nil
	}

	// 设置默认搜索类型
	if req.SearchType == "" {
		req.SearchType = "all"
	}

	// 如果是收藏搜索但没有用户ID，返回空
	if req.SearchType == "favorite" && req.UserID == 0 {
		return []model.Post{}, nil
	}

	// 调用DAO层
	posts, err := dao.SearchPostsDAO(
		s.DB,
		req.SearchType,
		req.Keyword,
		req.UserID,
		req.CursorTime,
		req.CursorID,
		req.PageSize,
	)

	if err != nil {
		return nil, err
	}

	// 如果是收藏搜索，标记IsFavorite字段
	if req.SearchType == "favorite" && len(posts) > 0 {
		for i := range posts {
			posts[i].IsFavorite = true
		}
	}

	return posts, nil
}

// DeleteComment 删除评论
func (s *ArticlePostServiceImpl) DeleteComment(userID int64, commentID int64, isAdmin bool) error {
	// 获取评论信息
	comment, err := dao.GetCommentByID(s.DB, commentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("评论不存在")
		}
		return err
	}

	// 权限校验：只有评论作者或管理员可以删除
	if comment.UserID != userID && !isAdmin {
		return errors.New("没有权限删除此评论")
	}

	// 删除评论
	err = dao.DeleteCommentDAO(s.DB, commentID)
	if err != nil {
		return err
	}

	// 清除相关缓存
	key := fmt.Sprintf(cache.PostComment, comment.PostID)
	s.Redis.Del(key)
	key2 := fmt.Sprintf(cache.PostDetail, comment.PostID)
	s.Redis.Del(key2)

	return nil
}

// GetUserPosts 获取用户的所有帖子
func (s *ArticlePostServiceImpl) GetUserPosts(userID int64, targetUserID int64, cursorTime time.Time,
	cursorID int64, pageSize int) ([]model.Post, error) {

	// 获取目标用户的帖子列表
	posts, err := dao.GetUserPosts(s.DB, targetUserID, cursorTime, cursorID, pageSize)
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 获取当前登录用户的收藏状态
	var ids []int64
	for _, p := range posts {
		ids = append(ids, p.ID)
	}

	favMap := dao.GetUserFavoritePostIDs(s.DB, userID, ids)

	for i := range posts {
		if favMap[posts[i].ID] {
			posts[i].IsFavorite = true
		}
	}

	return posts, nil
}
