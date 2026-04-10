package model

type UpdateProfileRequest struct {
	Nickname string `json:"input"`
	Avatar   string `json:"upload"`
	Gender   string `json:"sct"`
	Bio      string `json:"textarea"`
}

// SearchPostsRequest 统一搜索请求参数
type SearchPostsRequest struct {
	Keyword    string `form:"keyword" binding:"required,min=1,max=100"`                    // 搜索关键词
	SearchType string `form:"search_type" binding:"omitempty,oneof=all featured favorite"` // 搜索类型：all(全部), featured(精品), favorite(收藏)
	CursorTime string `form:"cursor_time"`                                                 // 游标时间
	CursorID   string `form:"cursor_id"`                                                   // 游标ID
	PageSize   string `form:"pageSize"`                                                    // 每页数量
}

// GetPostDetailRequest 获取帖子详情请求参数
type GetPostDetailRequest struct {
	PostID     int64  `form:"post_id" binding:"required"`
	CursorTime string `form:"cursor_time"`
	CursorID   string `form:"cursor_id"`
	PageSize   string `form:"pageSize"`
}
