package main

import (
	// "html/template"
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var AccessKeyID string
var SecretAccessKey string
var MyRegion string
var MyBucket string
var filepath string
var EndPoint string

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}

func ConnectAws() *session.Session {
	AccessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	SecretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	MyRegion = GetEnvWithKey("AWS_REGION")
	EndPoint = GetEnvWithKey("AWS_ENDPOINT")

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
			Endpoint: aws.String(EndPoint),
		})

	if err != nil {
		panic(err)
	}

	return sess
}

func SetupRouter(sess *session.Session) {
	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	// router.Get("/upload", Form)
	router.POST("/upload", UploadImage)
	// router.GET("/image", controllers.DisplayImage)

	_ = router.Run(":4000")
}

func UploadImage(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)

	MyBucket = GetEnvWithKey("BUCKET_NAME")

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		panic(err)
	}
	filename := header.Filename

	//upload to the s3 bucket
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(MyBucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(filename),
		Body:   file,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to upload file",
			"uploader": up,
		})
		return
	}
	filepath = "https://" + MyBucket + "." + "s3-" + MyRegion + ".amazonaws.com/" + filename
	c.JSON(http.StatusOK, gin.H{
		"filepath": filepath,
	})
}

func TestDownlod(c *gin.Context) {
	sess := c.MustGet("sess").(*session.Session)
	// s3Client := s3.New(sess)

	downloader := s3manager.NewDownloader(sess)
	filename := "1638845074280.jpeg"

	requestInput := s3.GetObjectInput{
		Bucket: aws.String(GetEnvWithKey("BUCKET_NAME")),
		Key:    aws.String(filename),
	}

	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(buf, &requestInput)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"hasilbyte": "sEnc",
		})
	}

	fmt.Printf("Downloaded %v bytes", len(buf.Bytes()))
	a := buf.Bytes()
	// mimeType := http.DetectContentType(a)
	sEnc := b64.StdEncoding.EncodeToString(a)
	// fmt.Println(sEnc)

	c.JSON(http.StatusOK, gin.H{
		"hasilbyte": sEnc,
	})
}

func main() {
	LoadEnv()

	sess := ConnectAws()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	router.POST("/upload", UploadImage)
	router.GET("/tes", TestDownlod)

	router.LoadHTMLGlob("templates/*")
	router.GET("/image", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	_ = router.Run(":4000")
}
