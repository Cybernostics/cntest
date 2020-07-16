package mysql

import (
	"fmt"
	// register the mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/wjase/dbcontainers"
)

//ForMySQL A configurer which configures a MysqlHost and connection string
func Apply(c *dbcontainers.DBContainer) *dbcontainers.DBContainer {

	env := map[string]string{
		"MYSQL_ALLOW_EMPTY_PASSWORD": "true",
		"MYSQL_DATABASE":             c.DatabaseName,
		"MYSQL_USER":                 c.User,
		"MYSQL_PASSWORD":             c.Password,
	}
	c.Environment = env
	c.DockerImage = "mysql"

	c.ConnectionStringFn = func() string {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.User, c.Password, c.IP, c.Port, c.DatabaseName)
	}
	return c.WithDriverType("mysql").WithPort("3306")
}
