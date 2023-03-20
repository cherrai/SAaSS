package gin_service

import (
	"fmt"
	"net/http"
	"strconv"

	conf "github.com/cherrai/SAaSS/config"
	"github.com/cherrai/SAaSS/routers"
	"github.com/cherrai/SAaSS/services/middleware"

	"github.com/gin-gonic/gin"
)

var log = conf.Log

var Router *gin.Engine

func Init() {
	gin.SetMode(conf.Config.Server.Mode)

	Router = gin.New()
	InitRouter()
	run()
}

func InitRouter() {
	// 处理跨域
	Router.Use(middleware.Cors("*"))
	Router.NoMethod(func(ctx *gin.Context) {
		ctx.String(200, "Meow Whisper!\nNot method.")
	})
	Router.Use(middleware.CheckRouteMiddleware())
	Router.Use(middleware.RoleMiddleware())
	// // 处理返回值
	Router.Use(middleware.Response())
	// // 请求时间中间件
	Router.Use(middleware.RequestTime())
	// // 错误中间件
	Router.Use(middleware.Error())
	// // 处理解密加密
	// Router.Use(middleware.Encryption())
	Router.Use(middleware.CheckApp())
	Router.Use(middleware.CheckUserToken())
	Router.Use(middleware.Authorize())

	// 测试上传
	// Router.POST("/testupload", func(c *gin.Context) {
	// 	// 		form, _ := c.MultipartForm()
	// 	// 		files := form.File["upload[]"]

	// 	// 		for _, file := range files {
	// 	// 			log.Println(file.Filename)

	// 	// 			// Upload the file to specific dst.
	// 	// 			// c.SaveUploadedFile(file, dst)
	// 	// 		}
	// 	file, err := c.FormFile("files")
	// 	log.Info(c.GetQuery("token"))
	// 	log.Info(c.GetHeader("Authorization"))
	// 	if err != nil {
	// 		c.String(500, "上传图片出错")
	// 	}
	// 	// c.JSON(200, gin.H{"message": file.Header.Context})
	// 	c.SaveUploadedFile(file, "./static/"+file.Filename)
	// 	c.String(http.StatusOK, file.Filename)
	// })
	// // 测试下载
	// Router.GET("/teststatic", func(c *gin.Context) {
	// 	// 测试调整分辨率
	// 	c.File("./static/1589036065311.jpeg")
	// })
	Router.StaticFS("/static", http.Dir("./static"))

	// midArr := [...]gin.HandlerFunc{GinMiddleware("*"), middleware.Authorize()}
	// fmt.Println(midArr)
	// for _, midFunc := range midArr {
	// 	//fmt.Println(index, "\t",value)
	// 	Router.Use(midFunc)
	// }
	Router.StaticFS("/public", http.Dir("./public"))
	routers.InitRouter(Router)

}

func run() {
	if err := Router.Run(":" + strconv.Itoa(conf.Config.Server.Port)); err != nil {
		log.Error("failed run app: ", err)

		// time.AfterFunc(500*time.Millisecond, func() {
		// 	run(router)
		// })
	} else {
		fmt.Println("Gin Http server created successfully. Listening at :" + strconv.Itoa(conf.Config.Server.Port))
	}
}
