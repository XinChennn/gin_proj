package initDB

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

func main() {

	err := initDB()
	if err != nil {
		fmt.Printf("init DB error:%v \n ", err)
	} else {
		fmt.Println("连接成功db")
	}

	u := new(UserModel)
	u.Id = 3

	//  ----------------------->  新增
	// save := u.Save()
	// fmt.Println(save)

	//  ----------------------->  删除
	// remove := u.Remove()
	// fmt.Println(remove)

	//  ----------------------->  修改
	// update := u.Update("aaa@a.com", 4)
	// fmt.Println(update)

	//  ----------------------->  查询（根据id查询用户）
	// userById := QueryUserById(4)
	// fmt.Println(userById)

	//  ----------------------->  查询（多条查询）
	users := QueryAllUser()
	for i := range users {
		fmt.Println(users[i])
	}

	// defer一般用于资源的释放和异常的捕捉, 是Go语言的特性之一
	// defer 语句会将其后面跟随的语句进行延迟处理. 意思就是说 跟在defer后面的语言 将会在程序进行最后的return之后再执行.
	// 在 defer 归属的函数即将返回时，将延迟处理的语句按 defer 的逆序进行执行，也就是说，先被 defer 的语句最后被执行，最后被 defer 的语句，最先被执行。
	defer DB.Close()

}

type UserModel struct {
	Id       int
	Email    string
	Password string
}

func (user *UserModel) Save() bool {  // 增加
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

func (user *UserModel) Remove() bool {  // 删除
	// 开启事务
	begin, err := DB.Begin()
	if err != nil {
		log.Panicln("user remove error ", err.Error())
		return false
	}
	// 初始化语句并赋值
	exec, err := begin.Exec("delete from user where id = ?;", user.Id)
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

func (user *UserModel) Update(email string, id int) bool {  // 修改
	begin, err := DB.Begin()
	if err != nil {
		log.Panicln("user update error ", err.Error())
		return false
	}
	exec, err := begin.Exec("update user set email = ? where id = ?", email, id)
	if err != nil {
		log.Panicln("user update error ", err.Error())
		return false
	}
	begin.Commit()
	fmt.Println(exec.LastInsertId())
	return true
}

func QueryUserById(id int) UserModel {  // 根据id查询用户
	var user UserModel
	err := DB.QueryRow("select * from user where id = ?", id).Scan(&user.Id, &user.Email, &user.Password)
	if err != nil {
		log.Panicln("user query error ", err.Error())
	}
	return user
}

func QueryAllUser() []UserModel {  // 多条查询
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
	DB.SetConnMaxLifetime(time.Second * 10) //连接存活最大时间
	DB.SetMaxIdleConns(200)                 //最大空闲连接数
	DB.SetMaxOpenConns(10)                  // 最大连接数
	return nil
}
