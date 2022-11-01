package main

// go-mysql: go get -u github.com/go-sql-driver/mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

func main() {
	// db
	err := initDB()
	if err != nil {
		fmt.Printf("init DB error:%v \n ", err)
	} else {
		fmt.Println("连接成功db")
	}

	// server
	engine := gin.Default()

	// static page
	engine.LoadHTMLGlob("templates/*")

	// 404
	engine.NoRoute(func(context *gin.Context) {
		context.HTML(http.StatusNotFound, "404.html", nil)
	})

	// 设置500提示中间件
	engine.Use(Recover)

	// json test
	engine.POST("/json", func(context *gin.Context) {
		// 获取数据
		data, _ := context.GetRawData()
		// 用来接收数据
		var m map[string]interface{}
		// json格式化，包装为json数据
		_ = json.Unmarshal(data, &m)
		context.JSON(http.StatusOK, m)
	})

	// add user
	engine.POST("/addUser", func(context *gin.Context) {
		data, _ := context.GetRawData()
		var m map[string]string
		_ = json.Unmarshal(data, &m)
		u := new(UserModel)
		u.Email = m["email"]
		u.Password = m["password"]
		save := Save(u)
		fmt.Println(save)
		context.JSON(http.StatusOK, save)
	})

	// delete by id
	engine.GET("deleteUser/:id", func(context *gin.Context) {
		id, _ := strconv.Atoi(context.Param("id"))
		remove := Remove(id)
		context.JSON(http.StatusOK, remove)
	})

	// update user
	engine.POST("/updateUser", func(context *gin.Context) {
		data, _ := context.GetRawData()
		var m map[string]string
		_ = json.Unmarshal(data, &m)

		fmt.Println(m)

		u2 := new(UserModel)
		u2.Id, _ = strconv.Atoi(m["id"])
		u2.Email = m["email"]
		u2.Password = m["password"]
		update := Update(u2)
		context.JSON(http.StatusOK, update)
	})

	// query all user
	engine.GET("/queryAllUser", func(context *gin.Context) {
		users := QueryAllUser()
		context.JSON(http.StatusOK, gin.H{"data": users})
	})

	// query user by id
	engine.GET("queryUser/:id", func(context *gin.Context) {
		id := context.Param("id")
		// string转成int
		atoi, _ := strconv.Atoi(id)
		user := QueryUserById(atoi)
		context.JSON(http.StatusOK, gin.H{
			"data": user,
		})
	})

	// port
	engine.Run(":8080")
}

// Recover 统一500错误处理函数
func Recover(context *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			// 打印错误堆栈信息
			log.Printf("panic: %v\n", r)
			debug.PrintStack()
			// 封装通用html返回
			context.HTML(200, "500.html", nil)
		}
	}()
	// 加载完 defer recover，继续后续接口调用
	context.Next()
}

type UserModel struct {
	// 首字母大写代表是公共的
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Save(user *UserModel) bool { // 增加
	// 开启事务
	tx, err := DB.Begin()
	if err != nil {
		log.Panicln("user insert error ", err.Error())
		return false
	}
	// 初始化语句并赋值
	result, e := tx.Exec("insert into ginhello.user (email, password) values (?,?);", user.Email, user.Password)
	if e != nil {
		log.Panicln("user insert error ", e.Error())
		return false
	}
	// 提交事务
	tx.Commit()
	// 打印对象id
	fmt.Println(result.LastInsertId())
	return true
}

func Remove(id int) bool { // 删除
	// 开启事务
	begin, err := DB.Begin()
	if err != nil {
		log.Panicln("user remove error ", err.Error())
		return false
	}
	// 初始化语句并赋值
	exec, err := begin.Exec("delete from user where id = ?;", id)
	if err != nil {
		log.Panicln("user remove error ", err.Error())
		return false
	}
	// 提交事务
	begin.Commit()
	// 打印对象id
	fmt.Println(exec.LastInsertId())
	return true
}

func Update(user *UserModel) bool { // 修改
	begin, err := DB.Begin()
	if err != nil {
		log.Panicln("user update error ", err.Error())
		return false
	}
	exec, err := begin.Exec("update user set email = ? where id = ?", user.Email, user.Id)
	if err != nil {
		log.Panicln("user update error ", err.Error())
		return false
	}
	begin.Commit()
	fmt.Println(exec.LastInsertId())
	return true
}

func QueryUserById(id int) UserModel { // 根据id查询用户
	var user UserModel
	err := DB.QueryRow("select * from user where id = ?", id).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		log.Panicln("user query error ", err.Error())
	}
	return user
}

func QueryAllUser() []UserModel { // 多条查询
	var users []UserModel
	rows, err := DB.Query("select * from user")
	if err != nil {
		log.Panicln("user query error ", err.Error())
	}
	for rows.Next() {
		var user UserModel
		err := rows.Scan(&user.Id, &user.Email, &user.Password)
		if err != nil {
			log.Panicln("user query error ", err.Error())
		}
		users = append(users, user)
	}
	return users
}

var DB *sql.DB // 定义全局DB变量

func initDB() (err error) { // 连接数据库
	dsn := "root:123456@tcp(127.0.0.1:3306)/ginhello"
	// 这里DB不要使用 := , 全局变量赋值 在main中使用
	// Open()方法不会校验用户名和密码是否正确，只会检验格式是否正确
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试数据库连接
	err = DB.Ping()
	if err != nil {
		return err
	}
	DB.SetConnMaxLifetime(time.Second * 10) // 连接存活最大时间
	DB.SetMaxIdleConns(200)                 // 最大空闲连接数
	DB.SetMaxOpenConns(10)                  // 最大连接数
	return nil
}
