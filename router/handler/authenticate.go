package handler

import (
	"fmt"
	"github.com/SasukeBo/pmes-data-center/cache"
	"github.com/SasukeBo/pmes-data-center/errormap"
	"github.com/SasukeBo/pmes-data-center/orm"
	"github.com/SasukeBo/pmes-data-center/util"
	"github.com/SasukeBo/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"net/http"
	"strconv"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		isGraphiQL := c.GetBool("isGraphiQL")
		if isGraphiQL {
			c.Next()
			return
		}

		sessionID, err := c.Cookie("access_token")
		if err != nil {
			errormap.SendHttpError(
				c, errormap.ErrorCodeUnauthenticated,
				errormap.NewOrigin("Get sessionID from cookie failed: %v", err),
			)
			return
		}

		var user orm.User
		userUUID, err := cache.GetString(sessionID)
		if err != nil { // 内存中未命中，前往db获取
			log.Info("User not found in cache")

			var userLogin orm.UserLogin
			err := orm.DB.Model(&orm.UserLogin{}).Where("access_token = ?", sessionID).First(&userLogin).Error
			if err != nil { // 获取Login记录失败
				errormap.SendHttpError(
					c, errormap.ErrorCodeUnauthenticated,
					errormap.NewOrigin("Get user_login with token=%v failed: %v", sessionID, err),
				)
				return
			}

			if !userLogin.KeepLogin {
				errormap.SendHttpError(
					c, errormap.ErrorCodeUnauthenticated,
					errormap.NewOrigin("User not keep login status."),
				)
				return
			}

			err = orm.DB.Model(&orm.User{}).Where("id = ?", userLogin.UserID).First(&user).Error
			if err != nil { // 获取用户失败
				errormap.SendHttpError(
					c, errormap.ErrorCodeUnauthenticated,
					errormap.NewOrigin("Get user with id=%v failed: %v", userLogin.UserID, err),
				)
				return
			}

			if user.Password != userLogin.EncryptedPassword {
				errormap.SendHttpError(
					c, errormap.ErrorCodePasswordChangedError,
					errormap.NewOrigin("User password has been changed."),
				)
				return
			}

			cache.Set(sessionID, user.UUID)
		} else {
			err := orm.DB.Model(&orm.User{}).Where("uuid = ?", userUUID).First(&user).Error
			if err != nil {
				errormap.SendHttpError(
					c, errormap.ErrorCodeUnauthenticated,
					errormap.NewOrigin("Get user with uuid=%v failed: %v", userUUID, err),
				)
				return
			}
			cache.Set(sessionID, user.UUID)
		}

		c.Set("current_user", user)
		c.Next()
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		account := c.PostForm("account")
		password := c.PostForm("password")

		var user orm.User
		if err := orm.DB.Model(&orm.User{}).Where("account = ?", account).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				errormap.SendHttpError(c, errormap.ErrorCodeAccountNotExistError, err)
				return
			}

			errormap.SendHttpError(c, errormap.ErrorCodeInternalError, err)
			return
		}

		// 校验密码
		encrypted := util.Encrypt(password)
		if encrypted != user.Password {
			errormap.SendHttpError(c, errormap.ErrorCodeAccountPasswordIncorrect, nil)
			return
		}

		remember, err := strconv.ParseBool(c.PostForm("remember"))
		if err != nil {
			remember = false
		}
		uid, _ := uuid.NewRandom()
		sessionID := fmt.Sprintf("%s-%s", uid, account)

		ua := c.Request.Header.Get("User-Agent")
		ip := c.Request.Header.Get("X-Real-IP")

		userLogin := orm.UserLogin{
			UserID:            user.ID,
			AccessToken:       sessionID,
			EncryptedPassword: encrypted,
			IP:                ip,
			UserAgent:         ua,
			KeepLogin:         remember,
		}

		err = orm.DB.Model(&orm.UserLogin{}).Create(&userLogin).Error
		if err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeRecordUserLoginError, err)
			return
		}

		// 记录用户登录状态
		cache.Set(sessionID, user.UUID)

		maxAge := 0
		if remember {
			maxAge = 7 * 24 * 60 * 60
		}

		c.SetCookie("access_token", sessionID, maxAge, "/", "", false, true)
		c.JSON(http.StatusOK, object{"status": "ok"})
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			errormap.SendHttpError(
				c, errormap.ErrorCodeUnauthenticated,
				errormap.NewOrigin("Get sessionID from cookie failed: %v", err),
			)
			return
		}

		if err := orm.DB.Where("access_token = ?", accessToken).Delete(&orm.UserLogin{}).Error; err != nil {
			errormap.SendHttpError(c, errormap.ErrorCodeLogoutFailedError, err)
			return
		}

		// 清除cache中记录的登录状态
		cache.FlushCacheWithKey(accessToken)

		c.JSON(http.StatusOK, object{"status": "ok"})
	}
}
