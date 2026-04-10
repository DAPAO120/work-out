package model

type Result struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data LoginData `json:"data"`
}

type LoginData struct {
	Token   string `json:"token"`
	Openid  string `json:"openId"`
	UserId  uint   `json:"userId"`
	IsAdmin bool   `json:"isAdmin"`
}

type UserProfileResp struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`

	FollowerCount  int64 `json:"follower_count"`  // 粉丝数
	FollowingCount int64 `json:"following_count"` // 关注数
	PostCount      int64 `json:"post_count"`      // 动态数

	IsFollow bool `json:"is_follow"` // 当前用户是否已关注
}

type UserMeResp struct {
	ID             uint   `json:"id"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Bio            string `json:"bio"`
	Gender         string `json:"gender"`
	FollowerCount  int64  `json:"follower_count"`  // 粉丝数（关注我的）
	FollowingCount int64  `json:"following_count"` // 我关注的
	PostCount      int64  `json:"post_count"`      // 我的动态
}

// PostDetailResponse 帖子详情响应
type PostDetailResponse struct {
	Post     Post          `json:"post"`
	Comments []PostComment `json:"comments"`
	HasMore  bool          `json:"has_more"`
	Cursor   interface{}   `json:"cursor"`
	Total    int           `json:"total"`
}
type UserProfileResponse struct {
	ID          int64  `json:"id"`
	OpenID      string `json:"open_id"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Background  string `json:"background"`
	Gender      int8   `json:"gender"`
	IsAdmin     bool   `json:"is_admin"`
	FollowCount int64  `json:"follow_count"` // 关注数
	FansCount   int64  `json:"fans_count"`   // 粉丝数
	PostCount   int64  `json:"post_count"`   // 帖子数
}
type RankUserResponse struct {
	UserID        int64  `json:"user_id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	PostCount     int64  `json:"post_count"`
	FeaturedCount int64  `json:"featured_count"`
	Rank          int64  `json:"rank"`
	Score         int64  `json:"score"`
}
