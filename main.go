package main

import (
	"demo/common"
	"demo/component"
	"demo/component/uploadprovider"
	"demo/memcache"
	"demo/middleware"
	"demo/modules/restaurant/restauranttransport/ginrestaurant"
	grpcrestaurantlike "demo/modules/restaurantlike/storage/grpc"
	"demo/modules/restaurantlike/transport/ginresturantlike"
	"demo/modules/upload/uploadtransport/ginupload"
	"demo/modules/user/userstorage"
	"demo/modules/user/usertransport/ginuser"
	"demo/proto"
	"demo/pubsub/pblocal"
	"demo/skio"
	"demo/subscriber"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	jg "go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	//dsn := os.Getenv("DBConnectionStr")
	dsn := "root:@tcp(127.0.0.1:3306)/food?charset=utf8mb4&parseTime=True&loc=Local"
	s3BucketName := os.Getenv("S3BucketName")
	s3Region := os.Getenv("S3Region")
	s3APIKey := os.Getenv("S3APIKey")
	s3SecretKey := os.Getenv("S3SecretKey")
	s3Domain := os.Getenv("S3Domain")
	secretKey := os.Getenv("SYSTEM_SECRET")

	s3Provider := uploadprovider.NewS3Provider(s3BucketName, s3Region, s3APIKey, s3SecretKey, s3Domain)
	//connect grpc
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051",opts)
	if err != nil{
		log.Fatalln(err)
	}
	defer cc.Close()


	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db = db.Debug()

	if err := runService(db, s3Provider, secretKey,cc); err != nil {
		log.Fatalln(err)
	}
}

func runService(db *gorm.DB, upProvider uploadprovider.UploadProvider, secretKey string, cc *grpc.ClientConn) error {
	appCtx := component.NewAppContext(db, upProvider, secretKey, pblocal.NewPubSub())

	userStore := userstorage.NewSQLStore(appCtx.GetMainDBConnection())
	userCachingStore := memcache.NewUserCaching(memcache.NewCaching(), userStore)

	r := gin.Default()

	//subscriber.Setup(appCtx)
	rtEngine := skio.NewEngine()

	if err := rtEngine.Run(appCtx, r); err != nil {
		log.Fatalln(err)
	}

	if err := subscriber.NewEngine(appCtx, rtEngine).Start(); err != nil {
		log.Fatalln(err)
	}

	r.Use(middleware.Recover(appCtx))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// CRUD
	r.StaticFile("/demo/", "./demo.html")

	v1 := r.Group("/v1")

	v1.POST("/upload", ginupload.Upload(appCtx))

	v1.POST("/register", ginuser.Register(appCtx))
	v1.POST("/login", ginuser.Login(appCtx))
	v1.GET("/profile", middleware.RequiredAuth(appCtx, userCachingStore), ginuser.GetProfile(appCtx))

	restaurants := v1.Group("/restaurants", middleware.RequiredAuth(appCtx, userCachingStore))
	{
		restaurants.POST("", ginrestaurant.CreateRestaurant(appCtx))
		restaurants.GET("/:id", ginrestaurant.GetRestaurant(appCtx))
		restaurants.GET("", ginrestaurant.ListRestaurant(appCtx,cc))
		restaurants.PATCH("/:id", ginrestaurant.UpdateRestaurant(appCtx))
		restaurants.DELETE("/:id", ginrestaurant.DeleteRestaurant(appCtx))

		restaurants.GET("/:id/liked-users", ginresturantlike.ListUser(appCtx))
		restaurants.POST("/:id/like", ginresturantlike.UserLikeRestaurant(appCtx))
		restaurants.DELETE("/:id/unlike", ginresturantlike.UserUnlikeRestaurant(appCtx))
	}

	v1.GET("/encode-uid", func(c *gin.Context) {
		type reqData struct {
			DbType int `form:"type"`
			RealId int `form:"id"`
		}

		var d reqData
		c.ShouldBind(&d)

		c.JSON(http.StatusOK, gin.H{
			"id": common.NewUID(uint32(d.RealId), d.DbType, 1),
		})
	})

	je, err := jg.NewExporter(jg.Options{
		AgentEndpoint: "localhost:6831",
		Process:       jg.Process{ServiceName: "G04-Food-Delivery"},
	})

	if err != nil {
		log.Println(err)
	}

	trace.RegisterExporter(je)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.ProbabilitySampler(1)})

	// create grpc
	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()
	// Attach the Greeter service to the server
	proto.RegisterRestaurantLikeServiceServer(s, grpcrestaurantlike.NewGRPCServer(db))
	// Serve gRPC Server
	go func() {
		log.Println("Serving gRPC on 0.0.0.0:50051")
		if err := s.Serve(lis) ; err != nil{
			log.Fatalln(err)
		}
	}()

	return http.ListenAndServe(
		":8888",
		&ochttp.Handler{
			Handler: r,
		},
	)

	//return r.Run()
}

//func startSocketIOServer(engine *gin.Engine, appCtx component.AppContext) {
//	server, _ := socketio.NewServer(&engineio.Options{
//		Transports: []transport.Transport{websocket.Default},
//	})
//	// moi client ket noi vao
//	server.OnConnect("/", func(s socketio.Conn) error {
//		//s.SetContext("")
//		fmt.Println("connected:", s.ID(), " IP:", s.RemoteAddr())
//
//		//s.Join("Shipper")
//		//server.BroadcastToRoom("/", "Shipper", "test", "Hello 200lab")
//
//		return nil
//	})
//
//	server.OnError("/", func(s socketio.Conn, e error) {
//		fmt.Println("meet error:", e)
//	})
//
//	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
//		fmt.Println("closed", reason)
//		// Remove socket from socket engine (from app context)
//	})
//
//	server.OnEvent("/", "authenticate", func(s socketio.Conn, token string) {
//
//		// Validate token
//		// If false: s.Close(), and return
//
//		// If true
//		// => UserId
//		// Fetch db find user by Id
//		// Here: s belongs to who? (user_id)
//		// We need a map[user_id][]socketio.Conn
//
//		db := appCtx.GetMainDBConnection()
//		store := userstorage.NewSQLStore(db)
//		//
//		tokenProvider := jwt.NewTokenJWTProvider(appCtx.SecretKey())
//		//
//		payload, err := tokenProvider.Validate(token)
//
//		if err != nil {
//			s.Emit("authentication_failed", err.Error())
//			s.Close()
//			return
//		}
//		//
//		user, err := store.FindUser(context.Background(), map[string]interface{}{"id": payload.UserId})
//		//
//		if err != nil {
//			s.Emit("authentication_failed", err.Error())
//			s.Close()
//			return
//		}
//
//		if user.Status == 0 {
//			s.Emit("authentication_failed", errors.New("you has been banned/deleted"))
//			s.Close()
//			return
//		}
//
//		user.Mask(false)
//
//		s.Emit("your_profile", user)
//	})
//
//	server.OnEvent("/", "test", func(s socketio.Conn, msg string) {
//		log.Println(msg)
//	})
//
//	type Person struct {
//		Name string `json:"name"`
//		Age  int    `json:"age"`
//	}
//
//	server.OnEvent("/", "notice", func(s socketio.Conn, p Person) {
//		fmt.Println("server receive notice:", p.Name, p.Age)
//
//		p.Age = 33
//		s.Emit("notice", p)
//
//	})
//
//	// server.OnEvent("/", "test", func(s socketio.Conn, msg string) {
//	// 	fmt.Println("server receive test:", msg)
//	// })
//	//
//	//server.OnEvent("/chat", "msg", func(s socketio.Conn, msg string) string {
//	//	s.SetContext(msg)
//	//	return "recv " + msg
//	//})
//	//
//	//server.OnEvent("/", "bye", func(s socketio.Conn) string {
//	//	last := s.Context().(string)
//	//	s.Emit("bye", last)
//	//	s.Close()
//	//	return last
//	//})
//	//
//	//server.OnEvent("/", "noteSumit", func(s socketio.Conn) string {
//	//	last := s.Context().(string)
//	//	s.Emit("bye", last)
//	//	s.Close()
//	//	return last
//	//})
//
//	go server.Serve()
//
//	engine.GET("/socket.io/*any", gin.WrapH(server))
//	engine.POST("/socket.io/*any", gin.WrapH(server))
//}
