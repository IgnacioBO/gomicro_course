package bootstrap

import (
	"fmt"
	"log"
	"os"

	"github.com/IgnacioBO/gomicro_domain/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//Aca tendremos funcionlaidades realciondads con el logger, para sacar funciones del main

// Funcion que inicia un logger
func InitLogger() *log.Logger {
	return log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
}

func DBConnection() (*gorm.DB, error) {
	//DSN (Data Source Name) es una cadena de conexion de BBDD (tipo, servidor, nombre bbdd, user, pass)
	//Estos valores no los tendremos hard codeados si no con variables de entorno usando godotenv
	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("DB_USER"), //user
		os.Getenv("DB_PASS"), //pass
		os.Getenv("DB_HOST"), //sv
		os.Getenv("DB_PORT"), //port
		os.Getenv("DB_NAME")) //bbdd
	fmt.Println(dsn)

	//Usaremos gorm.Open(mysql.Open(stringConexion), configuracionGorn{vacia}) usando las libreria gorm y mysql
	//Nos devuelve la base de datos y un error
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{}) //dentro de gorm.Config pueden configurarse x ejemplo sin logs Logger: logger.Default.LogMode(logger.Silent)
	if err != nil {
		return nil, err
	} //Con solo esto ya nos conectamos

	//Setereamos la base de datos en modo debug para ver lina por linea pero solo si esta en true la vriabel de entorn DB_DEBUG
	if os.Getenv("DB_DEBUG") == "true" {
		db = db.Debug()
	}

	//Ahora especificaremos que queremos CREAR la TABLA usando GORN (en base al struct user/domain.go del otro proyecto)
	//Usando automigrate y un struct (en este caso un puntero del struct) me creara la tabla automaticamente
	if os.Getenv("DB_MIGRATE") == "true" {
		err = db.AutoMigrate(&domain.Course{})
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
