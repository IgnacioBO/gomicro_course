package handler

//AQUI ESTARAN LOS RUTEOS Y LOS MIDDLWARE

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IgnacioBO/go_lib_response/response"
	"github.com/IgnacioBO/gomicro_course/internal/course"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// Este recibe contexto y un endpoint que definimos en la capa del endpionts
func NewUserHTTPServer(ctx context.Context, endpoints course.Endpoints) http.Handler {
	router := mux.NewRouter()

	//Esta se guarad en opciones y se pone al final en Handle
	opciones := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	//Ahora usaremos Handle, poreque a este se le puede pasar un server (httptranpsort)
	router.Handle("/courses", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Create),
		decodeCreateCourse,
		encodeResponse,
		opciones...,
	)).Methods("POST")

	router.Handle("/courses/{id}", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Get),
		decodeGetCourse,
		encodeResponse,
		opciones...,
	)).Methods("GET")

	router.Handle("/courses", httptransport.NewServer(
		endpoint.Endpoint(endpoints.GetAll),
		decodeGetAllCourse,
		encodeResponse,
		opciones...,
	)).Methods("GET")
	/*
		router.Handle("/users/{id}", httptransport.NewServer(
			endpoint.Endpoint(endpoints.Delete),
			decodeDeleteUser,
			encodeResponse,
			opciones...,
		)).Methods("DELETE")
	*/
	router.Handle("/courses/{id}", httptransport.NewServer(
		endpoint.Endpoint(endpoints.Update),
		decodeUpdateCourse,
		encodeResponse,
		opciones...,
	)).Methods("PATCH")

	return router
}

// *** MIDDLEWARE REQUEST ***
func decodeCreateCourse(_ context.Context, r *http.Request) (interface{}, error) {
	var reqStruct course.CreateRequest

	//Ahora hacemos el decode del body del json al srtuct de REquest de course
	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}
	return reqStruct, nil
}

// *** MIDDLEWARE RESPONSE ***
func encodeResponse(ctx context.Context, w http.ResponseWriter, resp interface{}) error {
	rInterface := resp.(response.Response)                            //Transformamos el resp a response.Respone (al interface) -> YA QUE LE ENAIREMOS SIEMPRE UN objeto RESPONSE (CREADO POR NOSOTROS, q tiene el code, mensage, meta, etc, todo el json)
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	w.WriteHeader(rInterface.StatusCode())
	return json.NewEncoder(w).Encode(rInterface) //resp tendra el user.User del domain y otroas datos si es necesario para ocnveritse en json

}

// *** MIDDLEWARE RESPONSE DE ERROR ***
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8") //Linea miea para que se determine que respondera un json
	respInterface := err.(response.Response)                          //Tranfosrmamos el error recibido a la interfac response.Response que craemos
	//¿Porque funciona esta conversion de tipo error al de nosotros?, porque la interfaz 'error' de go pide que haya un metodo Error() string [QUE CREAMOS EN nuestro respon.RESPONSE!]
	//Entonces como implementamos el metodo Error() string funcinoa, ademas tenemos al ventaja que vamos apoder obtener MAS DATOS porque repsonse.Response tiene mas metodos como (StatusCode())
	//Entonces podemos transofrmar un error a una interfac propia con MAS METODOS Y MAS DATOS UE UN ERROR NORMAL!
	w.WriteHeader(respInterface.StatusCode())
	_ = json.NewEncoder(w).Encode(respInterface)

}

// *** MIDDLEWARE REQUEST GET ***
func decodeGetCourse(_ context.Context, r *http.Request) (interface{}, error) {
	var getReq course.GetRequest
	variablesPath := mux.Vars(r)
	getReq.ID = variablesPath["id"] //OBtenemos el id y lo guardamos en el cmapo ID de getReq

	fmt.Println("id es:", getReq.ID)

	return getReq, nil

}

// *** MIDDLEWARE REQUEST GET All ***
// Funcion de decode, de GET
func decodeGetAllCourse(_ context.Context, r *http.Request) (interface{}, error) {
	//Query() devielve un objeto que permite acceder a los parametros d la url (...?campo=123&campo2=hola)
	variablesURL := r.URL.Query()

	//Ahora obtendremos el limit y la pagina desde los parametros
	limit, _ := strconv.Atoi(variablesURL.Get("limit"))
	page, _ := strconv.Atoi(variablesURL.Get("page"))

	getReqAll := course.GetAllRequest{
		Name:  variablesURL.Get("name"),
		Limit: limit,
		Page:  page,
	}

	return getReqAll, nil
}

/*
// *** MIDDLEWARE REQUEST Delete ***
func decodeDeleteUser(_ context.Context, r *http.Request) (interface{}, error) {
	variablesPath := mux.Vars(r)
	id := variablesPath["id"]
	fmt.Println("id a eliminar es:", id)
	deleteReq := user.DeleteRequest{ID: id}

	return deleteReq, nil

}
*/
// *** MIDDLEWARE REQUEST Delete***
func decodeUpdateCourse(_ context.Context, r *http.Request) (interface{}, error) {
	var reqStruct course.UpdateRequest

	err := json.NewDecoder(r.Body).Decode(&reqStruct)
	if err != nil {
		return nil, response.BadRequest(fmt.Sprintf("invalid request format: '%v'", err.Error()))
	}

	variablesPath := mux.Vars(r)
	reqStruct.ID = variablesPath["id"]

	return reqStruct, nil

}
