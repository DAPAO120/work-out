package controller

import (
	"Project001/internal/model"
	"Project001/internal/service/article"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ArticlePostController struct {
	articlePostService article.IArticlePostService
}

func NewArticleController(articlePostService article.IArticlePostService) *ArticlePostController {
	return &ArticlePostController{
		articlePostService: articlePostService,
	}
}

// 创建帖子
func (ctl *ArticlePostController) CreatePost(c *gin.Context) {

	userID := c.GetInt64("user_id")

	var req struct {
		Title   string `form:"textarea1"`
		Content string `form:"textarea"`
		Upload  string `form:"upload"` // 先接收字符串
	}

	if err := c.ShouldBind(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "参数解析失败",
		})
		return
	}

	// title 非空校验
	if strings.TrimSpace(req.Title) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "title不能为空",
		})
		return
	}

	// content 非空校验
	if strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "content不能为空",
		})
		return
	}

	// 解析图片URL
	var images []string
	if req.Upload != "" {

		arr := strings.Split(req.Upload, ",")

		for _, url := range arr {

			url = strings.TrimSpace(url)

			if url != "" {
				images = append(images, url)
			}
		}
	}

	err := ctl.articlePostService.CreatePost(
		userID,
		req.Content,
		req.Title,
		images,
	)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"msg":  "创建成功",
	})
}

// 获取帖子列表（滚动分页 + 收藏状态）
func (ctl *ArticlePostController) GetPostList(c *gin.Context) {

	userID := c.GetInt64("user_id")

	cursorTimeStr := c.Query("cursor_time")
	cursorIDStr := c.Query("cursor_id")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	pageSize, _ := strconv.Atoi(pageSizeStr)

	if pageSize > 50 {
		pageSize = 20
	}

	var cursorTime time.Time
	var cursorID int64

	if cursorTimeStr != "" {
		t, err := time.Parse(time.RFC3339, cursorTimeStr)
		if err == nil {
			cursorTime = t
		}
	}

	cursorID, _ = strconv.ParseInt(cursorIDStr, 10, 64)

	posts, err := ctl.articlePostService.GetPostList(
		userID,
		cursorTime,
		cursorID,
		pageSize,
	)

	if err != nil {

		c.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	hasMore := len(posts) == pageSize

	var nextCursor interface{} = nil

	if hasMore {

		last := posts[len(posts)-1]

		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	c.JSON(200, gin.H{
		"code":     200,
		"msg":      "success",
		"data":     posts,
		"cursor":   nextCursor,
		"has_more": strconv.FormatBool(hasMore),
	})
}

// 获取收藏列表
func (ctl *ArticlePostController) GetFavoritePosts(c *gin.Context) {

	userID := c.GetInt64("user_id")

	cursorTimeStr := c.Query("cursor_time")
	cursorIDStr := c.Query("cursor_id")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	pageSize, _ := strconv.Atoi(pageSizeStr)

	if pageSize > 50 {
		pageSize = 20
	}

	var cursorTime time.Time
	var cursorID int64

	if cursorTimeStr != "" {
		t, err := time.Parse(time.RFC3339, cursorTimeStr)
		if err == nil {
			cursorTime = t
		}
	}

	cursorID, _ = strconv.ParseInt(cursorIDStr, 10, 64)

	posts, err := ctl.articlePostService.GetFavoritePosts(userID,
		cursorTime,
		cursorID,
		pageSize)

	if err != nil {

		c.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	hasMore := len(posts) == pageSize

	var nextCursor interface{} = nil

	if hasMore {

		last := posts[len(posts)-1]

		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	c.JSON(200, gin.H{
		"code":     200,
		"msg":      "success",
		"data":     posts,
		"cursor":   nextCursor,
		"has_more": strconv.FormatBool(hasMore),
	})
}

// 收藏帖子
func (ctl *ArticlePostController) FavoritePost(c *gin.Context) {

	userID := c.GetInt64("user_id")

	var req struct {
		PostID int64 `json:"post_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctl.articlePostService.FavoritePost(userID, req.PostID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

// 取消收藏
func (ctl *ArticlePostController) CancelFavorite(c *gin.Context) {

	userID := c.GetInt64("user_id")

	var req struct {
		PostID int64 `json:"post_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ctl.articlePostService.CancelFavorite(userID, req.PostID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

// 删除帖子
func (ctl *ArticlePostController) DeletePost(c *gin.Context) {

	userID := c.GetInt64("user_id")

	postIDStr := c.Query("id")

	postID, _ := strconv.ParseInt(postIDStr, 10, 64)

	isAdmin := c.GetBool("is_admin")

	err := ctl.articlePostService.DeletePost(userID, postID, isAdmin)

	if err != nil {

		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

// 设置精品
func (ctl *ArticlePostController) SetFeatured(c *gin.Context) {
	isAdmin := c.GetBool("is_admin")
	if !isAdmin {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "没有权限"})
		return
	}

	postIDStr := c.Query("id")

	postID, _ := strconv.ParseInt(postIDStr, 10, 64)

	err := ctl.articlePostService.SetPostFeatured(postID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

// 取消精品
func (ctl *ArticlePostController) CancelFeatured(c *gin.Context) {
	isAdmin := c.GetBool("is_admin")
	if !isAdmin {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "没有权限"})
		return
	}

	postIDStr := c.Query("id")

	postID, _ := strconv.ParseInt(postIDStr, 10, 64)

	err := ctl.articlePostService.CancelFeatured(postID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

// 精品帖子列表
func (ctl *ArticlePostController) GetFeaturedPosts(c *gin.Context) {

	cursorTimeStr := c.Query("cursor_time")
	cursorIDStr := c.Query("cursor_id")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	pageSize, _ := strconv.Atoi(pageSizeStr)

	if pageSize > 50 {
		pageSize = 20
	}

	var cursorTime time.Time
	var cursorID int64

	if cursorTimeStr != "" {
		t, err := time.Parse(time.RFC3339, cursorTimeStr)
		if err == nil {
			cursorTime = t
		}
	}

	cursorID, _ = strconv.ParseInt(cursorIDStr, 10, 64)

	posts, err := ctl.articlePostService.GetFeaturedPosts(
		cursorTime,
		cursorID,
		pageSize)

	if err != nil {

		c.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	hasMore := len(posts) == pageSize

	var nextCursor interface{} = nil

	if hasMore {

		last := posts[len(posts)-1]

		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	c.JSON(200, gin.H{
		"code":     200,
		"msg":      "success",
		"data":     posts,
		"cursor":   nextCursor,
		"has_more": strconv.FormatBool(hasMore),
	})
}

// 获取帖子详情（包含评论分页）
func (ctl *ArticlePostController) GetPostDetail(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// 获取帖子ID
	postIDStr := c.Query("id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "帖子ID参数错误",
		})
		return
	}

	// 获取分页参数
	cursorTimeStr := c.Query("cursor_time")
	cursorIDStr := c.Query("cursor_id")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	// 解析每页数量
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	// 解析游标参数
	var cursorTime time.Time
	var cursorID int64

	if cursorTimeStr != "" {
		// 尝试多种时间格式
		if t, err := time.Parse(time.RFC3339, cursorTimeStr); err == nil {
			cursorTime = t
		} else if t, err := time.Parse("2006-01-02 15:04:05", cursorTimeStr); err == nil {
			cursorTime = t
		}
	}

	if cursorIDStr != "" {
		cursorID, _ = strconv.ParseInt(cursorIDStr, 10, 64)
	}

	// 调用服务层获取帖子详情和评论
	post, comments, hasMore, nextCursor, err := ctl.articlePostService.GetPostDetail(
		userID,
		postID,
		cursorTime,
		cursorID,
		pageSize,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"msg":      "success",
		"data":     post,
		"comments": comments,
		"cursor":   nextCursor,
		"has_more": hasMore,
		"total":    len(comments),
	})
}

// 创建评论
func (ctl *ArticlePostController) CreateComment(c *gin.Context) {

	userID := c.GetInt64("user_id")

	// 定义表单接收参数
	var req struct {
		PostID   int64  `form:"postId" binding:"required"`
		ParentID int64  `form:"parentId"`
		Content  string `form:"textarea" binding:"required"`
		Images   string `form:"upload"` // 逗号分割的图片URL字符串
	}

	// 绑定表单参数
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误：" + err.Error(),
		})
		return
	}

	// 校验内容非空
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "评论内容不能为空",
		})
		return
	}

	// 解析图片URL字符串为数组
	var imageList []model.CommentImage
	if req.Images != "" {
		// 按逗号分割
		imageURLs := strings.Split(req.Images, ",")
		for _, url := range imageURLs {
			url = strings.TrimSpace(url)
			if url != "" {
				imageList = append(imageList, model.CommentImage{
					ImageURL: url,
				})
			}
		}
	}

	// 调用服务层
	err := ctl.articlePostService.CreateComment(
		userID,
		req.PostID,
		req.ParentID,
		req.Content,
		imageList,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "评论成功",
	})
}

// SearchPosts 统一搜索接口
func (ctl *ArticlePostController) SearchPosts(c *gin.Context) {
	var req model.SearchPostsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误：" + err.Error(),
		})
		return
	}

	// 设置默认搜索类型
	if req.SearchType == "" {
		req.SearchType = "all"
	}

	// 获取用户ID（从认证中间件获取）
	userID := c.GetInt64("user_id")

	// 如果是搜索收藏帖子但用户未登录
	if req.SearchType == "favorite" && userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  "请先登录",
		})
		return
	}

	// 解析分页参数
	pageSize := 10
	if req.PageSize != "" {
		if size, err := strconv.Atoi(req.PageSize); err == nil && size > 0 {
			pageSize = size
		}
	}
	// 限制最大每页数量
	if pageSize > 50 {
		pageSize = 50
	}

	// 解析游标参数
	var cursorTime time.Time
	var cursorID int64

	if req.CursorTime != "" {
		if t, err := time.Parse(time.RFC3339, req.CursorTime); err == nil {
			cursorTime = t
		} else {
			// 尝试其他时间格式
			if t, err := time.Parse("2006-01-02 15:04:05", req.CursorTime); err == nil {
				cursorTime = t
			}
		}
	}

	if req.CursorID != "" {
		cursorID, _ = strconv.ParseInt(req.CursorID, 10, 64)
	}

	// 构建服务层请求
	searchReq := model.SearchPostsModel{
		SearchType: req.SearchType,
		Keyword:    req.Keyword,
		UserID:     userID,
		CursorTime: cursorTime,
		CursorID:   cursorID,
		PageSize:   pageSize,
	}

	// 调用服务层
	posts, err := ctl.articlePostService.SearchPosts(searchReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "搜索失败：" + err.Error(),
		})
		return
	}

	// 判断是否还有更多数据
	hasMore := len(posts) == pageSize

	// 构建下一个游标
	var nextCursor interface{} = nil
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code":        200,
		"msg":         "success",
		"data":        posts,
		"cursor":      nextCursor,
		"has_more":    strconv.FormatBool(hasMore),
		"keyword":     req.Keyword,
		"search_type": req.SearchType,
		"total":       len(posts),
	})
}

// DeleteComment 删除评论
func (ctl *ArticlePostController) DeleteComment(c *gin.Context) {
	userID := c.GetInt64("user_id")
	isAdmin := c.GetBool("is_admin")

	// 获取评论ID
	commentIDStr := c.Query("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "评论ID参数错误",
		})
		return
	}

	err = ctl.articlePostService.DeleteComment(userID, commentID, isAdmin)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"code": 403,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

// GetUserPosts 获取用户的所有帖子列表
func (ctl *ArticlePostController) GetUserPosts(c *gin.Context) {
	// 当前登录用户ID
	currentUserID := c.GetInt64("user_id")

	// 目标用户ID（从路径参数获取）
	targetUserIDStr := c.Query("user_id")
	targetUserID, err := strconv.ParseInt(targetUserIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "用户ID参数错误",
		})
		return
	}

	// 分页参数
	cursorTimeStr := c.Query("cursor_time")
	cursorIDStr := c.Query("cursor_id")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize > 50 {
		pageSize = 20
	}

	var cursorTime time.Time
	var cursorID int64

	if cursorTimeStr != "" {
		t, err := time.Parse(time.RFC3339, cursorTimeStr)
		if err == nil {
			cursorTime = t
		}
	}

	cursorID, _ = strconv.ParseInt(cursorIDStr, 10, 64)

	// 获取用户帖子列表
	posts, err := ctl.articlePostService.GetUserPosts(
		currentUserID,
		targetUserID,
		cursorTime,
		cursorID,
		pageSize,
	)

	if err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	hasMore := len(posts) == pageSize

	var nextCursor interface{} = nil
	if hasMore && len(posts) > 0 {
		last := posts[len(posts)-1]
		nextCursor = gin.H{
			"cursor_time": last.CreatedAt.Format(time.RFC3339),
			"cursor_id":   last.ID,
		}
	}

	c.JSON(200, gin.H{
		"code":     200,
		"msg":      "success",
		"data":     posts,
		"cursor":   nextCursor,
		"has_more": strconv.FormatBool(hasMore),
	})
}
