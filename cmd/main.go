package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IgnacioBO/gomicro_course/internal/course"

	"github.com/IgnacioBO/gomicro_course/pkg/bootstrap"
	"github.com/IgnacioBO/gomicro_course/pkg/handler" //Manejar ruteo facilmente (paths y metodos)
	"github.com/joho/godotenv"
	//Driver mysql para gorm
	//Para manejar bbdd facilmente con strcut y funciones (en vez de querys directa)
)

func main() {
	//Definiremos un logger
	//Creamos un script bootsrap.go que INICIALIZA el log
	l := bootstrap.InitLogger()

	//Con godotenv cargamos las variables de entorn denrto de .env para usarlas en el DSN
	//Al usar godotenv.Load() cargar autoamticametne los valores en archivo .env de root
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	//Creamos un config que tenga la cant max de pagina por defecto
	pageLimDef := os.Getenv("PAGINATOR_LIMIT_DEFAULT")
	if pageLimDef == "" {
		l.Fatal("paginator limit default is required")
	}
	courseConfig := course.Config{LimitPageDefault: pageLimDef}

	//DSN (Data Source Name) es una cadena de conexion de BBDD (tipo, servidor, nombre bbdd, user, pass)
	//Creamos un script bootsrap.go que INICIALIZA la conexion usando gorm y varaibles de entoero
	db, err := bootstrap.DBConnection()
	if err != nil {
		log.Fatal(err)
	}

	//Antes de repo, servicio, endpont, generamos un contexto
	ctx := context.Background()
	//Generaremos un objeto repo (que recibe la bbdd y logger) que luego le pasaremos a la capa servicio
	courseRepo := course.NewRepo(l, db)
	//Crearemos un objeto de tipo servicio pasandole un objeto Repository (y logger) para luego pasarselo a la capa enpdoint
	courseService := course.NewService(l, courseRepo)
	//Crearemo un objeto de tipo endpoint y le pasamos el objeto creado (Service). Ademas le pasamos un user.Config
	courseEndpoint := course.MakeEndpoints(courseService, courseConfig)
	h := handler.NewUserHTTPServer(ctx, courseEndpoint)

	port := os.Getenv("PORT")
	address := fmt.Sprintf("127.0.0.1:%s", port)

	srv := &http.Server{
		Handler:      accessControl(h), //Aquie le ponemos el acces control para deifnior op permitdas + el handler que definimos
		Addr:         address,
		ReadTimeout:  5 * time.Second, //Con estos SETEAMOS TIMEOUT DE ESCRITURA Y DE LECTURA (cuanto timepo maximo la api permite)
		WriteTimeout: 5 * time.Second, // Read es REQUEST, WRITE es RESPONE
	}
	//Ahora definim un channel, depsues de generar el server
	//Crearmos un goroutine, la ventaja es que podemos hacer otras cosas mientras se jecuta el serviodor ListenAndServe (como capturar en otro channel señales del sistema como conrtl+c u otras)
	errCh := make(chan error)

	//Aqui generamos una fncion anonimo de tipo GOROUTINE (por eso es **go func()**). Y la ejecutams altiro (por eso temrina en "()" despies de la llave "}" )
	//Ejecutamos el listenandserve() y si hay error retornamos eror al CANAL
	go func() {
		l.Println("listen in", address)
		errCh <- srv.ListenAndServe() //Aqui ejecutamos el ListenAndServe y VAMOS A DEVOLVER UN ERROR al CANAL
	}()

	//Aqui canal espera hasta recibir algo
	err = <-errCh //Recibimos  el error el channel
	if err != nil {
		log.Fatal(err)
	}
}

// Aqui definimo operaciones que PERMITIREMOS y ademas recibimso un Handler original (que sera el que creamo en el main)
func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Aqui definimos operacione spermitidas, origin con * para que puedan venir DEDE CUALQUIER CLIENTE O LADO
		w.Header().Set("Access-Control-Allow-Origin", "*")                                //origin con * para que puedan venir DEDE CUALQUIER CLIENTE O LADO
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS, HEAD") //Metodos permitidos
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept,Authorization,Cache-Control,Content-Type,DNT,If-Modified-Since,Keep-Alive,Origin,User-Agent,X-Requested-With") //Header permitidos

		if r.Method == "OPTIONS" {
			return
		}
		h.ServeHTTP(w, r)

	})

}
