package model

type RankItem struct {
	UserID        int64  `json:"user_id"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	PostCount     int64  `json:"post_count"`     // 总帖子数
	FeaturedCount int64  `json:"featured_count"` // 精品帖子数
	Score         int64  `json:"score"`          // 综合分数（可自定义权重）
}
