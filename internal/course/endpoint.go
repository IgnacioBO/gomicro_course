package course

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_meta/meta"
)

type (
	//Controller sera una funcion que reciba REspone y Request
	Controller func(ctx context.Context, request interface{}) (interface{}, error)
	Endpoints  struct {
		Create        Controller //Esto es lo mismo que decir Create func(w http.ResponseWriter, r *http.Request), pero como TODOS SON tipo Controller (Definido arriba) nos ahorramos ahcerlo
		Get           Controller
		GetAll        Controller
		Update        Controller
		Delete        Controller
		DeleteClassic Controller
	}
	//Definiremos una struct para definir el request del Craete, con los campos que quiero recibir y los tags de json
	CreateRequest struct {
		Name      string `json:"name"`
		StartDate string `json:"start_date"` //Sera de tipo STRING para convertir de string a date en capa servicio
		EndDate   string `json:"end_date"`
	}
	//Definiremos una struct para definir el request del UPDATE, con los campos que quiero y SE PODRAN ACTUALIZAR y los tags de json
	//Seran de tipo puntero * para que puedan venir vacios y poder separar entre vacios "" y que no vengan
	UpdateRequest struct {
		ID        string
		Name      *string `json:"name"`
		StartDate *string `json:"start_date"`
		EndDate   *string `json:"end_date"`
	}

	GetRequest struct {
		ID string `json:"id"`
	}

	//Este struct tendra los PARAMETROS de la URL para pasarselo
	GetAllRequest struct {
		Name  string
		Limit int
		Page  int
	}

	DeleteRequest struct {
		ID string `json:"id"`
	}

	Response struct {
		Status int         `json:"status"`
		Data   interface{} `json:"data,omitempty"` //omitempty, asi cuando queremos enviamos la data cuando eta ok y cuando este eror se envie el campo error
		Err    string      `json:"error,omitempty"`
		Meta   *meta.Meta  `json:"meta,omitempty"`
	}

	//Struct para guardar la cant page por defecto y otras conf
	Config struct {
		LimitPageDefault string
	}
)

// Funcion que se encargará de hacer los endopints
// Para eso necesitaremos una struct que se llamara endpoints
// Esta funcion va a DEVOLVER una struct de Endpoints, estos endpoints son los que vamos a poder utuaizlar en unestro dominio (course)
func MakeEndpoints(s Service, c Config) Endpoints {
	return Endpoints{
		Create: makeCreateEndpoint(s),
		Get:    makeGetEndpoint(s),
		Update: makeUpdateEndpoint(s),
		Delete: makeDeleteEndpoint(s),
		GetAll: makeGetAllEndpoint(s, c),
	}
}

func makeCreateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("create course")

		var reqStruct = request.(CreateRequest)

		//Validaciones
		if reqStruct.Name == "" {
			return nil, response.BadRequest(ErrNameRequired.Error())
		}
		if reqStruct.StartDate == "" {
			return nil, response.BadRequest(ErrStartDateRequired.Error())
		}
		if reqStruct.EndDate == "" {
			return nil, response.BadRequest(ErrEndDateRequired.Error())
		}

		fmt.Println(reqStruct)
		reqStrucEnJson, _ := json.MarshalIndent(reqStruct, "", " ")
		fmt.Println(string(reqStrucEnJson))

		//Usaremos la s recibida como parametro (de la capa Service y usaremos el metodo CREATE con lo que debe recibir)
		courseNuevo, err := s.Create(ctx, reqStruct.Name, reqStruct.StartDate, reqStruct.EndDate)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		return response.Created("success", courseNuevo, nil), nil
	}
}

func makeUpdateEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("update course")

		//Variable con struct de request (datos de atualizacion)
		reqStruct := request.(UpdateRequest)
		//r.Body tiene el body del request (se espera JSON) y lo decodifica al struct (reqStruct) (osea pasar el json enviado en el request a un struct)

		if reqStruct.Name != nil && *reqStruct.Name == "" {
			return nil, response.BadRequest(ErrNameNotEmpty.Error())
		}
		if reqStruct.StartDate != nil && *reqStruct.StartDate == "" {
			return nil, response.BadRequest(ErrStartDateNotEmpty.Error())

		}
		if reqStruct.EndDate != nil && *reqStruct.EndDate == "" {
			return nil, response.BadRequest(ErrEndDateNotEmpty.Error())

		}

		err := s.Update(ctx, reqStruct.ID, reqStruct.Name, reqStruct.StartDate, reqStruct.EndDate)
		if err != nil {
			if errors.As(err, &ErrCourseNotFound{}) {
				return nil, response.NotFound(err.Error())
			}
			if errors.As(err, &ErrDateBadFormat{}) {
				return nil, response.BadRequest(err.Error())
			}
			return nil, response.InternalServerError(err.Error())
		}
		return response.OK("success", map[string]string{"id": reqStruct.ID}, nil), nil
	}
}

func makeGetEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("get course")

		getReq := request.(GetRequest)

		course, err := s.Get(ctx, getReq.ID)
		if err != nil {
			if errors.As(err, &ErrCourseNotFound{}) {
				return nil, response.NotFound(err.Error())
			} else {
				return nil, response.InternalServerError(err.Error())

			}
		}

		return response.OK("success", course, nil), nil

	}
}

func makeGetAllEndpoint(s Service, c Config) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("getall course")

		getAllReq := request.(GetAllRequest)

		//Luego con podemos acceder a los parametos y guardarlos en el struct Filtro (creado en service.go)
		filtros := Filtros{
			Name: getAllReq.Name,
		}
		//Ahora obtendremos el limit y la pagina desde los parametros
		limit := getAllReq.Limit
		page := getAllReq.Page

		//Ahora llamaremos al Count del service que creamos (antes de hacer la consulta completa)
		cantidad, err := s.Count(ctx, filtros)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}
		//Luego crearemos un meta y le agregaremos la cantidad que consultamos, luego el meta lo ageregaremos a la respuesta
		meta, err := meta.New(page, limit, cantidad, c.LimitPageDefault)

		allCourses, err := s.GetAll(ctx, filtros, meta.Offset(), meta.Limit()) //GetAll recibe el offset (desde q resultado mostrar) y el limit (cuantos desde el offset)
		if err != nil {
			return nil, response.InternalServerError(err.Error())
		}

		return response.OK("success", allCourses, meta), nil
	}
}

// Este devolver un Controller, retora una función de tipo Controller (que definimos arriba) con esta caractesitica
// Es privado porque se llamar solo de este dominio
func makeDeleteEndpoint(s Service) Controller {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		fmt.Println("delete course")

		delReq := request.(DeleteRequest)

		err := s.Delete(ctx, delReq.ID)
		if err != nil {
			if errors.As(err, &ErrCourseNotFound{}) {
				return nil, response.NotFound(err.Error())
			} else {
				return nil, response.InternalServerError(err.Error())

			}
		}

		return response.OK("success", delReq, nil), nil
	}
}
