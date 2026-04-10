package impl

import (
	"Project001/global"
	"Project001/internal/model"
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"image"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

type WxProfileServiceImpl struct{}

func (s *WxProfileServiceImpl) GetUserProfile(currentUserID, targetUserID int64) (*model.UserProfileResp, error) {

	var user model.User
	if err := global.DB.First(&user, targetUserID).Error; err != nil {
		return nil, err
	}

	var followerCount int64
	global.DB.Model(&model.UserFollow{}).
		Where("follow_user_id = ?", targetUserID).
		Count(&followerCount)

	var followingCount int64
	global.DB.Model(&model.UserFollow{}).
		Where("user_id = ?", targetUserID).
		Count(&followingCount)

	var postCount int64
	global.DB.Model(&model.Post{}).
		Where("user_id = ?", targetUserID).
		Count(&postCount)

	var isFollow bool
	if currentUserID != 0 {
		var count int64
		global.DB.Model(&model.UserFollow{}).
			Where("user_id = ? AND follow_user_id = ?", currentUserID, targetUserID).
			Count(&count)

		isFollow = count > 0
	}

	return &model.UserProfileResp{
		ID:             user.ID,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		PostCount:      postCount,
		IsFollow:       isFollow,
	}, nil
}

func (s *WxProfileServiceImpl) GetMyProfile(userID int64) (*model.UserMeResp, error) {

	var user model.User
	if err := global.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	// 粉丝数（谁关注我）
	var followerCount int64
	global.DB.Model(&model.UserFollow{}).
		Where("follow_user_id = ?", userID).
		Count(&followerCount)

	// 关注数（我关注谁）
	var followingCount int64
	global.DB.Model(&model.UserFollow{}).
		Where("user_id = ?", userID).
		Count(&followingCount)

	// 我的动态数
	var postCount int64
	global.DB.Model(&model.Post{}).
		Where("user_id = ?", userID).
		Count(&postCount)

	return &model.UserMeResp{
		ID:             user.ID,
		Nickname:       user.Nickname,
		Avatar:         user.Avatar,
		Bio:            user.Bio,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		PostCount:      postCount,
		Gender:         strconv.Itoa(int(user.Gender)),
	}, nil
}

// 上传头像
func (s *WxProfileServiceImpl) UploadImage(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"msg": "file error"})
		return
	}

	// 限制大小（例如 4MB）
	if file.Size > 4*1024*1024 {
		c.JSON(400, gin.H{"msg": "file too large"})
		return
	}

	// 获取后缀,防止路径攻击
	ext := strings.ToLower(path.Ext(file.Filename))

	//限制文件类型
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	if !allowed[ext] {
		c.JSON(400, gin.H{"msg": "invalid file type"})
		return
	}

	filename := fmt.Sprintf("%d_%d%s", time.Now().Unix(), rand.Intn(1000), ext)

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"msg": "open file failed"})
		return
	}
	defer src.Close()

	//压缩图片大小
	//解码图片
	img, _, err := image.Decode(src)
	if err != nil {
		c.JSON(400, gin.H{"msg": "invalid image"})
		return
	}

	// 调整大小（最大 512px）
	img = imaging.Resize(img, 512, 0, imaging.Lanczos)

	// 压缩输出
	buf := new(bytes.Buffer)

	switch ext {
	case ".jpg", ".jpeg":
		err = imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(80))
	case ".png":
		err = imaging.Encode(buf, img, imaging.PNG)
	default:
		c.JSON(400, gin.H{"msg": "unsupported image"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"msg": "image compress failed"})
		return
	}
	//转换格式
	// imaging.Encode(buf, img, imaging.JPEG, imaging.JPEGQuality(80))
	fileBytes := buf.Bytes()

	// 上传到Nginx（WebDAV PUT）
	uploadURL := fmt.Sprintf(
		"%s/upload/work-out/avatar/%s",
		global.Config.Server.Domain,
		filename,
	)

	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(fileBytes))
	if err != nil {
		c.JSON(500, gin.H{"msg": "create request failed"})
		return
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"msg": "upload to nginx failed"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		c.JSON(500, gin.H{
			"msg": fmt.Sprintf("nginx upload failed: %d", resp.StatusCode),
		})
		return
	}

	// 返回访问URL
	fileURL := fmt.Sprintf(
		"%s/work-out/avatar/%s",
		global.Config.Server.Domain,
		filename,
	)

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"url": fileURL,
		},
	})
}

// 更新个人信息
func (s *WxProfileServiceImpl) UpdateUserProfile(c *gin.Context) {

	var req model.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"msg": "参数错误",
		})
		return
	}

	userID := c.GetInt64("user_id")

	var user model.User

	if err := global.DB.First(&user, userID).Error; err != nil {
		c.JSON(404, gin.H{
			"msg": "用户不存在",
		})
		return
	}
	genderInt, err := strconv.ParseInt(req.Gender, 10, 8)
	if err != nil {
		fmt.Println("性别转换错误:", err)
		return
	}
	user.Nickname = req.Nickname
	user.Avatar = req.Avatar
	user.Gender = int8(genderInt)
	user.Bio = req.Bio
	if err := global.DB.Save(&user).Error; err != nil {
		c.JSON(500, gin.H{
			"msg": "更新失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "成功",
		"data": user,
	})
}
