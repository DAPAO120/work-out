package dao

import (
	"Project001/internal/model"
	"time"

	"gorm.io/gorm"
)

func CreatePost(db *gorm.DB, post *model.Post) error {

	return db.Create(post).Error
}

func CreatePostImages(db *gorm.DB, images []model.PostImage) error {

	return db.Create(&images).Error
}

func GetPostDetail(db *gorm.DB, postID int64) (model.Post, error) {

	var post model.Post

	err := db.Preload("Images").
		Preload("User").
		First(&post, postID).Error

	if err != nil {
		return post, err
	}

	// 获取真实收藏数和评论数
	var favoriteCount int64
	db.Model(&model.PostFavorite{}).Where("post_id = ?", postID).Count(&favoriteCount)
	post.FavoriteCount = int(favoriteCount)

	var commentCount int64
	db.Model(&model.PostComment{}).Where("post_id = ? AND deleted_at IS NULL", postID).Count(&commentCount)
	post.CommentCount = int(commentCount)

	return post, nil
}

func GetPostList(
	db *gorm.DB,
	cursorTime time.Time,
	cursorID int64,
	pageSize int,
) ([]model.Post, error) {

	var posts []model.Post

	query := db.Preload("Images").
		Preload("User").
		Order("created_at DESC,id DESC").
		Limit(pageSize)

	if !cursorTime.IsZero() {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 批量获取所有帖子的收藏数和评论数
	postIDs := make([]int64, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询收藏数
	type CountResult struct {
		PostID int64
		Count  int
	}
	var favoriteCounts []CountResult
	db.Model(&model.PostFavorite{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Scan(&favoriteCounts)

	favCountMap := make(map[int64]int)
	for _, fc := range favoriteCounts {
		favCountMap[fc.PostID] = fc.Count
	}

	// 批量查询评论数
	var commentCounts []CountResult
	db.Model(&model.PostComment{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&commentCounts)

	commentCountMap := make(map[int64]int)
	for _, cc := range commentCounts {
		commentCountMap[cc.PostID] = cc.Count
	}

	// 赋值
	for i := range posts {
		if cnt, ok := favCountMap[posts[i].ID]; ok {
			posts[i].FavoriteCount = cnt
		}
		if cnt, ok := commentCountMap[posts[i].ID]; ok {
			posts[i].CommentCount = cnt
		}
	}

	return posts, nil
}

func GetFeaturedPosts(db *gorm.DB,
	cursorTime time.Time,
	cursorID int64,
	pageSize int) ([]model.Post, error) {

	var posts []model.Post

	query := db.Preload("Images").
		Preload("User").
		Where("is_featured = ?", true).
		Order("created_at DESC").
		Limit(pageSize)

	if !cursorTime.IsZero() {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}
	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 批量获取所有帖子的收藏数和评论数
	postIDs := make([]int64, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询收藏数
	type CountResult struct {
		PostID int64
		Count  int
	}
	var favoriteCounts []CountResult
	db.Model(&model.PostFavorite{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Scan(&favoriteCounts)

	favCountMap := make(map[int64]int)
	for _, fc := range favoriteCounts {
		favCountMap[fc.PostID] = fc.Count
	}

	// 批量查询评论数
	var commentCounts []CountResult
	db.Model(&model.PostComment{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&commentCounts)

	commentCountMap := make(map[int64]int)
	for _, cc := range commentCounts {
		commentCountMap[cc.PostID] = cc.Count
	}

	// 赋值
	for i := range posts {
		if cnt, ok := favCountMap[posts[i].ID]; ok {
			posts[i].FavoriteCount = cnt
		}
		if cnt, ok := commentCountMap[posts[i].ID]; ok {
			posts[i].CommentCount = cnt
		}
	}

	return posts, nil
}

func DeletePost(db *gorm.DB, postID int64) error {

	return db.Delete(&model.Post{}, postID).Error
}

func SetFeatured(db *gorm.DB, postID int64, featured bool) error {

	return db.Model(&model.Post{}).
		Where("id=?", postID).
		Update("is_featured", featured).Error
}

func CreateFavorite(db *gorm.DB, userID, postID int64) error {

	f := model.PostFavorite{
		UserID: userID,
		PostID: postID,
	}

	return db.Create(&f).Error
}

func DeleteFavorite(db *gorm.DB, userID, postID int64) error {

	return db.Where("user_id=? AND post_id=?", userID, postID).
		Delete(&model.PostFavorite{}).Error
}

func GetUserFavoritePostIDs(db *gorm.DB, userID int64, postIDs []int64) map[int64]bool {

	var list []model.PostFavorite

	db.Where("user_id=? AND post_id IN ?", userID, postIDs).
		Find(&list)

	result := map[int64]bool{}

	for _, v := range list {
		result[v.PostID] = true
	}

	return result
}

func GetFavoritePosts(db *gorm.DB, userID int64, cursorTime time.Time,
	cursorID int64,
	pageSize int) ([]model.Post, error) {

	var posts []model.Post

	query := db.Table("user_post p").
		Preload("User").
		Select("p.*").
		Joins("JOIN post_favorite f ON p.id=f.post_id").
		Where("f.user_id=?", userID).
		Order("p.created_at DESC, p.id DESC").
		Limit(pageSize)

	if !cursorTime.IsZero() {
		query = query.Where(
			"(p.created_at < ?) OR (p.created_at = ? AND p.id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 批量获取所有帖子的收藏数和评论数
	postIDs := make([]int64, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询收藏数
	type CountResult struct {
		PostID int64
		Count  int
	}
	var favoriteCounts []CountResult
	db.Model(&model.PostFavorite{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Scan(&favoriteCounts)

	favCountMap := make(map[int64]int)
	for _, fc := range favoriteCounts {
		favCountMap[fc.PostID] = fc.Count
	}

	// 批量查询评论数
	var commentCounts []CountResult
	db.Model(&model.PostComment{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&commentCounts)

	commentCountMap := make(map[int64]int)
	for _, cc := range commentCounts {
		commentCountMap[cc.PostID] = cc.Count
	}

	// 赋值
	for i := range posts {
		if cnt, ok := favCountMap[posts[i].ID]; ok {
			posts[i].FavoriteCount = cnt
		}
		if cnt, ok := commentCountMap[posts[i].ID]; ok {
			posts[i].CommentCount = cnt
		}
	}

	return posts, nil
}

func CreateComment(db *gorm.DB, c *model.PostComment) error {

	return db.Create(c).Error
}

func CreateCommentImages(db *gorm.DB, imgs []model.CommentImage) error {

	return db.Create(&imgs).Error
}

func GetPostComments(db *gorm.DB, postID int64, lastID int64) ([]model.PostComment, error) {

	var list []model.PostComment

	query := db.Preload("User").
		Preload("Images").
		Where("post_id=?", postID)

	if lastID > 0 {
		query = query.Where("id < ?", lastID)
	}

	err := query.Order("id DESC").
		Limit(20).
		Find(&list).Error

	return list, err
}

func DeleteComment(db *gorm.DB, id int64) error {

	return db.Delete(&model.PostComment{}, id).Error
}

// 管理员设置精品
func SetPostFeatured(db *gorm.DB, postID int64) error {

	return db.Model(&model.Post{}).
		Where("id=?", postID).
		Update("is_featured", true).Error
}

// 取消精品
func CancelFeatured(db *gorm.DB, postID int64) error {

	return db.Model(&model.Post{}).
		Where("id=?", postID).
		Update("is_featured", false).Error
}

// SearchPostsDAO 通用搜索DAO
// searchType: "all", "featured", "favorite"
// keyword: 搜索关键词
// userID: 用户ID（searchType为favorite时必填）
// cursorTime, cursorID: 游标分页参数
// pageSize: 每页数量
func SearchPostsDAO(
	db *gorm.DB,
	searchType string,
	keyword string,
	userID int64,
	cursorTime time.Time,
	cursorID int64,
	pageSize int,
) ([]model.Post, error) {

	var posts []model.Post

	// 构建基础查询
	query := db.Model(&model.Post{}).
		Preload("Images").
		Preload("User").
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%").
		Order("created_at DESC, id DESC").
		Limit(pageSize)

	// 根据搜索类型添加过滤条件
	switch searchType {
	case "featured":
		query = query.Where("is_featured = ?", true)
	case "favorite":
		if userID == 0 {
			return posts, nil
		}
		// 使用子查询方式，只返回用户收藏的帖子
		query = query.Where("id IN (?)",
			db.Table("post_favorite").
				Select("post_id").
				Where("user_id = ?", userID))
	}

	// 添加游标分页条件
	if !cursorTime.IsZero() {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 批量获取所有帖子的收藏数和评论数
	postIDs := make([]int64, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询收藏数
	type CountResult struct {
		PostID int64
		Count  int
	}
	var favoriteCounts []CountResult
	db.Model(&model.PostFavorite{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Scan(&favoriteCounts)

	favCountMap := make(map[int64]int)
	for _, fc := range favoriteCounts {
		favCountMap[fc.PostID] = fc.Count
	}

	// 批量查询评论数
	var commentCounts []CountResult
	db.Model(&model.PostComment{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&commentCounts)

	commentCountMap := make(map[int64]int)
	for _, cc := range commentCounts {
		commentCountMap[cc.PostID] = cc.Count
	}

	// 赋值
	for i := range posts {
		if cnt, ok := favCountMap[posts[i].ID]; ok {
			posts[i].FavoriteCount = cnt
		}
		if cnt, ok := commentCountMap[posts[i].ID]; ok {
			posts[i].CommentCount = cnt
		}
	}

	return posts, nil
}

// GetPostCommentsWithCursor 使用游标分页获取帖子评论
func GetPostCommentsWithCursor(
	db *gorm.DB,
	postID int64,
	cursorTime time.Time,
	cursorID int64,
	pageSize int,
) ([]model.PostComment, error) {

	var comments []model.PostComment

	query := db.Preload("User").
		Preload("Images").
		Where("post_id = ?", postID).
		Where("deleted_at IS NULL"). // 过滤软删除的评论
		Order("created_at DESC, id DESC").
		Limit(pageSize)

	// 添加游标分页条件
	if !cursorTime.IsZero() {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}

	err := query.Find(&comments).Error
	return comments, err
}

// DeleteCommentDAO 删除评论（软删除）
func DeleteCommentDAO(db *gorm.DB, commentID int64) error {
	return db.Where("id = ?", commentID).Delete(&model.PostComment{}).Error
}

// GetCommentByID 根据ID获取评论信息
func GetCommentByID(db *gorm.DB, commentID int64) (model.PostComment, error) {
	var comment model.PostComment
	err := db.First(&comment, commentID).Error
	return comment, err
}

// GetUserPosts 获取用户的所有帖子（滚动分页）
func GetUserPosts(
	db *gorm.DB,
	userID int64,
	cursorTime time.Time,
	cursorID int64,
	pageSize int,
) ([]model.Post, error) {
	var posts []model.Post

	query := db.Preload("Images").
		Preload("User").
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Limit(pageSize)

	if !cursorTime.IsZero() {
		query = query.Where(
			"(created_at < ?) OR (created_at = ? AND id < ?)",
			cursorTime,
			cursorTime,
			cursorID,
		)
	}

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	if len(posts) == 0 {
		return posts, nil
	}

	// 批量获取所有帖子的收藏数和评论数
	postIDs := make([]int64, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	// 批量查询收藏数
	type CountResult struct {
		PostID int64
		Count  int
	}
	var favoriteCounts []CountResult
	db.Model(&model.PostFavorite{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Scan(&favoriteCounts)

	favCountMap := make(map[int64]int)
	for _, fc := range favoriteCounts {
		favCountMap[fc.PostID] = fc.Count
	}

	// 批量查询评论数
	var commentCounts []CountResult
	db.Model(&model.PostComment{}).
		Select("post_id, count(*) as count").
		Where("post_id IN ? AND deleted_at IS NULL", postIDs).
		Group("post_id").
		Scan(&commentCounts)

	commentCountMap := make(map[int64]int)
	for _, cc := range commentCounts {
		commentCountMap[cc.PostID] = cc.Count
	}

	// 赋值
	for i := range posts {
		if cnt, ok := favCountMap[posts[i].ID]; ok {
			posts[i].FavoriteCount = cnt
		}
		if cnt, ok := commentCountMap[posts[i].ID]; ok {
			posts[i].CommentCount = cnt
		}
	}

	return posts, nil
}
